/**
 * Copyright 2019-2023 Patrick Gaskin
 * Requires a relatively recent version of the Chromium WebView.
 */
"use strict";

globalThis.Dictionary = (function() {
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
            } else if (literals.raw[i]?.endsWith(" style=") && typeof sub === "object") {
                sub = `"${escape(css(sub))}"`
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
     * Converts obj into a CSS string, normalizing property names and removing
     * false/undefined.
     */
    function css(obj) {
        const props = Object.entries(obj).map(([k, v]) => [
            k.replace(/[A-Z]/g, c => `-${c.toLowerCase()}`),
            (v === false || v === null) ? undefined : v
        ])
        const css = props
            .filter(([k, v]) => typeof v !== "undefined")
            .map(([k, v]) => `${k}:${v}`)
            .join(";")
        return css
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
     * theme update is sent on DOMContentLoaded.
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
        #wrapper
        #shadow
        #popup
        #inner
        #content
        #pos
        #posRect
        constructor() {
            this.#wrapper = document.createElement("x-popup")
            this.#wrapper.style.setProperty("display", "block", "important")
            this.#wrapper.style.setProperty("margin", "0", "important")
            this.#wrapper.style.setProperty("padding", "0", "important")
            this.#wrapper.style.setProperty("position", "fixed", "important")
            this.#wrapper.style.setProperty("left", "16px", "important")
            this.#wrapper.style.setProperty("right", "16px", "important")
            this.#wrapper.style.setProperty("height", "auto", "important")
            this.#wrapper.style.setProperty("width", "auto", "important")
            this.#wrapper.style.setProperty("overflow", "visible", "important")
            this.#wrapper.style.setProperty("user-select", "none", "important")
            this.#wrapper.style.setProperty("z-index", "9999999", "important")
            this.#wrapper.style.setProperty("transform", "none", "important")
            this.#wrapper.style.setProperty("filter", "none", "important")
            this.#wrapper.style.setProperty("animation", "none", "important")
            this.#wrapper.style.setProperty("transition", "opacity .1s ease-in", "important")

            this.#shadow = this.#wrapper.attachShadow({mode: "open"})

            this.#popup = this.#shadow.appendChild(document.createElement("div"))
            this.#popup.style.setProperty("box-sizing", "border-box")
            this.#popup.style.setProperty("position", "relative")
            this.#popup.style.setProperty("max-height", "220px")
            this.#popup.style.setProperty("min-height", "50px")
            this.#popup.style.setProperty("height", "30vh")
            this.#popup.style.setProperty("max-width", "480px")
            this.#popup.style.setProperty("width", "100%")
            this.#popup.style.setProperty("overflow", "hidden")
            this.#popup.style.setProperty("margin", "16px auto")
            this.#popup.style.setProperty("border", "1px solid transparent")
            this.#popup.style.setProperty("border-radius", "4px")

            const applyTheme = (dark = false, bg = undefined, fg = undefined) => {
                this.#popup.style.setProperty("border-color", dark ? "rgba(255, 255, 255, .25)" : "rgba(0, 0, 0, .25)")
                this.#popup.style.setProperty("box-shadow", dark ? "none" : "0 0 8px 0 rgba(0, 0, 0, .25)")
                this.#popup.style.setProperty("background-color", bg !== undefined ? bg : dark ? "#333" : "#fff")
                this.#popup.style.setProperty("color", fg !== undefined ? fg : dark ? "#eee" : "#000")
            }
            if (!hookLithiumTheme(theme => {
                applyTheme(theme.bgIsDark, hexColor(theme.backgroundColor), hexColor(theme.textColor))
            })) {
                const matcher = window.matchMedia('(prefers-color-scheme: dark)')
                if (matcher) {
                    applyTheme(matcher.matches)
                    matcher.addEventListener("change", ev => {
                        applyTheme(ev.matches)
                    })
                } else {
                    applyTheme()
                }
            }

            this.#popup.style.setProperty("line-height", "1.25")
            this.#popup.style.setProperty("font-size", "14px")
            this.#popup.style.setProperty("font-family", "serif")

            this.#inner = this.#popup.appendChild(document.createElement("div"))
            this.#inner.style.setProperty("box-sizing", "border-box")
            this.#inner.style.setProperty("position", "absolute")
            this.#inner.style.setProperty("overflow-x", "hidden")
            this.#inner.style.setProperty("overflow-y", "auto")
            this.#inner.style.setProperty("inset", "0")

            this.#content = this.#inner.appendChild(document.createElement("div"))

            this.#shadow.appendChild(document.createElement("style")).textContent = `
                ::-webkit-scrollbar {
                    display: none;
                }
            `

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
                document.body.appendChild(this.#wrapper);
                window.setTimeout(() => this.#wrapper.style.setProperty("opacity", "1", "important"), 0)
                return true
            }
            if (this.#wrapper.parentElement.lastElementChild != this.#wrapper) {
                this.#wrapper.parentElement.appendChild(this.#wrapper)
            }
            return false
        }

        /**
         * Hides the popup if it is visible.
         */
        hide() {
            if (this.visible) {
                this.#wrapper.remove()
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
            return this.#wrapper.parentElement
        }
    }

    const settings = {
        dict_disabled: 'LithiumApp' in globalThis ? globalThis.LithiumApp.getDictDisabled().split(" ") : [],
        dict_show_examples: 'LithiumApp' in globalThis ? globalThis.LithiumApp.getDictShowExamples() : true,
        dict_show_info: 'LithiumApp' in globalThis ? globalThis.LithiumApp.getDictShowInfo() : true,
    }

    const define = (() => {
        const script = document.querySelector("script[data-dictionaries]")
        const dicts = script.dataset.dictionaries.split(" ").filter(n => !settings.dict_disabled.includes(n))
        const url = new URL(script.getAttribute("src"), window.location.href)
        const urls = dicts.map(n => [n, new URL(n, url).href])
        return term => Promise.all(urls.map(async ([n, u]) => {
            try {
                var dict = await globalThis.dict(u)
            } catch (ex) {
                throw new Error(`load ${n}: ${ex}`)
            }
            try {
                var res = await dict.query(term)
            } catch (ex) {
                throw new Error(`query ${n}: ${ex}`)
            }
            return res
        })).then(r => r.filter(r => r).flatMap(r => r.entries))
    })()

    const render = (t, x) => !Array.isArray(x) ? html`
        <div style=${{padding: "8px"}}>
            <div style=${{fontSize: "1.14em", marginBottom: "4px"}}>
                <span style=${{fontWeight: "bold"}}>${t}</span>${"\u00a0"}
            </div>
            <div>
                ${x}
            </div>
        </div>
    ` : x.map((x, i) => html`
        <div style=${{padding: "8px", borderTop: !!i && "1px solid rgba(128,128,128,0.4)"}}>
            <div style=${{fontSize: "1.14em", marginBottom: "4px"}}>
                <span style=${{fontWeight: "bold"}}>${x.name}</span>
                ${!!x.pronunciation && html`
                    <span style=${{opacity: .75}}> ${"\u00b7\u00a0"}${x.pronunciation}</span>
                `}
                ${"\u00a0"}
            </div>
            ${x.meaningGroups.map(x => html`
                ${!!x.info.length && html`
                    <div style=${{fontStyle: "italic", marginBottom: "4px"}}>${x.info.join(" \u2014 ")}</div>
                `}
                ${!!x.meanings.length && html`
                    <ol style=${{margin: 0, marginTop: "8px", marginBottom: "16px", paddingLeft: "2em"}}>
                        ${x.meanings.map(x => html`
                            <li style=${{marginBottom: "4px"}}>
                                ${x.tags.map(x => html`
                                    <span style=${{
                                        display: "inline-block",
                                        fontSize: ".85em",
                                        verticalAlign: "baseline",
                                        padding: ".1em .25em",
                                        backgroundColor: "rgba(128,128,128,0.2)",
                                    }}>${x}</span>${"\u00a0"}
                                `)}
                                ${!!x.text.length && html`
                                    <span>${x.text}</span>
                                `}
                                ${settings.dict_show_examples && x.examples.map(x => html`
                                    <div style=${{fontStyle: "italic", marginTop: "4px"}}>${x}</div>
                                `)}
                            </li>
                        `)}
                    </ol>
                `}
            `)}
            ${settings.dict_show_info && !!x.info.length && html`
                <div style=${{opacity: .75, marginTop: "8px"}}>${x.info}</div>
            `}
            ${!!x.source.length && html`
                <div style=${{fontStyle: "italic", fontSize: ".85em", marginTop: "8px"}}>${x.source}</div>
            `}
        </div>
    `)

    let dictReq
    let dictSem
    const dictPopup = new Popup()

    document.addEventListener("selectionchange", () => {
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
            dictPopup.hide()
            return
        }

        // set the initial popup contents
        const tt = globalThis.dict.Dictionary.normalize(txt)
        const el = dictPopup.replace(render(tt, "Loading."))

        // show the popup
        if (dictPopup.show(false, () => rng.getBoundingClientRect())) {

            // if we're not modifying an existing selection, discard the old semaphore
            dictSem = Promise.resolve()
        }

        // wait for the selection to settle
        const ownSettle = dictReq = window.setTimeout(() => {

            // wait for the previous lookup to finish, then continue ours
            dictSem = dictSem.catch(() => {}).then(async () => {

                // check if we've been replaced or canceled
                if (ownSettle !== dictReq) {
                    return
                }

                // do the lookup
                try {
                    const es = await define(txt)
                    if (es.length) {
                        el.innerHTML = render(tt, es)
                    } else {
                        el.innerHTML = render(tt, "No matches found.")
                    }
                } catch (ex) {
                    el.innerHTML = render(tt, `Error: ${ex}.`)
                }

                // we're done
                if (ownSettle === dictReq) {
                    dictReq = undefined
                }
            })
        }, 50)
    }, true)

    return dictPopup
})()
