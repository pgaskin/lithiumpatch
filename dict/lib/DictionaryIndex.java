package net.pgaskin.dictionary;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;

import static net.pgaskin.dictionary.DictionaryUtil.*;

public class DictionaryIndex {
    private final int shardSize;
    private final int[] bucketCounts;
    private final ByteBuffer[] bucketWords;
    private final ByteBuffer[] bucketIndexes;

    public DictionaryIndex(ByteBuffer buf) {
        final DictionaryUtil.Buffer b = wrapBuffer(buf);
        this.shardSize = b.u32();
        this.bucketCounts = b.arrU32();
        this.bucketWords = new ByteBuffer[this.bucketCounts.length];
        for (int bucket = 0; bucket < this.bucketWords.length; bucket++) {
            final int count = this.bucketCounts[bucket];
            final int len = bucket + 1;
            this.bucketWords[bucket] = b.buf(count * len);
        }
        this.bucketIndexes = new ByteBuffer[this.bucketCounts.length];
        for (int bucket = 0; bucket < this.bucketIndexes.length; bucket++) {
            final int count = this.bucketCounts[bucket];
            this.bucketIndexes[bucket] = b.buf(count * 4);
        }
    }

    public int[] lookup(String term) {
        final ByteBuffer arr = ByteBuffer.wrap(term.getBytes(StandardCharsets.UTF_8));
        final int len = arr.limit();
        if (len >= this.bucketCounts.length) {
            return new int[0];
        }

        final int bucket = len - 1;
        final int count = this.bucketCounts[bucket];
        final ByteBuffer words = this.bucketWords[bucket];
        final ByteBuffer indexes = this.bucketIndexes[bucket];

        final int[] res = binarySearchRange(count, i -> {
            final ByteBuffer word = words.slice();
            word.position(i*len);
            word.limit(word.position() + len);
            return arr.compareTo(word);
        });
        final int lo = res[0], hi = res[1];
        if (lo == -1) {
            return new int[0];
        }

        final int[] es = new int[hi-lo+1];
        for (int i = lo; i <= hi; i++) {
            es[i - lo] = indexes.getInt(indexes.position() + i*4);
        }
        return es;
    }

    public int getShardSize() {
        return this.shardSize;
    }
}
