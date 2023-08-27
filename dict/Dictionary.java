package net.pgaskin.dictionary;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Arrays;
import java.text.Normalizer;
import java.util.Map;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.function.IntUnaryOperator;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

class Dictionary {
    public static String normalize(String term) {
        // decompose accents and stuff
        // convert similar characters with only stylistic differences
        // other unicode normalization stuff
        term = Normalizer.normalize(term, Normalizer.Form.NFKD);

        // to lowercase (unicode-aware)
        term = term.toLowerCase();

        // normalize whitespace
        term = term.trim().replaceAll("\\s+", " ");

        Map<String, String> repl = new HashMap<String, String>() {{
            // replace smart punctuation
            put("\u00ab", "\"");
            put("\u00bb", "\"");
            put("\u2010", "-");
            put("\u2011", "-");
            put("\u2012", "-");
            put("\u2013", "-");
            put("\u2014", "-");
            put("\u2015", "-");
            put("\u2018", "'");
            put("\u2019", "'");
            put("\u201a", "'");
            put("\u201b", "'");
            put("\u201c", "\"");
            put("\u201d", "\"");
            put("\u201e", "\"");
            put("\u201f", "\"");
            put("\u2024", ".");
            put("\u2032", "'");
            put("\u2033", "\"");
            put("\u2035", "'");
            put("\u2036", "\"");
            put("\u2038", "^");
            put("\u2039", "'");
            put("\u203a", "'");
            put("\u204f", ";");

            // expand ligatures
            put("\ua74f", "oo");
            put("\u00df", "ss");
            put("\u00e6", "ae");
            put("\u0153", "oe");
            put("\ufb00", "ff");
            put("\ufb01", "fi");
            put("\ufb02", "fl");
            put("\ufb03", "ffi");
            put("\ufb04", "ffl");
            put("\ufb05", "ft");
            put("\ufb06", "st");
            put("\u2025", "..");
            put("\u2026", "...");
            put("\u2042", "***");
            put("\u2047", "??");
            put("\u2048", "?!");
            put("\u2049", "!?");
        }};
        StringBuffer b = new StringBuffer();
        Matcher m = Pattern.compile(String.join("|", repl.keySet())).matcher(term);
        while (m.find()) {
            m.appendReplacement(b, repl.get(m.group()));
        }
        m.appendTail(b);
        term = b.toString();

        // normalize dashes
        term = term.replaceAll("-+", "-");

        // remove unknown characters/diacritics
        // note: since we decomposed diacritics, this will leave the base char
        term = term.replaceAll("[^abcdefghijklmnopqrstuvwxyz0123456789 -'_.,]", "");

        return term;
    }

    private final DictionaryFS fs;
    private final DictionaryIndex index;
    private final Map<String, DictionaryShard> cache;

    Dictionary(DictionaryFS fs, ByteBuffer index) {
        this(fs, index, 14);
    }

    Dictionary(DictionaryFS fs, ByteBuffer index, int shardCacheMax) {
        this.fs = fs;
        this.index = new DictionaryIndex(index);
        this.cache = new LinkedHashMap<String, DictionaryShard>(shardCacheMax + 1, .75F, true) {
            public boolean removeEldestEntry(Map.Entry eldest) {
                return this.size() > shardCacheMax;
            }
        };
    }

    private DictionaryShard getShardCached(String shard) {
        synchronized (this) {
            DictionaryShard s = this.cache.get(shard);
            if (s == null) {
                try {
                    s = new DictionaryShard(ByteBuffer.wrap(fs.read(shard)), this.index.shardSize());
                } catch (IOException ex) {
                    throw new RuntimeException(ex);
                }
            }
            this.cache.put(shard, s);
            return s;
        }
    }

    public DictionaryResult query(String term) {
        return this.query(term, false);
    }

