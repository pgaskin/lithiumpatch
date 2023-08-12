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
import org.json.JSONArray;
import org.json.JSONObject;

public class Dictionary {
    private static final Map<Path, Dictionary> dictionaryCache = new HashMap<>();

    public static Dictionary loadCached(Path base) throws IOException {
        base = base.toAbsolutePath();
        synchronized (dictionaryCache) {
            Dictionary d = dictionaryCache.get(base);
            if (d == null) {
                d = Dictionary.load(base);
                dictionaryCache.put(base, d);
            }
            return d;
        }
    }

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

    private final Path base;
    private final DictionaryIndex index;
    private final Map<String, DictionaryShard> cache;

    Dictionary(Path base, ByteBuffer index) {
        this(base, index, 14);
    }

    Dictionary(Path base, ByteBuffer index, int shardCacheMax) {
        this.base = base;
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
                    s = new DictionaryShard(ByteBuffer.wrap(Files.readAllBytes(base.resolve(shard))), this.index.shardSize());
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

    public static Dictionary load(Path base) throws IOException {
        return new Dictionary(base, ByteBuffer.wrap(Files.readAllBytes(base.resolve("index"))));
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
    final String name;
    final String pronunciation;
    final MeaningGroup[] meaningGroups;
    final String info;
    final String source;

    DictionaryEntry(JSONObject w) {
        Object name = w.opt("w");
        this.name = name instanceof String ? (String) name : "";

        Object pronunciation = w.opt("p");
        this.pronunciation = pronunciation instanceof String ? (String) pronunciation : "";

        Object meaningGroups = w.opt("m");
        if (meaningGroups instanceof JSONArray) {
            this.meaningGroups = new MeaningGroup[((JSONArray) meaningGroups).length()];
            for (int j = 0; j < this.meaningGroups.length; j++) {
                Object v = ((JSONArray) meaningGroups).get(j);
                this.meaningGroups[j] = new MeaningGroup(v instanceof JSONObject ? (JSONObject) v : new JSONObject());
            }
        } else {
            this.meaningGroups = new MeaningGroup[0];
        }

        Object info = w.opt("i");
        this.info = info instanceof String ? (String) info : "";

        Object source = w.opt("s");
        this.source = source instanceof String ? (String) source : "";
    }

    static class MeaningGroup {
        final String[] info;
        final Meaning[] meanings;
        final String[] wordVariants;

        MeaningGroup(JSONObject m) {
            Object info = m.opt("i");
            if (info instanceof JSONArray) {
                this.info = new String[((JSONArray) info).length()];
                for (int j = 0; j < this.info.length; j++) {
                    Object v = ((JSONArray) info).get(j);
                    this.info[j] = v instanceof String ? (String) v : "";
                }
            } else {
                this.info = new String[0];
            }

            Object meanings = m.opt("m");
            if (meanings instanceof JSONArray) {
                this.meanings = new Meaning[((JSONArray) meanings).length()];
                for (int j = 0; j < this.meanings.length; j++) {
                    Object v = ((JSONArray) meanings).get(j);
                    this.meanings[j] = new Meaning(v instanceof JSONObject ? (JSONObject) v : new JSONObject());
                }
            } else {
                this.meanings = new Meaning[0];
            }

            Object wordVariants = m.opt("v");
            if (wordVariants instanceof JSONArray) {
                this.wordVariants = new String[((JSONArray) wordVariants).length()];
                for (int j = 0; j < this.wordVariants.length; j++) {
                    Object vv = ((JSONArray) wordVariants).get(j);
                    this.wordVariants[j] = vv instanceof String ? (String) vv : "";
                }
            } else {
                this.wordVariants = new String[0];
            }
        }

        static class Meaning {
            final String[] tags;
            final String text;
            final String[] examples;

            Meaning(JSONObject m) {
                Object tags = m.opt("t");
                if (tags instanceof JSONArray) {
                    this.tags = new String[((JSONArray) tags).length()];
                    for (int i = 0; i < this.tags.length; i++) {
                        Object v = ((JSONArray) tags).get(i);
                        this.tags[i] = v instanceof String ? (String) v : "";
                    }
                } else {
                    this.tags = new String[0];
                }

                Object text = m.opt("x");
                this.text = text instanceof String ? (String) text : "";

                Object examples = m.opt("s");
                if (examples instanceof JSONArray) {
                    this.examples = new String[((JSONArray) examples).length()];
                    for (int i = 0; i < this.examples.length; i++) {
                        Object v = ((JSONArray) examples).get(i);
                        this.examples[i] = v instanceof String ? (String) v : "";
                    }
                } else {
                    this.examples = new String[0];
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

    public JSONObject get(int index) {
        ByteBuffer tmp = buf.slice();
        tmp.position(buf.position() + this.offsets[index]);
        tmp.limit(buf.position() + this.offsets[index + 1]);
        return new JSONObject(StandardCharsets.UTF_8.decode(tmp).toString());
    }
}
