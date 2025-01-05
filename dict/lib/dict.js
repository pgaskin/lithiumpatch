/**
 * Copyright 2023 Patrick Gaskin
 * Requires a relatively recent version of the Chromium WebView.
 */
"use strict";

const getDictionaryCached = makeSingleFlightCache(async base => await Dictionary.load(async fn => {
    const url = new URL(fn, base)
    const resp = await fetch(url, {
        cache: "no-store",
    })
    if (resp.status === 404) {
        throw new Error(`${url} not found`)
    } else if (resp.status !== 200) {
        throw new Error(`${url} response status ${resp.status} (${resp.statusText})`)
    }
    return resp.arrayBuffer()
}))

export default async function dictionary(base) {
    return getDictionaryCached(new URL(base + "/", import.meta.url).href)
}

export class Dictionary {
    /** @type {DictionaryIndex}                             */ #index
    /** @type {(shard: string) => Promise<DictionaryShard>} */ #shard

    constructor(index, shard) {
        this.#index = index
        this.#shard = shard
    }

    static async load(read, shardCacheMax = 14) {
        const index = new DictionaryIndex(await read("index"))
        const shard = makeSingleFlightCache(async shard => new DictionaryShard(await read(shard)), shardCacheMax)
        return new Dictionary(index, shard)
    }