    public DictionaryResult query(String term, boolean normalized) {
        if (!normalized) {
            term = normalize(term);
        }
        if (term.length() == 0) {
            return null;
        }

        DictionaryRef[] refs = this.index.lookup(term);
        if (refs.length == 0) {
            term = term.replaceAll("'s", "");
            refs = this.index.lookup(term);
        }
        if (refs.length == 0) {
            term = term.replaceAll("s$", "");
            refs = this.index.lookup(term);
        }
        if (refs.length == 0) {
            term = term.replaceAll("-", "");
            refs = this.index.lookup(term);
        }
        if (refs.length == 0) {
            term = term.replaceAll("-", "");
            refs = this.index.lookup(term);
        }
        if (refs.length == 0) {
            term = term.replaceAll("ly$", "");
            refs = this.index.lookup(term);
        }
        if (refs.length == 0) {
            term = term.replaceAll("ing$", "");
            refs = this.index.lookup(term);
        }
        if (refs.length == 0) {
            return null;
        }

        DictionaryEntry[] entries = new DictionaryEntry[refs.length];
        for (int i = 0; i < refs.length; i++) {
            entries[i] = this.get(refs[i]);
        }
        return new DictionaryResult(term, entries);
    }

    public DictionaryEntry get(DictionaryRef ref) {
        DictionaryShard shard = this.getShardCached(ref.shard());
        return new DictionaryEntry(shard.get(ref.index()));
    }

    public static Dictionary load(DictionaryFS fs) throws IOException {
        return new Dictionary(fs, ByteBuffer.wrap(fs.read("index")));
    }
}

interface DictionaryFS {
    byte[] read(String path) throws IOException;

    class FileSystem implements DictionaryFS {
        final Path base;

        public FileSystem(Path base) {
            this.base = base;
        }

        public byte[] read(String path) throws IOException {
            return Files.readAllBytes(base.resolve(path));
        }
    }
}

class DictionaryResult extends ArrayList<DictionaryEntry> {
    public final String term;

    DictionaryResult(String term, DictionaryEntry[] entries) {
        super(Arrays.asList(entries));
        this.term = term;
        this.sortRelevance();
    }

    public void sortRelevance() {
        this.sort((a, b) -> {
            // exact matches
            if (a.name.equals(this.term) && b.name.equals(this.term)) return -1;
            if (!a.name.equals(this.term) && b.name.equals(this.term)) return 1;

            boolean aVar = false, bVar = false;
            aVar:
            for (DictionaryEntry.MeaningGroup g : a.meaningGroups) {
                for (String v : g.wordVariants) {
                    if (v.equalsIgnoreCase(this.term)) {
                        aVar = true;
                        break aVar;
                    }
                }
            }
            bVar:
            for (DictionaryEntry.MeaningGroup g : b.meaningGroups) {
                for (String v : g.wordVariants) {
                    if (v.equalsIgnoreCase(this.term)) {
                        bVar = true;
                        break bVar;
                    }
                }
            }

            // exact variant matches
            if (aVar && !bVar) return -1;
            if (!aVar && bVar) return 1;

            String aHead = a.name.toLowerCase();
            String bHead = b.name.toLowerCase();

            // case-insensitive headword matches
            if (aHead.equals(this.term) && !bHead.equals(this.term)) return -1;
            if (!aHead.equals(this.term) && bHead.equals(this.term)) return 1;

            // non-abbreviations
            if (aHead.equals(a.name) && !bHead.equals(b.name)) return -1;
            if (!aHead.equals(a.name) && bHead.equals(b.name)) return 1;

            // more meaning groups
            if (a.meaningGroups.length > b.meaningGroups.length) return -1;
            if (a.meaningGroups.length < b.meaningGroups.length) return 1;

            int aN = 0, bN = 0;
            for (DictionaryEntry.MeaningGroup g : a.meaningGroups) {
                aN += g.meanings.length;
            }
            for (DictionaryEntry.MeaningGroup g : b.meaningGroups) {
                bN += g.meanings.length;
            }

            // more meanings
            if (aN > bN) return -1;
            if (aN < bN) return 1;

            // common prefix with headword
            if (aHead.startsWith(this.term) && !bHead.startsWith(this.term)) return -1;
            if (!aHead.startsWith(this.term) && bHead.startsWith(this.term)) return 1;

            return a.name.compareTo(b.name);
        });
    }
}

