package net.pgaskin.dictionary;

import java.nio.ByteBuffer;

import static net.pgaskin.dictionary.DictionaryUtil.*;

public class DictionaryEntry {
    public final String name;
    public final String pronunciation;
    public final MeaningGroup[] meaningGroups;
    public final String info;
    public final String source;

    DictionaryEntry(ByteBuffer buf) {
        final DictionaryUtil.Buffer b = wrapBuffer(buf);
        this.name = b.str();
        this.pronunciation = b.str();
        this.meaningGroups = new MeaningGroup[b.u32()];
        for (int i = 0; i < this.meaningGroups.length; i++) {
            this.meaningGroups[i] = new MeaningGroup(b);
        }
        this.info = b.str();
        this.source = b.str();
    }

    public static class MeaningGroup {
        public final String[] info;
        public final Meaning[] meanings;
        public final String[] wordVariants;

        MeaningGroup(DictionaryUtil.Buffer b) {
            this.info = b.arrStr();
            this.meanings = new Meaning[b.u32()];
            for (int i = 0; i < this.meanings.length; i++) {
                this.meanings[i] = new Meaning(b);
            }
            this.wordVariants = b.arrStr();
        }

        public static class Meaning {
            public final String[] tags;
            public final String text;
            public final String[] examples;

            Meaning(DictionaryUtil.Buffer b) {
                this.tags = b.arrStr();
                this.text = b.str();
                this.examples = b.arrStr();
            }
        }
    }

    public String toString() {
        return this.toString(true, true);
    }

    public String toString(boolean showExamples, boolean showEntryInfo) {
        final StringBuilder s = new StringBuilder();
        s.append(this.name);
        if (!this.pronunciation.isEmpty()) {
            s.append(" \u00b7 ");
            s.append(this.pronunciation);
        }
        s.append("\n");
        for (final MeaningGroup g : this.meaningGroups) {
            if (g.info.length != 0) {
                s.append("  ");
                s.append(String.join(" \u2014 ", g.info));
                s.append("\n");
            }
            int n = 0;
            for (final MeaningGroup.Meaning m : g.meanings) {
                s.append("  ");
                s.append(String.format("%4d", ++n));
                s.append(". ");
                if (m.tags.length != 0) {
                    s.append("[");
                    s.append(String.join("] [", m.tags));
                    s.append("] ");
                }
                s.append(m.text);
                s.append("\n");
                if (showExamples) {
                    for (final String x : m.examples) {
                        s.append("        - ");
                        s.append(x);
                        s.append("\n");
                    }
                }
            }
        }
        if (showEntryInfo && !this.info.isEmpty()) {
            s.append("  ");
            s.append(this.info);
            s.append("\n");
        }
        if (!this.source.isEmpty()) {
            s.append(this.source);
            s.append("\n");
        }
        return s.toString();
    }
}