    async query(term, normalized = false) {
        if (!normalized) {
            term = Dictionary.normalize(term)
        }
        if (!term.length) {
            return new DictionaryResult(term);
        }

        // look up the word, plus some basic fallbacks
        const origTerm = term
        let entries = this.#index.lookup(term)
        if (!entries.length && term.endsWith("'s")) {
            term = term.substring(0, term.length - "'s".length);
            entries = this.#index.lookup(term)
        }
        if (!entries.length && term.endsWith("s")) {
            term = term.substring(0, term.length - "s".length);
            entries = this.#index.lookup(term)
        }
        if (!entries.length && term.includes("-")) {
            term = removeChar(term, "-")
            entries = this.#index.lookup(term)
        }
        if (!entries.length && term.endsWith("ly")) {
            term = term.substring(0, term.length - "ly".length);
            entries = this.#index.lookup(term)
        }
        if (!entries.length && term.endsWith("ing")) {
            term = term.substring(0, term.length - "ing".length);
            entries = this.#index.lookup(term)
        }
        if (!entries.length) {
            return new DictionaryResult(origTerm);
        }

        const res = await Promise.all(entries.map(x => this.#get(x)))
        return new DictionaryResult(term, ...res)
    }

    async lookup(word) {
        return await Promise.all(this.#index.lookup(word).map(x => this.#get(x)))
    }

    async #get(entry) {
        const name = Math.floor(entry / this.#index.shardSize).toString(16).padStart(3, "0")
        const shard = await this.#shard(name)
        return shard.get(entry % this.#index.shardSize)
    }

    autocomplete(term, limit = -1, normalized = false) {
        if (!normalized) {
            term = Dictionary.normalize(term)
        }
        if (!term.length){
            return []
        }
        return this.#index.lookupPrefix(term, limit)
    }

    static normalize(term) {
        let n = ""
        let lastS = true // trim leading whitespace
        let lastD = false

        // decompose accents and stuff
        // convert similar characters with only stylistic differences
        // convert all whitespace to the ascii equivalent (incl nbsp,em-space,en-space,etc->space)
        // other unicode normalization stuff
        // to lowercase (unicode-aware)
        for (let r of term.normalize("NFKD").toLowerCase()) {
            r = r.charCodeAt(0)

            // replace smart punctuation
            switch (r) {
                case 0x00ab: r = `"`.charCodeAt(0); break
                case 0x00bb: r = `"`.charCodeAt(0); break
                case 0x2010: r = `-`.charCodeAt(0); break
                case 0x2011: r = `-`.charCodeAt(0); break
                case 0x2012: r = `-`.charCodeAt(0); break
                case 0x2013: r = `-`.charCodeAt(0); break
                case 0x2014: r = `-`.charCodeAt(0); break
                case 0x2015: r = `-`.charCodeAt(0); break
                case 0x2018: r = `'`.charCodeAt(0); break
                case 0x2019: r = `'`.charCodeAt(0); break
                case 0x201a: r = `'`.charCodeAt(0); break
                case 0x201b: r = `'`.charCodeAt(0); break
                case 0x201c: r = `"`.charCodeAt(0); break
                case 0x201d: r = `"`.charCodeAt(0); break
                case 0x201e: r = `"`.charCodeAt(0); break
                case 0x201f: r = `"`.charCodeAt(0); break
                case 0x2024: r = `.`.charCodeAt(0); break
                case 0x2032: r = `'`.charCodeAt(0); break
                case 0x2033: r = `"`.charCodeAt(0); break
                case 0x2035: r = `'`.charCodeAt(0); break
                case 0x2036: r = `"`.charCodeAt(0); break
                case 0x2038: r = `^`.charCodeAt(0); break
                case 0x2039: r = `'`.charCodeAt(0); break
                case 0x203a: r = `'`.charCodeAt(0); break
                case 0x204f: r = `;`.charCodeAt(0); break
            }

            // collapse whitespace
            if (r === 32 || (r >= 9 && r <= 12)) {
                if (lastS) {
                    continue
                }
                lastS = true
                r = 32
            } else {
                lastS = false
            }

            // collapse dashes
            if (r === 45) {
                if (lastD) {
                    continue
                }
                lastD = true
            } else {
                lastD = false
            }

            // expand ligatures
            // remove unknown characters/diacritics
            switch (r) {
                case 0xa74f: n += `oo`;  continue
                case 0x00df: n += `ss`;  continue
                case 0x00e6: n += `ae`;  continue
                case 0x0153: n += `oe`;  continue
                case 0xfb00: n += `ff`;  continue
                case 0xfb01: n += `fi`;  continue
                case 0xfb02: n += `fl`;  continue
                case 0xfb03: n += `ffi`; continue
                case 0xfb04: n += `ffl`; continue
                case 0xfb05: n += `ft`;  continue
                case 0xfb06: n += `st`;  continue
            }
            if (
                (r >= 97 && r <= 122) ||                                               // a-z
                (r >= 48 && r <= 57) ||                                                // 0-9
                (r === 32 || r === 39 || r === 44 || r === 45 || r === 46 || r === 95) // space and ',-._
            ) {
                n += String.fromCharCode(r)
            }
        }
        if (lastS && n.length > 0) {
            // trim trailing whitespace
            n = n.slice(0, -1)
        }
        return n
    }
}

export class DictionaryResult extends Array {
    constructor(term, ...entries) {
        super(...entries)
        this.term = term // term which matched
        this.sort()
    }

    sort(compareFn = undefined) {
        if (compareFn !== undefined) {
            super.sort(compareFn)
            return
        }

        // sort the entries by relevance (since they aren't inherently ordered in the dictionary)
        this.sort((a, b) => {
            // exact matches
            if (a.name === this.term && b.name !== this.term) return -1
            if (a.name !== this.term && b.name === this.term) return 1

            const aVar = a.meaningGroups.flatMap(g => g.wordVariants).map(x => x.toLowerCase()).includes(this.term)
            const bVar = b.meaningGroups.flatMap(g => g.wordVariants).map(x => x.toLowerCase()).includes(this.term)

            // exact variant matches
            if (aVar && !bVar) return -1
            if (!aVar && bVar) return 1

            const aHead = a.name.toLowerCase()
            const bHead = b.name.toLowerCase()

            // case-insensitive headword matches
            if (aHead === this.term && bHead !== this.term) return -1
            if (aHead !== this.term && bHead === this.term) return 1

            // non-abbreviations
            if (aHead === a.name && bHead !== b.name) return -1
            if (aHead !== a.name && bHead === b.name) return 1

            // more meaning groups
            if (a.meaningGroups.length > b.meaningGroups.length) return -1
            if (a.meaningGroups.length < b.meaningGroups.length) return 1

            const aN = a.meaningGroups.reduce((acc, cur) => acc + cur.meanings.length, 0)
            const bN = b.meaningGroups.reduce((acc, cur) => acc + cur.meanings.length, 0)

            // more meanings
            if (aN > bN) return -1
            if (aN < bN) return 1

            // common prefix with headword
            if (aHead.startsWith(this.term) && !bHead.startsWith(this.term)) return -1
            if (!aHead.startsWith(this.term) && bHead.startsWith(this.term)) return 1

            return a.name.localeCompare(b.name)
        })

        // sort meaning groups by relevance
        for (const entry of this) {
            entry.meaningGroups.sort((a, b) => {
                const aVar = a.wordVariants.map(x => x.toLowerCase()).includes(this.term)
                const bVar = b.wordVariants.map(x => x.toLowerCase()).includes(this.term)

                // exact variant matches
                if (aVar && !bVar) return -1
                if (!aVar && bVar) return 1

                return 0
            })
        }
    }

    toString(showExamples = true, showEntryInfo = true) {
        let s = ""
        if (this.term.length) {
            s += this.term
            s += "\n"
        }
        for (const e of this) {
            s += "\n"
            s += e.toString(showExamples, showEntryInfo)
        }
        return s
    }
}

export class DictionaryIndex {
    /** @type {number}       */ #shardSize
    /** @type {number[]}     */ #bucketCounts
    /** @type {Uint8Array[]} */ #bucketWords
    /** @type {DataView[]}   */ #bucketIndexes
    /** @type {TextEncoder}  */ #enc
    /** @type {TextDecoder}  */ #dec

    constructor(buf) {
        const b = wrapBuffer(buf)
        this.#shardSize = b.u32()
        this.#bucketCounts = b.arr(b.u32)
        this.#bucketWords = b.arr(
            bucket => {
                const count = this.#bucketCounts[bucket]
                const len = bucket + 1
                return new Uint8Array(b.buf(count * len))
            },
            this.#bucketCounts.length,
        )
        this.#bucketIndexes = b.arr(
            bucket => {
                const count = this.#bucketCounts[bucket]
                return new DataView(b.buf(count * 4))
            },
            this.#bucketCounts.length,
        )
        this.#enc = new TextEncoder()
        this.#dec = new TextDecoder()
    }

    lookup(term) {
        const arr = this.#enc.encode(term)
        const len = arr.length
        if (len >= this.#bucketCounts.length) {
            return []
        }

        const bucket = len-1
        const count = this.#bucketCounts[bucket]
        const words = this.#bucketWords[bucket]
        const indexes = this.#bucketIndexes[bucket]

        const [lo, hi] = binarySearchRange(count, i => {
            for (let c = 0; c < len; c++) {
                const x = arr[c]
                const y = words[i*len + c]
                if (x < y) {
                    return -1
                }
                if (x > y) {
                    return 1
                }
            }
            return 0
        })
        if (lo === -1) {
            return []
        }

        const es = new Array(hi-lo+1)
        for (let i = lo; i <= hi; i++) {
            es[i-lo] = indexes.getUint32(i*4)
        }
        return es
    }

    lookupPrefix(term, limit = -1) {
        const arr = this.#enc.encode(term)
        const len = arr.length
        if (len >= this.#bucketCounts.length) {
            return []
        }

        const ws = []
        for (let len1 = len; len1 <= this.#bucketCounts.length; len1++) {
            const bucket = len1-1
            const count = this.#bucketCounts[bucket]
            const words = this.#bucketWords[bucket]

            const [lo, hi] = binarySearchRange(count, i => {
                for (let c = 0; c < len; c++) {
                    const x = arr[c]
                    const y = words[i*len1 + c]
                    if (x < y) {
                        return -1
                    }
                    if (x > y) {
                        return 1
                    }
                }
                return 0
            })
            if (lo === -1) {
                continue
            }

            let last
            for (let i = lo; i <= hi; i++) {
                const w = this.#dec.decode(words.slice(i*len1, i*len1 + len1))
                if (last !== w) {
                    ws.push(w)
                }
                if (ws.length == limit) {
                    return ws
                }
                last = w
            }
        }
        return ws
    }

    get shardSize() {
        return this.#shardSize
    }
}

export class DictionaryShard {
    /** @type {DataView} */ #data

    constructor(buf) {
        this.#data = new DataView(buf)
    }

    get(index) {
        const offset = this.#data.getUint32(index * 4)
        const buf = this.#data.buffer.slice(offset)
        return new DictionaryEntry(buf)
    }
}

export class DictionaryEntry {
    constructor(buf) {
        const b = wrapBuffer(buf)
        this.name = b.str()
        this.pronunciation = b.str()
        this.meaningGroups = b.arr(i => ({
            info: b.arr(b.str),
            meanings: b.arr(i => ({
                tags: b.arr(b.str),
                text: b.str(),
                examples: b.arr(b.str),
            })),
            wordVariants: b.arr(b.str),
        }))
        this.info = b.str()
        this.source = b.str()
    }

    toString(showExamples = true, showEntryInfo = true) {
        let s = ""
        s += this.name
        if (this.pronunciation.length) {
            s += " \u00b7 "
            s += this.pronunciation
        }
        s += "\n"
        for (const g of this.meaningGroups) {
            if (g.info.length) {
                s += "  "
                s += g.info.join(" \u2014 ")
                s += "\n"
            }
            let n = 0
            for (const m of g.meanings) {
                s += "  "
                s += (++n).toString().padStart(4, " ")
                s += ". "
                if (m.tags.length) {
                    s += "["
                    s += m.tags.join("] [")
                    s += "] "
                }
                s += m.text
                s += "\n"
                if (showExamples) {
                    for (const x of m.examples) {
                        s += "        - "
                        s += x
                        s += "\n"
                    }
                }
            }
        }
        if (showEntryInfo && this.info.length) {
            s += "  "
            s += this.info
            s += "\n"
        }
        if (this.source.length) {
            s += this.source
            s += "\n"
        }
        return s
    }
}

function wrapBuffer(b) {
    let c = 0
    const dv = new DataView(b)
    const td = new TextDecoder("utf-8")
    const u32 = () => {
        const x = dv.getUint32(c)
        c += 4
        return x
    }
    const buf = n => {
        const x = dv.buffer.slice(c, c + n)
        c += n
        return x
    }
    const str = () => {
        return td.decode(buf(u32()))
    }
    const arr = (fn, n = undefined) => {
        const x = new Array(n ?? u32())
        for (let i = 0; i < x.length; i++) {
            x[i] = fn(i)
        }
        return x
    }
    return { u32, buf, str, arr }
}

function makeSingleFlightCache(get, max = 0) {
    const cache = new Map()
    const pending = new Map()

    return async key => {
        let obj = cache.get(key)
        if (!obj) {
            let p = pending.get(key)
            if (!p) {
                p = get(key)
                pending.set(key, p)
            }
            obj = await p
        }
        if (max > 0 && cache.size > max) {
            cache.delete(cache.keys().next().value)
        }
        cache.delete(key)
        cache.set(key, obj)
        pending.delete(key)
        return obj
    }
}

function binarySearch(n, cmp) {
    let lo = 0
    let hi = n - 1
    while (lo <= hi) {
        const mi = Math.floor((lo + hi) / 2)
        const cv = cmp(mi)
        if (!cv) {
            return mi
        } else if (cv < 0) {
            hi = mi - 1
        } else {
            lo = mi + 1
        }
    }
    return -1
}

function binarySearchRange(n, cmp) {
    let lo = binarySearch(n, cmp)
    let hi = lo
    if (lo !== -1) {
        while (lo > 0 && !cmp(lo-1)) {
            lo--
        }
        while (hi < n-1 && !cmp(hi+1)) {
            hi++
        }
    }
    return [lo, hi]
}

function removeChar(s, c) {
    let t = ""
    for (const x of s) {
        if (x !== c) {
            t += x;
        }
    }
    return t;
}
