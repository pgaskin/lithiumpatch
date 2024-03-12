/**
 * Copyright 2019-2023 Patrick Gaskin
 * Requires a relatively recent version of the Chromium WebView.
 */
"use strict";

var Dictionary = (function() {
    const ENABLE_DEEP_SELECTION = true // if it acts funny, try disabling this

    /**
     * Symbol used by html to mark processed strings.
     */
    const isHtml = Symbol("isHtml")

    /**
     * ES6 template function for safely building HTML, optionally with inline CSS.
     */
    function html(literals, ...subs) {
        const raw = literals.raw.reduce((acc, lit, i) => {
            let sub = subs[--i]
            if (Array.isArray(sub)) {
                sub = sub.map(x => x[isHtml] ? x : (x !== undefined && x !== null) ? escape(x.toString()) : "").join("")
            } else if (sub === undefined || sub === null || sub === false) {
                sub = ""
            } else if (literals.raw[i]?.endsWith("$")) {
                acc = acc.slice(0, -1)
                sub = sub.toString()
            } else if (sub[isHtml]) {
                sub = sub.toString()
            } else {
                sub = escape(sub.toString())
            }
            return acc + sub + lit
        })
        return Object.defineProperties(new String(raw), {
            [isHtml]: {
                value: true,
            },
        })
    }

    /**
     * Escapes a string for use in HTML.
     */
    function escape(str) {
        str = str.replace(/&/g, "&amp;")
        str = str.replace(/>/g, "&gt;")
        str = str.replace(/</g, "&lt;")
        str = str.replace(/"/g, "&quot;")
        str = str.replace(/'/g, "&#39;")
        return str
    }

    /**
     * Listens for Lithium theme changes, returning true if successful. The first
     * theme update is sent on DOMContentLoaded. Note that the default white
     * theme is null.
     */
    function hookLithiumTheme(callback) {
        if ("LithiumThemes" in globalThis) {
            const orig = globalThis.LithiumThemes.set
            globalThis.LithiumThemes.set = theme => {
                window.setTimeout(() => callback(theme), 0)
                return orig(theme)
            }
            return true
        }
        return false
    }

    /**
     * Converts an integer color into a CSS hex value.
     */
    function hexColor(color) {
        return `#${color.toString(16).padStart(6, "0")}`
    }

    /**
     * Theme-aware popup.
     */
    class Popup {
        #root
        #cssBase
        #cssExtra
        #shadow
        #wrapper
        #popup
        #inner
        #content
        #pos
        #posRect
        constructor(css = undefined) {
            this.#root = document.createElement("x-popup")

            this.#cssBase = new CSSStyleSheet()
            this.#cssBase.replaceSync(`
                #wrapper {
                    display: block;
                    position: fixed;
                    left: 16px;
                    right: 16px;
                    height: auto;
                    width: auto;
                    overflow: visible;
                    user-select: none;
                    z-index: 9999999;
                    transition: opacity .1s ease-in;
                }
                #popup {
                    contain: strict;
                    display: block;
                    box-sizing: border-box;
                    position: relative;
                    max-height: 220px;
                    min-height: 50px;
                    height: 30vh;
                    max-width: 480px;
                    width: 100%;
                    overflow: hidden;
                    margin: 16px auto;
                    border: 1px solid transparent;
                    border-radius: 4px;
                }
                #inner {
                    overscroll-behavior: contain;
                    display: block;
                    box-sizing: border-box;
                    position: absolute;
                    overflow-x: hidden;
                    overflow-y: auto;
                    inset: 0;
                    line-height: 1.25;
                    font-size: 14px;
                    font-family: serif;
                }
                #inner::-webkit-scrollbar {
                    display: none;
                }
            `)

            this.#cssExtra = new CSSStyleSheet()
            if (css !== undefined) {
                this.#cssExtra.replace(css)
            }

            this.#shadow = this.#root.attachShadow({mode: "open"})
            this.#shadow.adoptedStyleSheets = [this.#cssBase, this.#cssExtra]

            this.#wrapper = this.#shadow.appendChild(document.createElement("div"))
            this.#wrapper.id = "wrapper"

            this.#popup = this.#wrapper.appendChild(document.createElement("div"))
            this.#popup.id = "popup"

            this.#inner = this.#popup.appendChild(document.createElement("div"))
            this.#inner.id = "inner"

            this.#content = this.#inner.appendChild(document.createElement("div"))

            const applyTheme = (dark = false, bg = undefined, fg = undefined) => {
                this.#popup.style.setProperty("border-color", dark ? "rgba(255, 255, 255, .25)" : "rgba(0, 0, 0, .25)")
                this.#popup.style.setProperty("box-shadow", dark ? "none" : "0 0 8px 0 rgba(0, 0, 0, .25)")
                this.#popup.style.setProperty("background-color", bg !== undefined ? bg : dark ? "#333" : "#fff")
                this.#popup.style.setProperty("color", fg !== undefined ? fg : dark ? "#eee" : "#000")
            }
            applyTheme()

            if (!hookLithiumTheme(theme => {
                if (theme) {
                    applyTheme(theme.bgIsDark, hexColor(theme.backgroundColor), hexColor(theme.textColor))
                } else {
                    applyTheme()
                }
            })) {
                const matcher = window.matchMedia('(prefers-color-scheme: dark)')
                if (matcher) {
                    applyTheme(matcher.matches)
                    matcher.addEventListener("change", ev => {
                        applyTheme(ev.matches)
                    })
                }
            }

            window.addEventListener("resize", () => this.visible && this.move(), true)
            document.addEventListener("scroll", () => this.visible && this.move(), true)
        }

        /**
         * Replaces the contents of the popup, returning a element for future
         * updates (which will continue to take effect until replace is called
         * again).
         */
        replace(initial = "") {
            const el = document.createElement("div")
            el.innerHTML = initial
            this.#content.replaceWith(el)
            this.#content = el
            return el
        }

        /**
         * Shows the popup if it isn't already visible. Returns true if the popup
         * was previously hidden.
         *
         * If getClientRect is set, pos=true means to put the popup near the rect,
         * and pos=false means to put it against the opposite screen edge.
         * Otherwise, pos=true means to put the popup at the top edge, and pos=false
         * means the bottom.
         */
        show(pos = false, getClientRect = undefined) {
            this.#pos = pos
            this.#posRect = getClientRect
            this.move()

            if (!this.visible) {
                this.#wrapper.style.setProperty("opacity", "0", "important")
                document.body.appendChild(this.#root);
                window.setTimeout(() => this.#wrapper.style.setProperty("opacity", "1", "important"), 0)
                return true
            }
            if (this.#root.parentElement.lastElementChild != this.#root) {
                this.#root.parentElement.appendChild(this.#root)
            }
            return false
        }

        /**
         * Hides the popup if it is visible.
         */
        hide() {
            if (this.visible) {
                this.#root.remove()
                this.#pos = false
                this.#posRect = undefined
            }
        }

        /**
         * Updates the popup's position if it is visible.
         */
        move() {
            let top, bot
            if (this.#posRect === undefined) {
                if (this.#pos) {
                    top = 0
                } else {
                    bot = 0
                }
            } else {
                const rect = this.#posRect()
                if (rect.y > document.documentElement.clientHeight / 2) {
                    if (this.#pos) {
                        bot = document.documentElement.clientHeight - rect.y
                    } else {
                        top = 0
                    }
                } else {
                    if (this.#pos) {
                        top = rect.y + rect.height
                    } else {
                        bot = 0
                    }
                }
            }
            this.#wrapper.style.setProperty("top", top === undefined ? "auto" : `${top}px`, "important")
            this.#wrapper.style.setProperty("bottom", bot === undefined ? "auto" : `${bot}px`, "important")
        }

        /**
         * Whether the popup is currently visible.
         */
        get visible() {
            return !!this.#root.parentElement
        }

        /**
         * Gets the Shadow DOM root.
         */
        get shadowRoot() {
            return this.#shadow
        }
    }

    const settings = {
        dict_disabled: 'LithiumApp' in globalThis ? globalThis.LithiumApp.getDictDisabled().split(" ") : [],
        dict_show_examples: 'LithiumApp' in globalThis ? globalThis.LithiumApp.getDictShowExamples() : true,
        dict_show_info: 'LithiumApp' in globalThis ? globalThis.LithiumApp.getDictShowInfo() : true,
        dict_small_font: 'LithiumApp' in globalThis ? globalThis.LithiumApp.getDictSmallFont() : false,
    }

    const init = {
        dict: new URL(document.currentScript.dataset.dict || "./dict.js", document.currentScript.src).href,
        dicts: document.currentScript.dataset.dicts.split(" "),
    }

    const dictPopup = new Popup(`
        section {
            padding: 8px;
            border-top: 1px solid rgba(128,128,128,0.4);
            font-size: ${settings.dict_small_font ? ".85em" : "1em"};
        }
        section:first-of-type {
            border-top: none;
        }
        section > header {
            display: flex;
            fontSize: 1.14em;
            margin-bottom: 4px;
        }
        section > header::after {
            content: '\u00a0';
        }
        section > header > .headword {
            font-weight: bold;
        }
        section > header > .pronunciation {
            opacity: .75;
        }
        section > header > .pronunciation::before {
            content: '\u00a0\u00b7\u00a0';
        }
        section > .meaning-group-info {
            font-style: italic;
            margin-bottom: 4px;
        }
        section > ol.meaning-group-definitions {
            margin: 8px 0 16px;
            padding: 0 0 0 2em;
        }
        section > ol.meaning-group-definitions > li {
            margin: 0 0 4px;
        }
        section > ol.meaning-group-definitions > li > span.tag {
            display: inline-block;
            font-size: .85em;
            vertical-align: baseline;
            padding: .1em .25em;
            background-color: rgba(128,128,128,0.2);
        }
        section > ol.meaning-group-definitions > li > div.example {
            font-style: italic;
            margin-top: 4px;
        }
        section > .entry-info {
            opacity: 0.75;
            margin-top: 8px;
        }
        section > .source {
            font-style: italic;
            font-size: .85em;
            margin-top: 8px;
        }
    `)

    const render = (t, x) => !Array.isArray(x) ? html`
        <section>
            <header>
                <div class="headword">${t}</div>${"\u00a0"}
            </header>
            <div>
                ${x}
            </div>
        </section>
    ` : x.map((x, i) => html`
        <section>
            <header>
                <div class="headword">${x.name}</div>
                ${!!x.pronunciation && html`
                    <div class="pronunciation">${x.pronunciation}</div>
                `}
            </header>
            ${x.meaningGroups.map(x => html`
                ${!!x.info.length && html`
                    <div class="meaning-group-info">${x.info.join(" \u2014 ")}</div>
                `}
                ${!!x.meanings.length && html`
                    <ol class="meaning-group-definitions">
                        ${x.meanings.map(x => html`
                            <li>
                                ${x.tags.map(x => html`
                                    <span class="tag">${x}</span>
                                `)}
                                ${!!x.text.length && html`
                                    <span class="definition">${x.text}</span>
                                `}
                                ${settings.dict_show_examples && x.examples.map(x => html`
                                    <div class="example">${x}</div>
                                `)}
                            </li>
                        `)}
                    </ol>
                `}
            `)}
            ${settings.dict_show_info && !!x.info.length && html`
                <div class="entry-info">${x.info}</div>
            `}
            ${!!x.source.length && html`
                <div class="source">${x.source}</div>
            `}
        </section>
    `).join("")

    import(init.dict).then(({default: dictionary, Dictionary: Dictionary}) => {
        let dictReq
        let dictSem
        let dictDeepHidePending
        let dictDeepSelecting = new Set() // pointer IDs
        const lookup = (txt, rng, isDeepLookup = false) => {

            if (ENABLE_DEEP_SELECTION) {
                // clear the previous hide timer if it's a deep selection
                if (isDeepLookup) {
                    if (dictDeepHidePending !== undefined) {
                        window.clearTimeout(dictDeepHidePending)
                        dictDeepHidePending = undefined
                    }
                }

                // reset the deep selection flag
                dictDeepSelecting.clear()
            }

            // copy the range (so we have a constant reference to the rect)
            rng = rng.cloneRange()

            // set the initial popup contents
            const tt = Dictionary.normalize(txt)
            const el = dictPopup.replace(render(tt, "Loading."))

            // show the popup
            if (dictPopup.show(false, () => rng.getBoundingClientRect()) || (ENABLE_DEEP_SELECTION && isDeepLookup)) {

                // if we're not modifying an existing selection (or it's a deep selection), discard the old semaphore
                dictSem = Promise.resolve()
            }

            // wait for the selection to settle
            //
            // note: we still want the timeout even if it's a deep selection so
            // the popup hide gets called before we show it again
            const ownSettle = window.setTimeout(() => {

                // wait for the previous lookup to finish, then continue ours
                dictSem = dictSem.finally(async () => {

                    // check if we've been replaced or canceled
                    if (ownSettle !== dictReq) {
                        if (ENABLE_DEEP_SELECTION && isDeepLookup) {
                            // deep selection - we don't care if it's a deep selection
                        } else {
                            return
                        }
                    }

                    if (ENABLE_DEEP_SELECTION) {
                        // deep selection - if it's a deep selection, ensure the popup is visible
                        if (isDeepLookup) {
                            dictPopup.show(false, () => rng.getBoundingClientRect())
                        }
                    }

                    // do stuff
                    try {
                        // do the lookup
                        const es = await Promise.all(init.dicts.map(async n => {
                            if (!settings.dict_disabled.includes(n)) {
                                try {
                                    var d = await dictionary(n)
                                } catch (ex) {
                                    throw new Error(`load ${n}: ${ex}`)
                                }
                                try {
                                    var r = await d.query(tt, true)
                                } catch (ex) {
                                    throw new Error(`query ${n}: ${ex}`)
                                }
                                return r
                            }
                        }))

                        // render the entries
                        const ee = es.flat()
                        if (ee.length) {
                            el.innerHTML = render(tt, ee)
                        } else {
                            el.innerHTML = render(tt, "No matches found.")
                        }
                    } catch (ex) {
                        el.innerHTML = render(tt, `${ex}.`)
                    }

                    if (ENABLE_DEEP_SELECTION) {
                        // deep selection - on chrome, if we drag a selection out of the popup,
                        // the up/cancel events won't get called, so we need to do this
                        el.addEventListener("pointerleave", e => {
                            dictDeepSelecting.clear()
                        })

                        // deep selection - common event handler code
                        const deepPointerAdd = e => {
                            dictDeepSelecting.add(e.pointerId)
                        }
                        const deepPointerDel = e => {
                            dictDeepSelecting.delete(e.pointerId)
                        }
                        const deepSelProc = e => {

                            // cancel the event
                            //
                            // note: this is especially required on Lithium to
                            // prevent the menus being shown when tapping the
                            // middle
                            e.stopPropagation()

                            // get the selection within the popup
                            var selection = dictPopup.shadowRoot.getSelection
                                ? dictPopup.shadowRoot.getSelection() // chrome supports shadowRoot.getSelection
                                : window.getSelection() // firefox includes shadow in window.getSelection

                            // ensure we have a selection
                            if (!selection || !selection.rangeCount) {
                                return
                            }

                            // ensure the selection is within the popup
                            if (!dictPopup.shadowRoot.contains(selection.anchorNode)) {
                                return
                            }

                            // get the selection range
                            const deepRng = selection.getRangeAt(0)
                            const wordRe = /^\w*$/

                            // collapse the range to the start
                            //
                            // note: this is required since chrome will select
                            // the existing range if an existing selection is
                            // clicked (firefox doesn't do this)
                            deepRng.collapse(true)

                            // extend the range backward until it matches word beginning
                            while ((deepRng.startOffset > 0) && deepRng.toString().match(wordRe)) {
                                deepRng.setStart(selection.anchorNode, deepRng.startOffset-1)
                            }

                            // restore the valid word match after overshooting
                            if (!deepRng.toString().match(wordRe)) {
                                deepRng.setStart(selection.anchorNode, deepRng.startOffset+1)
                            }

                            // extend the range forward until it matches word ending
                            while ((deepRng.endOffset < selection.anchorNode.length) && deepRng.toString().match(wordRe)) {
                                deepRng.setEnd(selection.anchorNode, deepRng.endOffset+1)
                            }

                            // restore the valid word match after overshooting
                            if (!deepRng.toString().match(wordRe)) {
                                deepRng.setEnd(selection.anchorNode, deepRng.endOffset-1)
                            }

                            // do a new lookup
                            if (deepRng.toString().length && deepRng.toString().match(wordRe)) {
                                // note: we use the old range on purpose for correct positioning
                                lookup(deepRng.toString(), rng, true)
                            }
                        }

                        // deep selection - add word click handlers
                        for (const x of el.querySelectorAll(".definition, .example")) {

                            // allow text selection within the element
                            //
                            // note: we want this for each individual element so
                            // clicking between them doesn't select the first word
                            // of the nearest one
                            x.style.userSelect = "text"

                            // add handlers to check if a selection is being made
                            x.addEventListener("pointerdown", deepPointerAdd, false)
                            x.addEventListener("pointerup", deepPointerAdd, false)
                            x.addEventListener("pointercancel", deepPointerDel, false)

                            // handle a selection being made by clicking a word
                            x.addEventListener("click", deepSelProc, false)
                        }
                    }

                    // we're done
                    if (ownSettle === dictReq) {
                        dictReq = undefined
                    }
                })
            }, 50)

            // deep selection - if we're not doing a deep selection, unset the settle timer so it doesn't get canceled for blank selections
            if (!ENABLE_DEEP_SELECTION || !isDeepLookup) {
                dictReq = ownSettle
            }
        }
        document.addEventListener("selectionchange", () => {
            // note: this gets called for zero-length selections (i.e., clicks) too, which is why it works for hiding it

            if (ENABLE_DEEP_SELECTION) {
                // deep selection - clear the previous hide timer
                if (dictDeepHidePending !== undefined) {
                    window.clearTimeout(dictDeepHidePending)
                    dictDeepHidePending = undefined
                }
            }

            // clear the previous settle timer
            if (dictReq !== undefined) {
                window.clearTimeout(dictReq)
                dictReq = undefined
            }

            // get the current selection if it's valid for a lookup
            let sel = document.getSelection()
            let rng, txt
            if (sel.rangeCount) {
                rng = sel.getRangeAt(0)
                txt = rng.toString()
            }
            if (rng !== undefined && txt.length < 1) {
                rng = undefined
            }
            if (rng !== undefined && txt.length > 100) {
                rng = undefined
            }
            if (rng !== undefined && (txt.match(/\s+/g) || []).length > 5) {
                rng = undefined
            }

            // if we don't have a valid selection, hide the popup
            if (rng === undefined) {
                if (ENABLE_DEEP_SELECTION) {
                    // deep selection - add a short delay to give time for a deep
                    // selection to be processed (i.e., ensure the pointer events
                    // within get handled before this)
                    dictDeepHidePending = window.setTimeout(() => {
                        if (!dictDeepSelecting.size) {
                            dictPopup.hide()
                        }
                    }, 5)
                } else {
                    dictPopup.hide()
                }
                return
            }

            // look up the word
            lookup(txt, rng)
        }, true)
    })

    return dictPopup
})()
