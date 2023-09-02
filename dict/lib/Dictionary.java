package net.pgaskin.dictionary;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.file.Files;
import java.nio.file.Path;
import java.text.Normalizer;

import static net.pgaskin.dictionary.DictionaryUtil.*;

public class Dictionary {
    private final DictionaryIndex index;
    private final DictionaryShard.Provider shard;

    public interface FS {
        ByteBuffer read(String name);

        static FS local(Path base) {
            return name -> {
                try {
                    return ByteBuffer.wrap(Files.readAllBytes(base.resolve(name)));
                } catch (IOException ex) {
                    throw new RuntimeException(ex);
                }
            };
        }
    }

    public Dictionary(DictionaryIndex index, DictionaryShard.Provider shard) {
        this.index = index;
        this.shard = shard;
    }

    public static Dictionary load(FS fs) {
        return Dictionary.load(fs, 14);
    }

    public static Dictionary load(FS fs, int shardCacheMax) {
        DictionaryIndex index = new DictionaryIndex(fs.read("index"));
        DictionaryShard.Provider shard = DictionaryUtil.<String, DictionaryShard>makeCache(x -> new DictionaryShard(fs.read(x)), shardCacheMax)::apply;
        return new Dictionary(index, shard);
    }

    public DictionaryResult query(String term) {
        return this.query(term, false);
    }

    public DictionaryResult query(String term, boolean normalized) {
        if (!normalized) {
            term = Dictionary.normalize(term);
        }
        if (term.isEmpty()) {
            return new DictionaryResult(term);
        }

        // look up the word, plus some basic fallbacks
        final String origTerm = term;
        int[] entries = this.index.lookup(term);
        if (entries.length == 0 && term.endsWith("'s")) {
            term = term.substring(0, term.length() - "'s".length());
            entries = this.index.lookup(term);
        }
        if (entries.length == 0 && term.endsWith("s")) {
            term = term.substring(0, term.length() - "s".length());
            entries = this.index.lookup(term);
        }
        if (entries.length == 0 && term.indexOf('-') != -1) {
            term = removeChar(term, '-');
            entries = this.index.lookup(term);
        }
        if (entries.length == 0 && term.endsWith("ly")) {
            term = term.substring(0, term.length() - "ly".length());
            entries = this.index.lookup(term);
        }
        if (entries.length == 0 && term.endsWith("ing")) {
            term = term.substring(0, term.length() - "ing".length());
            entries = this.index.lookup(term);
        }
        if (entries.length == 0) {
            return new DictionaryResult(origTerm);
        }

        final DictionaryEntry[] res = new DictionaryEntry[entries.length];
        for (int i = 0; i < entries.length; i++) {
            res[i] = this.get(entries[i]);
        }
        return new DictionaryResult(term, res);
    }

    public DictionaryEntry get(int entry) {
        final String name = String.format("%03x", entry / this.index.getShardSize());
        final DictionaryShard shard = this.shard.getShard(name);
        return shard.get(entry % this.index.getShardSize());
    }

    public static String normalize(String term) {
        final StringBuilder n = new StringBuilder();
        n.ensureCapacity(term.length());
        boolean lastS = true;
        boolean lastD = false;

        // decompose accents and stuff
        // convert similar characters with only stylistic differences
        // convert all whitespace to the ascii equivalent (incl nbsp,em-space,en-space,etc->space)
        // other unicode normalization stuff
        // to lowercase (unicode-aware)
        term = Normalizer.normalize(term, Normalizer.Form.NFKD).toLowerCase();
        for (int i = 0; i < term.length(); i++) {
            char r = term.charAt(i);

            // replace smart punctuation
            switch (r) {
                case 0x00ab: r = '"'; break;
                case 0x00bb: r = '"'; break;
                case 0x2010: r = '-'; break;
                case 0x2011: r = '-'; break;
                case 0x2012: r = '-'; break;
                case 0x2013: r = '-'; break;
                case 0x2014: r = '-'; break;
                case 0x2015: r = '-'; break;
                case 0x2018: r = '\''; break;
                case 0x2019: r = '\''; break;
                case 0x201a: r = '\''; break;
                case 0x201b: r = '\''; break;
                case 0x201c: r = '"'; break;
                case 0x201d: r = '"'; break;
                case 0x201e: r = '"'; break;
                case 0x201f: r = '"'; break;
                case 0x2024: r = '.'; break;
                case 0x2032: r = '\''; break;
                case 0x2033: r = '"'; break;
                case 0x2035: r = '\''; break;
                case 0x2036: r = '"'; break;
                case 0x2038: r = '^'; break;
                case 0x2039: r = '\''; break;
                case 0x203a: r = '\''; break;
                case 0x204f: r = ';'; break;
            }

            // collapse whitespace
            if (r == 32 || (r >= 9 && r <= 12)) {
                if (lastS) {
                    continue;
                }
                lastS = true;
                r = 32;
            } else {
                lastS = false;
            }

            // collapse dashes
            if (r == 45) {
                if (lastD) {
                    continue;
                }
                lastD = true;
            } else {
                lastD = false;
            }

            // expand ligatures
            // remove unknown characters/diacritics
            switch (r) {
                case 0xa74f: n.append("oo");  continue;
                case 0x00df: n.append("ss");  continue;
                case 0x00e6: n.append("ae");  continue;
                case 0x0153: n.append("oe");  continue;
                case 0xfb00: n.append("ff");  continue;
                case 0xfb01: n.append("fi");  continue;
                case 0xfb02: n.append("fl");  continue;
                case 0xfb03: n.append("ffi"); continue;
                case 0xfb04: n.append("ffl"); continue;
                case 0xfb05: n.append("ft");  continue;
                case 0xfb06: n.append("st");  continue;
            }
            if (
                (r >= 97 && r <= 122) ||                                         // a-z
                (r >= 48 && r <= 57) ||                                          // 0-9
                (r == 32 || r == 39 || r == 44 || r == 45 || r == 46 || r == 95) // space and ',-._
            ) {
                n.append(r);
            }
        }
        if (lastS && n.length() > 0) {
            // trim trailing whitespace
            n.setLength(n.length() - 1);
        }
        return n.toString();
    }
}