class DictionaryEntry {
    public final String name;
    public final String pronunciation;
    public final MeaningGroup[] meaningGroups;
    public final String info;
    public final String source;

    DictionaryEntry(ByteBuffer buf) {

        // Name
        final byte[] name = new byte[buf.getInt()];
        buf.get(name);
        this.name = new String(name, StandardCharsets.UTF_8);

        // Pronunciation
        final byte[] pronunciation = new byte[buf.getInt()];
        buf.get(pronunciation);
        this.pronunciation = new String(pronunciation, StandardCharsets.UTF_8);

        // MeaningGroups
        final int meaningGroups = buf.getInt();
        this.meaningGroups = new MeaningGroup[meaningGroups];
        for (int i = 0; i < meaningGroups; i++) {
            this.meaningGroups[i] = new MeaningGroup(buf);
        }

        // Info
        final byte[] info = new byte[buf.getInt()];
        buf.get(info);
        this.info = new String(info, StandardCharsets.UTF_8);

        // Pronunciation
        final byte[] source = new byte[buf.getInt()];
        buf.get(source);
        this.source = new String(source, StandardCharsets.UTF_8);
    }

    static class MeaningGroup {
        public final String[] info;
        public final Meaning[] meanings;
        public final String[] wordVariants;

        MeaningGroup(ByteBuffer buf) {

            // Info
            final int info = buf.getInt();
            this.info = new String[info];
            for (int i = 0; i < info; i++) {
                final byte[] str = new byte[buf.getInt()];
                buf.get(str);
                this.info[i] = new String(str, StandardCharsets.UTF_8);
            }

            // Meanings
            final int meanings = buf.getInt();
            this.meanings = new Meaning[meanings];
            for (int i = 0; i < meanings; i++) {
                this.meanings[i] = new Meaning(buf);
            }

            // WordVariants
            final int wordVariants = buf.getInt();
            this.wordVariants = new String[wordVariants];
            for (int i = 0; i < wordVariants; i++) {
                final byte[] str = new byte[buf.getInt()];
                buf.get(str);
                this.wordVariants[i] = new String(str, StandardCharsets.UTF_8);
            }
        }

        static class Meaning {
            public final String[] tags;
            public final String text;
            public final String[] examples;

            Meaning(ByteBuffer buf) {

                // Tags
                final int tags = buf.getInt();
                this.tags = new String[tags];
                for (int i = 0; i < tags; i++) {
                    final byte[] str = new byte[buf.getInt()];
                    buf.get(str);
                    this.tags[i] = new String(str, StandardCharsets.UTF_8);
                }

                // Text
                final byte[] text = new byte[buf.getInt()];
                buf.get(text);
                this.text = new String(text, StandardCharsets.UTF_8);

                // Examples
                final int examples = buf.getInt();
                this.examples = new String[examples];
                for (int i = 0; i < examples; i++) {
                    final byte[] str = new byte[buf.getInt()];
                    buf.get(str);
                    this.examples[i] = new String(str, StandardCharsets.UTF_8);
                }
            }
        }
    }
}

class DictionaryRef {
    private final int shard;
    private final int index;

    DictionaryRef(int shard, int index) {
        this.shard = shard;
        this.index = index;
    }

    public String shard() {
        return String.format("%03x", this.shard);
    }

    public int index() {
        return this.index;
    }

    public String toString() {
        return this.shard() + "/" + this.index();
    }
}

class DictionaryIndex {
    private final int shardSize;
    private final int[] bucketCounts;
    private final ByteBuffer[] bucketWords;
    private final ByteBuffer[] bucketIndexes;

