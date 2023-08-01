/**
 * Copyright 2023 Patrick Gaskin
 * Requires a relatively recent version of the Chromium WebView.
 */
"use strict";

globalThis["dict"] = function () {
    class Dictionary {
        /** @type {string}          */ #base
        /** @type {DictionaryIndex} */ #index
        /** @type {WeakMap}         */ #pending
        /** @type {WeakMap}         */ #cache // note: Map preserves insert order

        constructor(base, index) {
            this.#base = base
            this.#index = new DictionaryIndex(index)
            this.#pending = new Map()
            this.#cache = new Map()
        }

        async query(term) {
            term = Dictionary.normalize(term)

            // look up the word, plus some basic fallbacks
            let refs = this.#index.lookup(term)
            if (!refs.length) {
                term = term.replace(/'s$/, "")
                refs = this.#index.lookup(term)
            }
            if (!refs.length) {
                term = term.replace(/s$/, "")
                refs = this.#index.lookup(term)
            }
            if (!refs.length) {
                term = term.replace(/-/g, "")
                refs = this.#index.lookup(term)
            }
            if (!refs.length) {
                term = term.replace(/$/, "-")
                refs = this.#index.lookup(term)
            }
            if (!refs.length) {
                term = term.replace(/ly$/, "")
                refs = this.#index.lookup(term)
            }
            if (!refs.length) {
                term = term.replace(/ing$/, "")
                refs = this.#index.lookup(term)
            }
            if (!refs.length) {
                return null
            }

            const entries = await Promise.all(refs.map(ref => this.get(ref)))
            return new DictionaryResult(term, entries)
        }

        async get(ref) {
            let obj = this.#cache.get(ref.shard)
            if (!obj) {
                let p = this.#pending.get(ref.shard)
                if (!p) {
                    p = read(`${this.#base}/${ref.shard}`)
                    this.#pending.set(ref.shard, p)
                }
                obj = new DictionaryShard(await p, this.#index.shardSize)
            }
            if (this.#cache.size > 12) {
                this.#cache.delete(this.#cache.keys().next().value)
            }
            this.#cache.delete(ref.shard)
            this.#cache.set(ref.shard, obj)
            this.#pending.delete(ref.shard)

            return new DictionaryEntry(obj.get(ref.index))
        }

        static async load(base) {
            return new Dictionary(base, await read(`${base}/index`))
        }

        static normalize(term) {
            // decompose accents and stuff
            // convert similar characters with only stylistic differences
            // other unicode normalization stuff
            term = term.normalize("NFKD")

            // to lowercase (unicode-aware)
            term = term.toLowerCase()

            // normalize whitespace
            term = term.trim().replace(/\s+/g, " ")

            // replace smart punctuation
            term = term.split("").map(c => {
                switch (c) {
                    case "\u00ab": return `"`
                    case "\u00bb": return `"`
                    case "\u2010": return `-`
                    case "\u2011": return `-`
                    case "\u2012": return `-`
                    case "\u2013": return `-`
                    case "\u2014": return `-`
                    case "\u2015": return `-`
                    case "\u2018": return `'`
                    case "\u2019": return `'`
                    case "\u201a": return `'`
                    case "\u201b": return `'`
                    case "\u201c": return `"`
                    case "\u201d": return `"`
                    case "\u201e": return `"`
                    case "\u201f": return `"`
                    case "\u2024": return `.`
                    case "\u2032": return `'`
                    case "\u2033": return `"`
                    case "\u2035": return `'`
                    case "\u2036": return `"`
                    case "\u2038": return `^`
                    case "\u2039": return `'`
                    case "\u203a": return `'`
                    case "\u204f": return `;`
                    default: return c
                }
            }).join("")

            // expand ligatures
            term = term.split("").map(c => {
                switch (c) {
                    case "\ua74f": return `oo`
                    case "\u00df": return `ss`
                    case "\u00e6": return `ae`
                    case "\u0153": return `oe`
                    case "\ufb00": return `ff`
                    case "\ufb01": return `fi`
                    case "\ufb02": return `fl`
                    case "\ufb03": return `ffi`
                    case "\ufb04": return `ffl`
                    case "\ufb05": return `ft`
                    case "\ufb06": return `st`
                    case "\u2025": return `..`
                    case "\u2026": return `...`
                    case "\u2042": return `***`
                    case "\u2047": return `??`
                    case "\u2048": return `?!`
                    case "\u2049": return `!?`
                    default: return c
                }
            }).join("")

            // normalize dashes
            term = term.replace(/-+/g, "-")

            // remove unknown characters/diacritics
            // note: since we decomposed diacritics, this will leave the base char
            term = term.split("").filter(c => `abcdefghiklmnopqrstuvwxyz0123456789 -'_.,`.includes(c)).join("")

            return term
        }
    }

    class DictionaryResult extends Array {
        constructor(term, entries) {
            super(...entries)
            this.term = term // term which matched
            this.sortRelevance()
        }

        sortRelevance() {
            // sort the entries by relevance (since they aren't inherently ordered in the dictionary)
            this.sort((a, b) => {
                // exact matches
                if (a.name == this.term && b.name != this.term)
                    return -1
                if (a.name != this.term && b.name == this.term)
                    return 1

                const aVar = a.meaningGroups.flatMap(g => g.wordVariants).map(x => x.toLowerCase())
                const bVar = b.meaningGroups.flatMap(g => g.wordVariants).map(x => x.toLowerCase())

                // exact variant matches
                if (aVar.includes(this.term) && !bVar.includes(this.term))
                    return -1
                if (!aVar.includes(this.term) && bVar.includes(this.term))
                    return 1

                const aHead = a.name.toLowerCase()
                const bHead = b.name.toLowerCase()

                // case-insensitive headword matches
                if (aHead == this.term && bHead != this.term)
                    return -1
                if (aHead != this.term && bHead == this.term)
                    return 1

                // non-abbreviations
                if (aHead == a.name && bHead != b.name)
                    return -1
                if (aHead != a.name && bHead == b.name)
                    return 1

                // more meaning groups
                if (a.meaningGroups.length > b.meaningGroups.length)
                    return -1
                if (a.meaningGroups.length < b.meaningGroups.length)
                    return 1

                const aN = a.meaningGroups.reduce((acc, cur) => acc + cur.meanings.length, 0)
                const bN = b.meaningGroups.reduce((acc, cur) => acc + cur.meanings.length, 0)

                // more meanings
                if (aN > bN)
                    return -1
                if (aN < bN)
                    return 1

                // common prefix with headword
                if (aHead.startsWith(this.term) && !bHead.startsWith(this.term))
                    return -1
                if (!aHead.startsWith(this.term) && bHead.startsWith(this.term))
                    return 1

                return a.name.localeCompare(b.name)
            })

            // sort meaning groups by relevance
            for (const entry of this) {
                entry.meaningGroups.sort((a, b) => {
                    const aVar = a.wordVariants.map(x => x.toLowerCase())
                    const bVar = b.wordVariants.map(x => x.toLowerCase())

                    // sort exact word form matches first
                    if (aVar.includes(this.term) && !bVar.includes(this.term))
                        return -1
                    if (!aVar.includes(this.term) && bVar.includes(this.term))
                        return 1

                    return 0
                })
            }
        }
    }

    class DictionaryEntry {
        constructor(w) {
            this.name = typeof w["w"] == "string" ? w["w"] : ""
            this.pronunciation = typeof w["p"] == "string" ? w["w"] : ""
            this.meaningGroups = (typeof w["m"] == "object" ? w["m"] || []: []).map(m => ({
                info: typeof m["i"] == "object" ? m["i"] || []: [],
                meanings: (typeof m["m"] == "object" ? m["m"] : []).map(m => ({
                    tags: typeof m["t"] == "object" ? m["t"] : [],
                    text: typeof m["x"] == "string" ? m["x"] : "",
                    examples: typeof m["s"] == "object" ? m["s"] || []: [],
                })),
                wordVariants: typeof m["v"] == "object" ? m["v"] || []: [],
            }))
            this.info = typeof w["i"] == "string" ? w["i"] : ""
            this.source = typeof w["s"] == "string" ? w["s"] : ""
        }
    }

    class DictionaryRef {
        /** @type {number} */ #shard
        /** @type {number} */ #index

        constructor(shard, index) {
            this.#shard = shard
            this.#index = index
        }

        get shard() {
            return this.#shard.toString(16).padStart(3, "0")
        }

        get index() {
            return this.#index
        }

        toString() {
            return `${this.shard}/${this.index}`
        }
    }

    class DictionaryIndex {
        /** @type {number}        */ #shardSize
        /** @type {number[]}      */ #bucketCounts
        /** @type {Uint8Array[]}  */ #bucketWords
        /** @type {DataView[]}    */ #bucketIndexes
        /** @type {TextEncoder[]} */ #enc

        constructor(buf) {
            const dv = new DataView(buf)

            let n = 8
            this.#shardSize = dv.getUint32(0)
            this.#bucketCounts = new Array(dv.getUint32(4))

            for (let i = 0; i < this.#bucketCounts.length; i++) {
                this.#bucketCounts[i] = dv.getUint32(n)
                n += 4
            }

            this.#bucketWords = this.#bucketCounts.map((count, bucket) => {
                const len = bucket+1
                const arr = new Uint8Array(buf, n, count*len)
                n += count*len
                return arr
            })

            this.#bucketIndexes = this.#bucketCounts.map(count => {
                const dv = new DataView(buf, n, count*4)
                n += count*4
                return dv
            })

            this.#enc = new TextEncoder()
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
            if (lo == -1) {
                return []
            }

            const refs = new Array(hi-lo+1)
            for (let i = lo; i <= hi; i++) {
                const ent = indexes.getUint32(i*4)
                const ref = new DictionaryRef(Math.floor(ent/this.#shardSize), ent%this.#shardSize)
                refs[i-lo] = ref
            }
            return refs
        }

        get shardSize() {
            return this.#shardSize
        }
    }

    class DictionaryShard {
        /** @type {number[]}    */ #offsets
        /** @type {TextDecoder} */ #dec
        /** @type {ArrayBuffer} */ #buf

        constructor(buf, shardSize) {
            const dv = new DataView(buf)

            this.#offsets = new Array(shardSize).fill(0).map((v, i) => dv.getUint32(i*4))
            this.#dec = new TextDecoder("utf-8")
            this.#buf = buf
        }

        get(index) {
            if (index > this.#offsets.length) {
                return null
            }
            return JSON.parse(this.#dec.decode(this.#buf.slice(this.#offsets[index], this.#offsets[index+1])))
        }
    }


    function binarySearch(n, cmp) {
        let lo = 0
        let hi = n - 1
        while (lo <= hi) {
            const mi = Math.floor((lo + hi) / 2)
            const cv = cmp(mi)
            if (cv == 0) {
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
        if (lo != -1) {
            while (lo > 0 && cmp(lo-1) == 0) {
                lo--
            }
            while (hi < n-1 && cmp(hi+1) == 0) {
                hi++
            }
        }
        return [lo, hi]
    }

    async function read(path) {
        const resp = await fetch(path, {
            cache: "no-store",
        })
        if (resp.status == 404) {
            throw new Error(`${path} not found`)
        } else if (resp.status != 200) {
            throw new Error(`dict index response status ${resp.status} (${resp.statusText})`)
        }
        return resp.arrayBuffer()
    }

    const cache = new Map()
    const pending = new Map()

    async function dictionary(base) {
        const url = new URL(base, globalThis.location).href

        let dict = cache.get(url)
        if (!dict) {
            let p = pending.get(url)
            if (!p) {
                p = Dictionary.load(url)
                pending.set(url)
            }
            dict = await p
            cache.set(url, dict)
            pending.delete(url)
        }
        return dict
    }

    return Object.assign(dictionary, {
        Dictionary,
    })
}()
