package net.pgaskin.dictionary;

import java.nio.ByteBuffer;

public class DictionaryShard {
    public interface Provider {
        DictionaryShard getShard(String shard);
    }

    private final ByteBuffer data;

    public DictionaryShard(ByteBuffer buf) {
        this.data = buf.slice();
    }

    public DictionaryEntry get(int index) {
        final int offset = this.data.getInt(index * 4);
        final ByteBuffer buf = this.data.slice();
        buf.position(offset);
        return new DictionaryEntry(buf);
    }
}