    DictionaryIndex(ByteBuffer buf) {
        int n = 8;
        this.shardSize = buf.getInt(buf.position());
        this.bucketCounts = new int[buf.getInt(buf.position() + 4)];

        for (int i = 0; i < this.bucketCounts.length; i++) {
            this.bucketCounts[i] = buf.getInt(buf.position() + n);
            n += 4;
        }

        this.bucketWords = new ByteBuffer[this.bucketCounts.length];
        for (int i = 0; i < this.bucketCounts.length; i++) {
            int count = this.bucketCounts[i];
            int len = i + 1;
            this.bucketWords[i] = buf.slice();
            this.bucketWords[i].position(buf.position() + n);
            this.bucketWords[i].limit(this.bucketWords[i].position() + count * len);
            n += count * len;
        }

        this.bucketIndexes = new ByteBuffer[this.bucketCounts.length];
        for (int i = 0; i < this.bucketCounts.length; i++) {
            int count = this.bucketCounts[i];
            this.bucketIndexes[i] = buf.slice();
            this.bucketIndexes[i].position(buf.position() + n);
            this.bucketIndexes[i].limit(this.bucketIndexes[i].position() + count * 4);
            n += count * 4;
        }
    }

    public DictionaryRef[] lookup(String term) {
        ByteBuffer arr = ByteBuffer.wrap(term.getBytes(StandardCharsets.UTF_8));
        int len = arr.limit();
        if (len >= this.bucketCounts.length) {
            return new DictionaryRef[0];
        }

        int bucket = len - 1;
        int count = this.bucketCounts[bucket];
        ByteBuffer words = this.bucketWords[bucket];
        ByteBuffer indexes = this.bucketIndexes[bucket];

        int[] res = binarySearchRange(count, i -> {
            ByteBuffer word = words.slice();
            word.position(word.position() + i * len);
            word.limit(word.position() + len);
            return arr.slice().compareTo(word);
        });

        int lo = res[0], hi = res[1];
        if (lo == -1) {
            return new DictionaryRef[0];
        }

        DictionaryRef[] refs = new DictionaryRef[hi - lo + 1];
        for (int i = lo; i <= hi; i++) {
            int ent = indexes.getInt(indexes.position() + i * 4);
            refs[i - lo] = new DictionaryRef(ent / this.shardSize, ent % this.shardSize);
        }
        return refs;
    }

    public int shardSize() {
        return this.shardSize;
    }

    private static int binarySearch(int n, IntUnaryOperator cmp) {
        int lo = 0;
        int hi = n - 1;
        while (lo <= hi) {
            int mi = (lo + hi) / 2;
            int cv = cmp.applyAsInt(mi);
            if (cv == 0) {
                return mi;
            } else if (cv < 0) {
                hi = mi - 1;
            } else {
                lo = mi + 1;
            }
        }
        return -1;
    }

    private static int[] binarySearchRange(int n, IntUnaryOperator cmp) {
        int lo = binarySearch(n, cmp);
        int hi = lo;
        if (lo != -1) {
            while (lo > 0 && cmp.applyAsInt(lo - 1) == 0) {
                lo--;
            }
            while (hi < n - 1 && cmp.applyAsInt(hi + 1) == 0) {
                hi++;
            }
        }
        return new int[]{lo, hi};
    }
}

class DictionaryShard {
    private final int[] offsets;
    private final ByteBuffer buf;

    DictionaryShard(ByteBuffer buf, int shardSize) {
        this.offsets = new int[shardSize];
        for (int i = 0; i < this.offsets.length; i++) {
            this.offsets[i] = buf.getInt(buf.position() + i * 4);
        }
        this.buf = buf;
    }

    public ByteBuffer get(int index) {
        ByteBuffer tmp = buf.slice();
        tmp.position(buf.position() + this.offsets[index]);
        tmp.limit(buf.position() + this.offsets[index + 1]);
        return tmp;
    }
}
