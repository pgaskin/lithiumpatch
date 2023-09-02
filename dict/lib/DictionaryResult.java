package net.pgaskin.dictionary;

import java.util.ArrayList;
import java.util.Arrays;

public class DictionaryResult extends ArrayList<DictionaryEntry> {
    public final String term;

    public DictionaryResult(String term) {
        super();
        this.term = term;
    }

    public DictionaryResult(String term, DictionaryEntry[] entries) {
        super(Arrays.asList(entries));
        this.term = term;
        this.sort();
    }

    public void sort() {
        // sort the entries by relevance (since they aren't inherently ordered in the dictionary)
        super.sort((a, b) -> {
            // exact matches
            if (a.name.equals(this.term) && !b.name.equals(this.term)) return -1;
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

        // sort meaning groups by relevance
        for (final DictionaryEntry entry : this) {
            Arrays.sort(entry.meaningGroups, (a, b) -> {
                boolean aVar = false, bVar = false;
                for (String v : a.wordVariants) {
                    if (v.equalsIgnoreCase(this.term)) {
                        aVar = true;
                        break;
                    }
                }
                for (String v : b.wordVariants) {
                    if (v.equalsIgnoreCase(this.term)) {
                        bVar = true;
                        break;
                    }
                }

                // exact variant matches
                if (aVar && !bVar) return -1;
                if (!aVar && bVar) return 1;

                return 0;
            });
        }
    }

    public String toString() {
        return this.toString(true, true);
    }

    public String toString(boolean showExamples, boolean showEntryInfo) {
        final StringBuilder s = new StringBuilder();
        if (!this.term.isEmpty()) {
            s.append(this.term);
            s.append("\n");
        }
        for (final DictionaryEntry e : this) {
            s.append("\n");
            s.append(e.toString(showExamples, showEntryInfo));
        }
        return s.toString();
    }
}
