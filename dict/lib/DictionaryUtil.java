package net.pgaskin.dictionary;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.function.Function;
import java.util.function.IntUnaryOperator;

class DictionaryUtil {
    static class Buffer {
        public final ByteBuffer dv;

        Buffer(ByteBuffer buf) {
            this.dv = buf.slice();
        }

        int u32() {
            return dv.getInt();
        }

        ByteBuffer buf(int n) {
            ByteBuffer x = dv.slice();
            x.limit(n);
            dv.position(dv.position() + n);
            return x;
        }

        String str() {
            byte[] x = new byte[u32()];
            dv.get(x);
            return new String(x, StandardCharsets.UTF_8);
        }

        String[] arrStr() {
            String[] x = new String[u32()];
            for (int i = 0; i < x.length; i++) {
                x[i] = str();
            }
            return x;
        }

        int[] arrU32() {
            int[] x = new int[u32()];
            for (int i = 0; i < x.length; i++) {
                x[i] = u32();
            }
            return x;
        }
    }

    static Buffer wrapBuffer(ByteBuffer buf) {
        return new DictionaryUtil.Buffer(buf);
    }

    static <K, V> Function<K, V> makeCache(Function<K, V> get, int max) {
        LinkedHashMap<K, V> cache = new LinkedHashMap<K, V>(max, 0.75f, true) {
            protected boolean removeEldestEntry(Map.Entry<K, V> eldest) {
                return this.size() > max;
            }
        };
        return k -> {
            synchronized (cache) {
                V v = cache.get(k);
                if (v == null) {
                    v = get.apply(k);
                    cache.put(k, v);
                }
                return v;
            }
        };
    }

    static int binarySearch(int n, IntUnaryOperator cmp) {
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

    static int[] binarySearchRange(int n, IntUnaryOperator cmp) {
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

    static String removeChar(String s, char c) {
        final StringBuilder b = new StringBuilder();
        b.ensureCapacity(s.length());
        for (int i = 0; i < s.length(); i++) {
            char x = s.charAt(i);
            if (x != c) {
                b.append(x);
            }
        }
        return b.toString();
    }
}
