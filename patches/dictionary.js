/**
 * Copyright 2019-2023 Patrick Gaskin
 * Requires a relatively recent version of the Chromium WebView.
 */
"use strict";

var Dictionary = (function() {

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
     * Raw HTML.
     */
    function rawHTML(s) {
        return Object.defineProperties(new String(s), {
            [isHtml]: {
                value: true,
            },
        })
    }

    /**
     * Material Icons Rounded
     */
    const matIconSearch = rawHTML(`<svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 0 24 24" width="24px" fill="#e8eaed"><path d="M0 0h24v24H0z" fill="none"/><path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/></svg>`)
    const matIconClose = rawHTML(`<svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 0 24 24" width="24px" fill="#e8eaed"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M18.3 5.71c-.39-.39-1.02-.39-1.41 0L12 10.59 7.11 5.7c-.39-.39-1.02-.39-1.41 0-.39.39-.39 1.02 0 1.41L10.59 12 5.7 16.89c-.39.39-.39 1.02 0 1.41.39.39 1.02.39 1.41 0L12 13.41l4.89 4.89c.39.39 1.02.39 1.41 0 .39-.39.39-1.02 0-1.41L13.41 12l4.89-4.89c.38-.38.38-1.02 0-1.4z"/></svg>`)

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
        #expand
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
                #wrapper.expand {
                    display: flex;
                    flex-direction: column;
                }
                #wrapper.expand #popup {
                    flex: 1;
                    max-height: none;
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
                this.#popup.style.setProperty("--popup-background", bg !== undefined ? bg : dark ? "#333" : "#fff")
                this.#popup.style.setProperty("--popup-border-color", dark ? "rgba(255, 255, 255, .25)" : "rgba(0, 0, 0, .25)")
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
         * Expands the popup to fill the screen if it's visible. Reset upon the
         * next call to hide.
         */
        expand() {
            this.#expand = true
            this.move()
        }

        /**
         * Hides the popup if it is visible.
         */
        hide() {
            if (this.visible) {
                this.#root.remove()
                this.#pos = false
                this.#posRect = undefined
                this.#expand = false
            }
        }

        /**
         * Updates the popup's position if it is visible.
         */
        move() {
            let top, bot
            if (this.#expand) {
                top = 0
                bot = 0
            } else if (this.#posRect === undefined) {
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
            if (this.#expand) {
                this.#wrapper.classList.add("expand")
            } else {
                this.#wrapper.classList.remove("expand")
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

        /**
         * Whether the popup is currently expanded.
         */
        get expanded() {
            return !!this.#expand
        }
    }

    class SelectionController {
        #selectionClearPending // timer

        #deepSelectionRoot // element
        #deepSelectionRootShadow // shadow root
        #deepSelectionPointers = new Set() // of pointer IDs which are down inside the root
        #deepSelectionWasLastSelection = false
        #_handleDeepPointerAdd // bound event handler
        #_handleDeepPointerDel // bound event handler
        #_handleDeepClick // bound event handler

        constructor() {
            document.addEventListener("selectionchange", this.#handleSelectionChange.bind(this), false)
        }

        /** 
         * Callback to check whether a range is valid for initial selections.
         * Takes a Range, returns a boolean.
         */
        rangeValidate

        /**
         * Callback for when a valid range is selected. A selection is when the
         * selected text changes (and is not in the deep selection root). Takes
         * a Range (which is cloned).
         */
        rangeSelected

        /**
         * Callback for when a range is deep-selected. The range will have a
         * length of zero. Takes a Range (which is cloned) and the selection
         * anchorNode.
         */
        rangeSelectedDeep

        /**
         * Called when no range is selected anymore.
         */
        rangeCleared

        /**
         * Sets the root element used for deep selection. If the element is
         * within a Shadow DOM, the shadow root must also be passed.
         */
        changeDeepSelectionRoot(el, shadowRoot = undefined) {
            if (this.#deepSelectionRoot) {
                this.#deepSelectionRoot.removeEventListener("pointerdown", this.#_handleDeepPointerAdd, false)
                this.#deepSelectionRoot.removeEventListener("pointerup", this.#_handleDeepPointerDel, false)
                this.#deepSelectionRoot.removeEventListener("pointercancel", this.#_handleDeepPointerDel, false)
                this.#deepSelectionRoot.removeEventListener("pointerleave", this.#_handleDeepPointerDel, false)
                this.#deepSelectionRoot.removeEventListener("pointerout", this.#_handleDeepPointerDel, false)
                this.#deepSelectionRoot.removeEventListener("click", this.#_handleDeepClick, false)
                this.#deepSelectionPointers.clear()
                this.#deepSelectionRootShadow = undefined
                this.#deepSelectionRoot = undefined
            }
            this.#deepSelectionRoot = el
            this.#deepSelectionRootShadow = shadowRoot
            if (this.#deepSelectionRoot) {
                this.#_handleDeepPointerAdd = this.#handleDeepPointerAdd.bind(this)
                this.#_handleDeepPointerDel = this.#handleDeepPointerDel.bind(this)
                this.#_handleDeepClick = this.#handleDeepClick.bind(this)
                this.#deepSelectionRoot.addEventListener("pointerdown", this.#_handleDeepPointerAdd, false)
                this.#deepSelectionRoot.addEventListener("pointerup", this.#_handleDeepPointerDel, false)
                this.#deepSelectionRoot.addEventListener("pointercancel", this.#_handleDeepPointerDel, false)
                this.#deepSelectionRoot.addEventListener("pointerleave", this.#_handleDeepPointerDel, false)
                this.#deepSelectionRoot.addEventListener("pointerout", this.#_handleDeepPointerDel, false)
                this.#deepSelectionRoot.addEventListener("click", this.#_handleDeepClick, false)
                // yes, we don't want handleDeepPointerAdd for pointerenter
            }
        }

        /**
         * Clears the current selection. Does not call rangeCleared.
         */
        clear() {
            const sel = document.getSelection()
            if (sel?.rangeCount) {
                sel?.empty?.()
                sel?.removeAllRanges?.()
            }
        }

        #handleSelectionChange(event) {
            // note: this gets called for zero-length selections (i.e., clicks)
            // too, which is why it works for hiding it

            // clear the previous clear timer
            if (this.#selectionClearPending !== undefined) {
                window.clearTimeout(this.#selectionClearPending)
                this.#selectionClearPending = undefined
            }

            // get the selection
            const sel = document.getSelection()

            // ensure the selection didn't start within the deep selection root
            if (sel && this.#deepSelectionRoot?.contains(sel.anchorNode)) {
                return
            }

            let rng
            if (sel?.rangeCount) {
                // get the range
                rng = sel.getRangeAt(0)
            }
            if (rng) {
                // clone the range
                rng = rng.cloneRange()
            }
            if (rng) {
                // check if it's empty
                if (!rng.toString().length) {
                    rng = undefined
                }
            }
            if (rng) {
                // validate the range
                if (this.rangeValidate && !this.rangeValidate(rng)) {
                    rng = undefined
                }
            }
            if (rng) {
                // call the callback for a selection
                this.rangeSelected?.(rng, false)
            } else {
                if (this.#deepSelectionRoot) {
                    // add a short delay to give time for a deep selection to be
                    // processed (i.e., don't mark the selection as cleared
                    // while a deep selection might be in progress)
                    this.#selectionClearPending = window.setTimeout(() => {
                        if (!this.#deepSelectionPointers.size) {

                            // save the deep selection flag
                            const deepSelectionWasLastSelection = this.#deepSelectionWasLastSelection

                            // reset the deep selection flag
                            this.#deepSelectionWasLastSelection = false

                            // call the callback for a cleared selection
                            if (!deepSelectionWasLastSelection) {
                                this.rangeCleared?.()
                            }
                        }
                    }, 5)
                } else {

                    // save the deep selection flag
                    const deepSelectionWasLastSelection = this.#deepSelectionWasLastSelection

                    // reset the deep selection flag
                    this.#deepSelectionWasLastSelection = false

                    // call the callback for a cleared selection
                    if (!deepSelectionWasLastSelection) {
                        this.rangeCleared?.()
                    }
                }
            }
        }

        #handleDeepPointerAdd(event) {
            this.#deepSelectionPointers.add(event.pointerId)
        }

        #handleDeepPointerDel(event) {
            this.#deepSelectionPointers.delete(event.pointerId)
        }

        #handleDeepClick(event) {
            // clicking sets the selection for a non-user-select-none element
            //
            // note: this is usually a zero-length selection, but on chrome (as
            // of 122), clicking on an existing selection will use that

            // stop propagating the event
            //
            // note: this is especially important on Lithium to prevent the
            // menus being shown when tapping the middle
            event.stopPropagation()

            // get the selection
            const sel = this.#deepSelectionRootShadow?.getSelection
                ? this.#deepSelectionRootShadow.getSelection() // non-standard, only supported on chrome
                : window.getSelection();
            
            // ensure we have a selection
            if (!sel?.rangeCount) {
                return
            }

            // ensure the selection started within the deep selection root
            if (!this.#deepSelectionRoot.contains(sel.anchorNode)) {
                return
            }

            // get the range
            let rng = sel.getRangeAt(0)

            // normalize it to the start
            rng.collapse(true)

            // clone the range
            rng = rng.cloneRange()

            // set the deep selection flag to inhibit the next selection clear
            // event (i.e., the one caused by de-selecting the first selection
            // and doing the deep selection)
            this.#deepSelectionWasLastSelection = true

            // call the callback
            this.rangeSelectedDeep?.(rng, sel.anchorNode)
        }
    }

    const createAutocompleteSearch = (el, normalize, autocomplete, search) => {
        const input = el.querySelector("input")
        const ul = el.querySelector("ul")

        let last
        const update = query => {
            if (normalize) {
                query = normalize(query)
            }
            if (query === last) {
                return
            }
            const ws = Array.from(new Set(autocomplete ? autocomplete(query) : [])).sort((a, b) => {
                return a.length - b.length || b.localeCompare(a)
            })
            for (let i = 0; i < ws.length; i++) {
                const li = i < ul.children.length
                    ? ul.children[i] // reuse where possible
                    : ul.appendChild(document.createElement("li"))
                li.tabindex = -1
                li.dataset.term = ws[i]
                li.textContent = ws[i] // TODO: show source dicts? ws.get(sws[i])
            }
            for (let i = ul.children.length-1; i >= ws.length; i--) {
                ul.children[i].remove()
            }
            if (ul.children.length) {
                ul.children[0].scrollIntoView?.({behavior: "instant", block: "nearest", inline: "nearest"})
            }
        }

        // handle the enter key
        input.addEventListener("keypress", event => {
            if (event.keyCode == 13) {
                event.preventDefault()
                event.stopPropagation()
                if (search) {
                    search(acInput.value)
                }
            }
        }, true)

        // handle an autocomplete entry
        ul.addEventListener("click", event => {
            event.preventDefault()
            event.stopPropagation()
            if (event.target?.dataset?.term && search) {
                search(event.target.dataset.term)
            }
        }, true)

        // handle input
        // note: I would use the input event, but this is more reliable across devices
        const interval = window.setInterval(() => {
            if (!input.isConnected) {
                window.clearInterval(interval)
                return
            }
            update(input.value)
        }, 250)
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
        nav.toolbar {
            display: flex;
            white-space: nowrap;
            position: fixed;
            top: 0;
            right: 0;
            z-index: 1000;
        }
        nav.toolbar > button {
            appearance: none;
            border: 0;
            padding: 0;
            margin: 0;
            font: inherit;
            color: inherit;
            background: none;
            border-radius: 0;
            outline: 0;
            line-height: 1;
        }
        nav.toolbar > button {
            display: flex;
            justify-content: center;
            align-items: center;
            flex: 0 0 auto;
            padding: 8px;
        }
        nav.toolbar > button:hover {
            background: rgba(0, 0, 0, 0.1);
        }
        nav.toolbar > button:active {
            background: rgba(0, 0, 0, 0.15);
        }
        nav.toolbar > button > svg {
            flex: 0 0 auto;
            fill: currentColor;
            height: 16px;
            width: 16px;
        }
        nav.toolbar > button.close {
            display: none;
        }
        aside.lookup {
            display: none;
        }
        aside.lookup > input {
            appearance: none;
            border: 0;
            padding: 0;
            margin: 0;
            font: inherit;
            color: inherit;
            background: none;
            border-radius: 0;
            outline: 0;
            line-height: 1;
        }
        aside.lookup > input {
            display: block;
            width: 100%;
            position: sticky;
            top: 0;
            left: 0;
            right: 0;
            padding: 8px;
            font-weight: inherit;
            background: var(--popup-background);
            border-bottom: 1px solid var(--popup-border-color);
        }
        aside.lookup > ul {
            list-style-type: none;
            padding: 0;
            margin: 0;
            text-indent: 0;
        }
        aside.lookup > ul > li {
            display: block;
            padding: 0 8px;
            line-height: 2.25;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        aside.lookup > ul > li:hover {
            background: rgba(0, 0, 0, 0.1);
        }
        aside.lookup > ul > li:active {
            background: rgba(0, 0, 0, 0.15);
        }
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
        section > ol.meaning-group-definitions > li > span.definition,
        section > ol.meaning-group-definitions > li > div.example {
            /*
                for deep selection - we want this for each individual element so
                clicking between them doesn't select the first word of the
                nearest one
            */
            user-select: text;
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
        let dictSettle // timer
        let dictSem // promise
        let dictClientRect // function -> rect

        const controller = new SelectionController()

        const lookup = (txt, deep, query) => {

            // set the initial popup
            const tt = Dictionary.normalize(txt)
            const pw = dictPopup.replace(html`
                <nav class="toolbar">
                    <button class="lookup">${matIconSearch}</button>
                    <button class="close">${matIconClose}</button>
                </nav>
                <aside class="lookup">
                    <input type="text" autocomplete="off" placeholder="Search dictionary..."/>
                    <ul></ul>
                </aside>
            `)

            // allow expanding the popup by clicking on a header
            pw.addEventListener("click", event => {
                if (!event.target || (event.target.tagName.toLowerCase() != "HEADER" && Array.from(event.target.parentElement.querySelectorAll("header *")).indexOf(event.target) == -1))
                    return
                event.preventDefault()
                event.stopPropagation()
                dictPopup.expand()
                controller.clear()
                pw.querySelector("button.close").style.display = "block"
            }, true)

            // handle the lookup button
            pw.querySelector("button.lookup").addEventListener("click", event => {
                event.preventDefault()
                event.stopPropagation()
                pw.querySelector("aside.lookup").style.display = "block"
                pw.querySelector("button.lookup").style.display = "none"
                pw.querySelector("button.close").style.display = "block"
                pw.querySelector("main").style.display = "none"
                dictPopup.expand()
                controller.clear()
                if (query?.length) {
                    pw.querySelector("aside.lookup > input").value = query
                } else {
                    pw.querySelector("aside.lookup > input").focus()
                }
            }, true)

            // handle the close button
            pw.querySelector("button.close").addEventListener("click", event => {
                event.preventDefault()
                event.stopPropagation()
                dictPopup.hide()
            }, true)

            // don't let the autocomplete trigger the expand/lookup or gestures
            pw.querySelector("aside.lookup").addEventListener("click", event => {
                event.preventDefault()
                event.stopPropagation()
            }, false)

            // initialize the autocomplete
            const acDicts = []
            createAutocompleteSearch(pw.querySelector("aside.lookup"),
                query => Dictionary.normalize(query),
                query => {
                    const ws = new Map()
                    if (query.length >= 2) {
                        for (const x of acDicts) {
                            for (const w of x.d.autocomplete(query, 15, true)) {
                                let src = ws.get(w)
                                if (src === undefined) {
                                    src = new Set()
                                    ws.set(w, src)
                                }
                                src.add(x.n)
                            }
                        }
                    }
                    return ws.keys() // TODO: include the dict source?
                },
                term => {
                    lookup(term, true, pw.querySelector("aside.lookup > input").value.trim())
                },
            )

            // render the contents
            const el = document.createElement("main")
            el.innerHTML = render(tt, "Loading.")
            pw.appendChild(el)

            // show the popup
            if (dictPopup.show(false, dictClientRect) || deep) {

                // if we're not modifying an existing selection (or it's a deep selection), discard the old semaphore
                dictSem = Promise.resolve()
            }
            if (dictPopup.expanded) {
                pw.querySelector("button.close").style.display = "block"
            }

            // wait for the selection to settle
            const ownSettle = dictSettle = window.setTimeout(() => {

                // wait for the previous lookup to finish, then continue ours
                dictSem = dictSem.finally(async () => {

                    // check if we've been replaced or canceled
                    if (ownSettle === dictSettle) {
                        dictSettle = undefined
                    } else {
                        return
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
                                acDicts.push({n, d})
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

                    // allow text selection within the element for deep selection
                    //
                    // note: we want this for each individual element so
                    // clicking between them doesn't select the first word
                    // of the nearest one
                    controller.changeDeepSelectionRoot(el, dictPopup.shadowRoot)
                })
            }, deep ? 0 : 50)
        }

        controller.rangeValidate = rng => {
            const txt = rng.toString()
            if (txt.length < 1) {
                return false
            }
            if (txt.length > 100) {
                return false
            }
            if ((txt.match(/\s+/g) || []).length > 5) {
                return false
            }
            return true
        }
        controller.rangeCleared = () => {
            // clear the settle timer
            if (dictSettle !== undefined) {
                window.clearTimeout(dictSettle)
                dictSettle = undefined
            }

            // hide the popup
            if (!dictPopup.expanded) {
                dictPopup.hide()
            }
        }
        controller.rangeSelected = rng => {
            // clear the settle timer
            if (dictSettle !== undefined) {
                window.clearTimeout(dictSettle)
                dictSettle = undefined
            }

            // save the range location
            dictClientRect = () => rng.getBoundingClientRect()

            // do the lookup
            if (!dictPopup.expanded) {
                lookup(rng.toString(), false)
            }
        }
        controller.rangeSelectedDeep = (rng, anchorNode) => {
            const re = /^\w*$/

            // extend the range backward until it matches word beginning
            while ((rng.startOffset > 0) && rng.toString().match(re)) {
                rng.setStart(anchorNode, rng.startOffset-1)
            }

            // restore the valid word match after overshooting
            if (!rng.toString().match(re)) {
                rng.setStart(anchorNode, rng.startOffset+1)
            }

            // extend the range forward until it matches word ending
            while ((rng.endOffset < anchorNode.length) && rng.toString().match(re)) {
                rng.setEnd(anchorNode, rng.endOffset+1)
            }

            // restore the valid word match after overshooting
            if (!rng.toString().match(re)) {
                rng.setEnd(anchorNode, rng.endOffset-1)
            }

            // ignore it if it there's nothing left
            if (!rng.toString().length || !rng.toString().match(re)) {
                return
            }

            // do the lookup
            lookup(rng.toString(), true)
        }
    })

    return dictPopup
})()
