/**
 * Copyright 2019-2022 Patrick Gaskin
 * Licensed under the MIT license.
 * Requires WebView 4.4.4+.
 * Requires an instance of https://github.com/pgaskin/dictserver.
 */
'use strict';

var Dictionary = function () {
    var Popup = function () {
        Popup.POPUP_POSITION_TOP                = 1,
        Popup.POPUP_POSITION_BOTTOM             = 2,
        Popup.POPUP_POSITION_SELECTION_NEAR     = 3,
        Popup.POPUP_POSITION_SELECTION_OPPOSITE = 4;

        function Popup(popupPosition) {
            this._popupPosition = popupPosition || Popup.POPUP_POSITION_TOP;

            this._wrapper = css(document.createElement('div'), [
                ['all', 'initial'],
                ['display', 'block'],
                ['box-sizing', 'border-box'],
                ['position', 'fixed'],
                ['bottom', '0'],
                ['left', '16px'],
                ['right', '16px'],
                ['height', 'auto'],
                ['width', 'auto'],
                ['overflow', 'visible'],
                ['margin', '0'],
                ['padding', '0'],
                ['border', '0'],
                ['z-index', '100000'],
                ['user-select', 'none'],
                ['-webkit-user-select', 'none']
            ], true);

            if (this._wrapper.attachShadow)
                this._shadow = this._wrapper.attachShadow({mode: 'open'});
            else if (this._wrapper.webkitCreateShadowRoot) // Android 4.4+
                this._shadow = this._wrapper.webkitCreateShadowRoot();
            else if (this._wrapper.createShadowRoot)
                this._shadow = this._wrapper.createShadowRoot();
            else
                throw new Error('Popup requires shadow DOM support');

            this._popup = css(this._shadow.appendChild(document.createElement('div')), [
                ['box-sizing', 'border-box'],
                ['position', 'relative'],
                ['max-height', '220px'],
                ['minHeight', '50px'],
                ['height', '30vh'],
                ['max-width', '480px'],
                ['overflow', 'hidden'],
                ['margin', '16px auto'],
                ['border', '1px solid transparent'],
                ['border-radius', '4px']
            ]);

            this._inner = css(this._popup.appendChild(document.createElement('div')), [
                ['boxSizing', 'border-box'],
                ['position', 'absolute'],
                ['overflow-x', 'hidden'],
                ['overflow-y', 'auto'],
                ['top', '0'],
                ['bottom', '0'],
                ['left', '0'],
                ['right', '0'],
            ]);

            this.onThemeApplied(null);
        }

        Popup.prototype.onThemeApplied = function (theme) {
            css(this._popup, [
                ['border', (theme && theme.bgIsDark)
                    ? '1px solid rgba(255, 255, 255, .25)'
                    : '1px solid rgba(0, 0, 0, .25)'],
                ['box-shadow', (theme && theme.bgIsDark)
                    ? '0'
                    : '0 0 8px 0 rgba(0, 0, 0, .25)'],
                ['background-color', theme
                    ? hex(theme.backgroundColor)
                    : '#fff'],
                ['color', theme
                    ? hex(theme.textColor)
                    : '#000'],
                ['line-height', '1.25'],
                ['font-size', '14px'],
                ['font-family', 'serif'],
            ]);
        };

        Popup.prototype.reposition = function (range) {
            var rect = range.getBoundingClientRect();
            switch (this._popupPosition) {
                case Popup.POPUP_POSITION_TOP:
                    css(this._wrapper, [
                        ['top', '0'],
                        ['bottom'],
                    ], true);
                    break;
                case Popup.POPUP_POSITION_BOTTOM:
                    css(this._wrapper, [
                        ['bottom', '0'],
                        ['top'],
                    ], true);
                    break;
                case Popup.POPUP_POSITION_SELECTION_NEAR:
                    if (rect.y > document.documentElement.clientHeight / 2) {
                        css(this._wrapper, [
                            ['bottom', rect.y.toFixed() + 'px'],
                            ['top'],
                        ], true);
                    } else {
                        css(this._wrapper, [
                            ['top', (rect.y+rect.height).toFixed() + 'px'],
                            ['bottom'],
                        ], true);
                    }
                    break;
                case Popup.POPUP_POSITION_SELECTION_OPPOSITE:
                    if (rect.y > document.documentElement.clientHeight / 2) {
                        css(this._wrapper, [
                            ['top', '0'],
                            ['bottom'],
                        ], true);
                    } else {
                        css(this._wrapper, [
                            ['bottom', '0'],
                            ['top'],
                        ], true);
                    }
                    break;
                default:
                    throw new TypeError('Invalid popup position');
            }
        }

        Popup.prototype.show = function () {
            this.hide();
            this._inner.innerHTML = '';
            this._wrapper.style.setProperty('opacity', '0', 'important');
            var self = this;
            window.requestAnimationFrame(function step (ts) {
                if (!step.st) {
                    step.st = ts;
                }
                var opacity = Math.min((ts-step.st)*0.01, 1);
                self._wrapper.style.setProperty('opacity', opacity, 'important');
                if (opacity < 1) {
                    window.requestAnimationFrame(step);
                }
            });
            document.body.appendChild(this._wrapper);
            return new Popup.Item(this._inner);
        }

        Popup.prototype.hide = function () {
            if (this._wrapper.parentElement) {
                this._wrapper.parentElement.removeChild(this._wrapper);
            }
        }

        Popup.Item = (function () {
            function Item(parent) {
                this._created = [];
                this._element = document.createElement('div');
                this._title = css(this._element.appendChild(document.createElement('div')), [
                    ['margin', '8px 8px 4px 8px'],
                    ['font-weight', 'bold'],
                    ['font-size', '1.14em'],
                ]);
                this._body = css(this._element.appendChild(document.createElement('div')), [
                    ['margin', '0 8px 8px 8px'],
                ]);
                this._divider = css(this._element.appendChild(document.createElement('div')), [
                    ['margin', '8px 0'],
                    ['opacity', '0.2'],
                    ['border-bottom', '1px solid currentColor'],
                ]);
                if (parent instanceof Item) {
                    parent._element.insertAdjacentElement('afterend', this._element);
                } else {
                    parent.appendChild(this._element);
                }
            }

            Item.prototype.remove = function () {
                this.removeAll();
                if (this._element.parentElement) {
                    this._element.parentElement.removeChild(this._element);
                    this._element = null;
                }
            };

            Item.prototype.removeAll = function () {
                var item;
                while ((item = this._created.shift())) {
                    item.remove();
                }
                return this;
            };

            Item.prototype.addNew = function () {
                var last = this
                while (last._created.length) {
                    var t = false;
                    for (var i = last._created.length-1; i >= 0; i--) {
                        if (last._created[i]._element.parentElement) {
                            last = last._created[i];
                            break;
                        }
                        t = true;
                    }
                    if (t) {
                        break;
                    }
                }
                var item =  new Item(last);
                this._created.push(item);
                return item;
            };

            Item.prototype.status = function (title, body) {
                this._title.textContent = title;
                this._body.textContent = body;
                return this;
            };

            Item.prototype.word = function (word) {
                var wd = document.createDocumentFragment();

                if (word.info && word.info.trim() != '') {
                    css(wd.appendChild(document.createElement('div')), [
                        ['font-style', 'italic'],
                    ]).textContent = word.info;
                }

                var wdm = css(wd.appendChild(document.createElement('ol')), [
                    ['margin', '0'],
                    ['margin-top', '8px'],
                    ['padding-left', '2em'],
                ]);

                for (var i = 0; word.meanings && i < word.meanings.length; i++) {
                    var wdmm = css(wdm.appendChild(document.createElement('li')), [
                        ['margin-bottom', '4px'],
                    ]);
                    wdmm.appendChild(document.createElement('div')).textContent = word.meanings[i].text;

                    if (word.meanings[i].example && word.meanings[i].example.trim() != '') {
                        css(wdmm.appendChild(document.createElement('div')), [
                            ['font-style', 'italic'],
                        ]).textContent = word.meanings[i].example;
                    }
                }

                for (var i = 0; word.notes && i < word.notes.length; i++) {
                    css(wd.appendChild(document.createElement('div')), [
                        ['margin-top', '8px'],
                    ]).textContent = word.notes[i];
                }

                if (word.credit && word.credit.trim() != '') {
                    css(wd.appendChild(document.createElement('div')), [
                        ['font-style', 'italic'],
                        ['margin-top', '8px'],
                        ['font-size', '.85em'],
                    ]).textContent = word.credit;
                }

                this._title.textContent = word.word;
                this._body.innerHTML = '';
                this._body.appendChild(wd);

                return this;
            };

            return Item;
        })();

        function css(el, props, important) {
            for (var i = 0; i < props.length; i++) {
                switch (props[i].length) {
                    case 1:
                        if (props[i][0] != '')
                            el.style.removeProperty(props[i][0]);
                        break;
                    case 2:
                        if (props[i][0] != '')
                            el.style.setProperty(props[i][0], props[i][1], important ? 'important' : '');
                        break;
                    default:
                        throw new TypeError('Expected length 1 or 2 for CSS property');
                }
            }
            return el;
        }

        function hex(color) {
            var hex = color.toString(16);
            while (hex.length < 6) {
                hex = '0' + hex;
            }
            return '#' + hex;
        }

        return Popup;
    }();

    var API = function () {
        API.DEFAULT_SERVER_SET = [['https://dict.api.pgaskin.net/']];

        function API(serverSet) {
            this._ss = (serverSet && serverSet.length > 0)
                ? serverSet
                : API.DEFAULT_SERVER_SET;
        }

        API.fromLithiumSettings = function () {
            return new API(parse((typeof LithiumApp !== 'undefined' && LithiumApp.getDictionaryURL)
                ? LithiumApp.getDictionaryURL()
                : null
            ));
        }

        API.prototype.define = function (word) {
            return trySet(this._ss, api.bind(null, word)).map(transform);
        }

        // format: [server[|fallback]... ]...
        function parse(ss) {
            if (!ss) {
                return null;
            }
            ss = ss.split(' ');
            for (var i = 0; i < ss.length; i++) {
                ss[i] = ss[i].split('|');
                for (var j = 0; j < ss[i].length; j++) {
                    if (ss[i][j].length == 0) {
                        ss[i].splice(j, 1);
                        j--;
                    } else {
                        if (ss[i][j].startsWith('//'))
                            ss[i][j] = 'http:' + ss[i][j];
                        if (!ss[i][j].startsWith('http://') && !ss[i][j].startsWith('https://'))
                            ss[i][j] = 'http://' + ss[i][j];
                        if (!ss[i][j].endsWith('/'))
                            ss[i][j] += '/';
                    }
                }
                if (ss[i].length == 0) {
                    ss.splice(i, 1);
                    i--;
                }
            }
            return ss;
        }

        function transform(resp) {
            return resp.then(function (word) {
                return word
                    ? [word].concat(word.additional_words || []).concat(word.referenced_words || [])
                    : [];
            });
        };

        function trySet(sss, fn) {
            return sss.map(function (ss) {
                return ss.reduce(function (acc, s) {
                    return acc.catch(function (ex) {
                        return fn(s).catch(function(exc) {
                            return '[' + s + '] ' + exc.toString() + '; ' + ex.toString();
                        });
                    });
                }, Promise.reject('No servers left to try'), ss);
            });
        }

        function api(word, base) {
            return new Promise(function (resolve, reject) {
                var xhr = new XMLHttpRequest();
                xhr.timeout = 3000;
                xhr.onerror = function () {
                    reject('Network error');
                };
                xhr.onload = function () {
                    var obj;
                    if (xhr.status == 0) {
                        reject('Unknown error');
                        return;
                    }
                    if (xhr.status == 404) {
                        resolve(null);
                        return;
                    }
                    try {
                        obj = JSON.parse(xhr.responseText);
                    } catch (ex) {
                        if (xhr.status != 200)
                            reject('Response status ' + xhr.status.toString() + ' ' + xhr.statusText);
                        else
                            reject('Parse JSON: ' + ex.toString());
                        return;
                    }
                    if (!obj.status || obj.status != 'success') {
                        if (typeof obj.result === 'string')
                            reject('API error: ' + obj.result.toString());
                        else
                            reject('API error: ' + JSON.stringify(obj.result));
                        return;
                    }
                    resolve(obj.result);
                };
                xhr.open('GET', base + 'word/' + encodeURIComponent(word.trim()));
                xhr.send();
            });
        }

        return API;
    }();

    var Watcher = function () {
        // https://apps.timwhitlock.info/js/regex with:
        // ' ', ''', '"', Pi, Ps, Mc
        // /^(?:[...])+/
        var normIRe = /^(?:[ "'-([{«\u0903\u093e-\u0940\u0949-\u094c\u0982-\u0983\u09be-\u09c0\u09c7-\u09c8\u09cb-\u09cc\u09d7\u0a03\u0a3e-\u0a40\u0a83\u0abe-\u0ac0\u0ac9\u0acb-\u0acc\u0b02-\u0b03\u0b3e\u0b40\u0b47-\u0b48\u0b4b-\u0b4c\u0b57\u0bbe-\u0bbf\u0bc1-\u0bc2\u0bc6-\u0bc8\u0bca-\u0bcc\u0bd7\u0c01-\u0c03\u0c41-\u0c44\u0c82-\u0c83\u0cbe\u0cc0-\u0cc4\u0cc7-\u0cc8\u0cca-\u0ccb\u0cd5-\u0cd6\u0d02-\u0d03\u0d3e-\u0d40\u0d46-\u0d48\u0d4a-\u0d4c\u0d57\u0d82-\u0d83\u0dcf-\u0dd1\u0dd8-\u0ddf\u0df2-\u0df3༺༼\u0f3e-\u0f3f\u0f7f\u102b-\u102c\u1031\u1038\u103b-\u103c\u1056-\u1057\u1062-\u1064\u1067-\u106d\u1083-\u1084\u1087-\u108c\u108f᚛\u17b6\u17be-\u17c5\u17c7-\u17c8\u1923-\u1926\u1929-\u192b\u1930-\u1931\u1933-\u1938\u19b0-\u19c0\u19c8-\u19c9\u1a19-\u1a1b\u1b04\u1b35\u1b3b\u1b3d-\u1b41\u1b43-\u1b44\u1b82\u1ba1\u1ba6-\u1ba7\u1baa\u1c24-\u1c2b\u1c34-\u1c35‘‚-“„-‟‹⁅⁽₍〈❨❪❬❮❰❲❴⟅⟦⟨⟪⟬⟮⦃⦅⦇⦉⦋⦍⦏⦑⦓⦕⦗⧘⧚⧼⸂⸄⸉⸌⸜⸠⸢⸤⸦⸨〈《「『【〔〖〘〚〝\ua823-\ua824\ua827\ua880-\ua881\ua8b4-\ua8c3\ua952-\ua953\uaa2f-\uaa30\uaa33-\uaa34\uaa4d﴾︗︵︷︹︻︽︿﹁﹃﹇﹙﹛﹝（［｛｟｢]|\ud834[\udd65-\udd66\udd6d-\udd72])+/

        // https://apps.timwhitlock.info/js/regex with:
        // ' ', ''', '"', Pe, Pf, Mc
        // /(?:[...])+$/
        var normFRe = /(?:[ "')\]}»\u0903\u093e-\u0940\u0949-\u094c\u0982-\u0983\u09be-\u09c0\u09c7-\u09c8\u09cb-\u09cc\u09d7\u0a03\u0a3e-\u0a40\u0a83\u0abe-\u0ac0\u0ac9\u0acb-\u0acc\u0b02-\u0b03\u0b3e\u0b40\u0b47-\u0b48\u0b4b-\u0b4c\u0b57\u0bbe-\u0bbf\u0bc1-\u0bc2\u0bc6-\u0bc8\u0bca-\u0bcc\u0bd7\u0c01-\u0c03\u0c41-\u0c44\u0c82-\u0c83\u0cbe\u0cc0-\u0cc4\u0cc7-\u0cc8\u0cca-\u0ccb\u0cd5-\u0cd6\u0d02-\u0d03\u0d3e-\u0d40\u0d46-\u0d48\u0d4a-\u0d4c\u0d57\u0d82-\u0d83\u0dcf-\u0dd1\u0dd8-\u0ddf\u0df2-\u0df3༻༽-\u0f3f\u0f7f\u102b-\u102c\u1031\u1038\u103b-\u103c\u1056-\u1057\u1062-\u1064\u1067-\u106d\u1083-\u1084\u1087-\u108c\u108f᚜\u17b6\u17be-\u17c5\u17c7-\u17c8\u1923-\u1926\u1929-\u192b\u1930-\u1931\u1933-\u1938\u19b0-\u19c0\u19c8-\u19c9\u1a19-\u1a1b\u1b04\u1b35\u1b3b\u1b3d-\u1b41\u1b43-\u1b44\u1b82\u1ba1\u1ba6-\u1ba7\u1baa\u1c24-\u1c2b\u1c34-\u1c35’”›⁆⁾₎〉❩❫❭❯❱❳❵⟆⟧⟩⟫⟭⟯⦄⦆⦈⦊⦌⦎⦐⦒⦔⦖⦘⧙⧛⧽⸃⸅⸊⸍⸝⸡⸣⸥⸧⸩〉》」』】〕〗〙〛〞-〟\ua823-\ua824\ua827\ua880-\ua881\ua8b4-\ua8c3\ua952-\ua953\uaa2f-\uaa30\uaa33-\uaa34\uaa4d﴿︘︶︸︺︼︾﹀﹂﹄﹈﹚﹜﹞）］｝｠｣]|\ud834[\udd65-\udd66\udd6d-\udd72])+$/

        // https://apps.timwhitlock.info/js/regex with:
        // '.', '_', '-', ' ', ''', '’', Zp, Zs, Nd, Nl, No, Pc, Ll, Lm, Lo, Lt, Lu, Sc, Sk, Sm, So, Mn, Mc, Po
        // /^(?:[...]+)$/
        var checkRe = /^(?:[ -’'*-Z\\^-z|~\u00a0-ª¬®-º¼-ͷͺ-;΄-ΊΌΎ-ΡΣ-\u0487Ҋ-ԣԱ-Ֆՙ-՟ա-և։\u0591-\u05bd\u05bf-\u05c7א-תװ-״؆-؛؞-؟ء-\u065e٠-\u06dc\u06df-܍ܐ-\u074aݍ-ޱ߀-ߺ\u0901-ह\u093c-\u094dॐ-\u0954क़-ॲॻ-ॿ\u0981-\u0983অ-ঌএ-ঐও-নপ-রলশ-হ\u09bc-\u09c4\u09c7-\u09c8\u09cb-ৎ\u09d7ড়-ঢ়য়-\u09e3০-৺\u0a01-\u0a03ਅ-ਊਏ-ਐਓ-ਨਪ-ਰਲ-ਲ਼ਵ-ਸ਼ਸ-ਹ\u0a3c\u0a3e-\u0a42\u0a47-\u0a48\u0a4b-\u0a4d\u0a51ਖ਼-ੜਫ਼੦-\u0a75\u0a81-\u0a83અ-ઍએ-ઑઓ-નપ-રલ-ળવ-હ\u0abc-\u0ac5\u0ac7-\u0ac9\u0acb-\u0acdૐૠ-\u0ae3૦-૯૱\u0b01-\u0b03ଅ-ଌଏ-ଐଓ-ନପ-ରଲ-ଳଵ-ହ\u0b3c-\u0b44\u0b47-\u0b48\u0b4b-\u0b4d\u0b56-\u0b57ଡ଼-ଢ଼ୟ-\u0b63୦-ୱ\u0b82-ஃஅ-ஊஎ-ஐஒ-கங-சஜஞ-டண-தந-பம-ஹ\u0bbe-\u0bc2\u0bc6-\u0bc8\u0bca-\u0bcdௐ\u0bd7௦-௺\u0c01-\u0c03అ-ఌఎ-ఐఒ-నప-ళవ-హఽ-\u0c44\u0c46-\u0c48\u0c4a-\u0c4d\u0c55-\u0c56ౘ-ౙౠ-\u0c63౦-౯౸-౿\u0c82-\u0c83ಅ-ಌಎ-ಐಒ-ನಪ-ಳವ-ಹ\u0cbc-\u0cc4\u0cc6-\u0cc8\u0cca-\u0ccd\u0cd5-\u0cd6ೞೠ-\u0ce3೦-೯ೱ-ೲ\u0d02-\u0d03അ-ഌഎ-ഐഒ-നപ-ഹഽ-\u0d44\u0d46-\u0d48\u0d4a-\u0d4d\u0d57ൠ-\u0d63൦-൵൹-ൿ\u0d82-\u0d83අ-ඖක-නඳ-රලව-ෆ\u0dca\u0dcf-\u0dd4\u0dd6\u0dd8-\u0ddf\u0df2-෴ก-\u0e3a฿-๛ກ-ຂຄງ-ຈຊຍດ-ທນ-ຟມ-ຣລວສ-ຫອ-\u0eb9\u0ebb-ຽເ-ໄໆ\u0ec8-\u0ecd໐-໙ໜ-ໝༀ-\u0f39\u0f3e-ཇཉ-ཬ\u0f71-ྋ\u0f90-\u0f97\u0f99-\u0fbc྾-࿌࿎-࿔က-႙႞-Ⴥა-ჼᄀ-ᅙᅟ-ᆢᆨ-ᇹሀ-ቈቊ-ቍቐ-ቖቘቚ-ቝበ-ኈኊ-ኍነ-ኰኲ-ኵኸ-ኾዀዂ-ዅወ-ዖዘ-ጐጒ-ጕጘ-ፚ\u135f-፼ᎀ-᎙Ꭰ-Ᏼᐁ-ᙶ\u1680-ᚚᚠ-\u16f0ᜀ-ᜌᜎ-\u1714ᜠ-᜶ᝀ-\u1753ᝠ-ᝬᝮ-ᝰ\u1772-\u1773ក-ឳ\u17b6-\u17dd០-៩៰-៹᠀-᠅᠇-\u180e᠐-᠙ᠠ-ᡷᢀ-ᢪᤀ-ᤜ\u1920-\u192b\u1930-\u193b᥀᥄-ᥭᥰ-ᥴᦀ-ᦩ\u19b0-\u19c9᧐-᧙᧞-\u1a1b᨞-᨟\u1b00-ᭋ᭐-᭼\u1b80-\u1baaᮮ-᮹ᰀ-\u1c37᰻-᱉ᱍ-᱿ᴀ-\u1de6\u1dfe-ἕἘ-Ἕἠ-ὅὈ-Ὅὐ-ὗὙὛὝὟ-ώᾀ-ᾴᾶ-ῄῆ-ΐῖ-Ί῝-`ῲ-ῴῶ-῾\u2000-\u200a‖-‗†-‧\u2029\u202f-‸※-⁄⁇-\u205f⁰-ⁱ⁴-⁼ⁿ-₌ₐ-ₔ₠-₵\u20d0-\u20dc\u20e1\u20e5-\u20f0℀-⅏⅓-\u2188←-⌨⌫-⏧␀-␦⑀-⑊①-⚝⚠-⚼⛀-⛃✁-✄✆-✉✌-✧✩-❋❍❏-❒❖❘-❞❡-❧❶-➔➘-➯➱-➾⟀-⟄⟇-⟊⟌⟐-⟥⟰-⦂⦙-⧗⧜-⧻⧾-⭌⭐-⭔Ⰰ-Ⱞⰰ-ⱞⱠ-Ɐⱱ-ⱽⲀ-⳪⳹-ⴥⴰ-ⵥⵯⶀ-ⶖⶠ-ⶦⶨ-ⶮⶰ-ⶶⶸ-ⶾⷀ-ⷆⷈ-ⷎⷐ-ⷖⷘ-ⷞ\u2de0-⸁⸆-⸈⸋⸎-⸖⸘-⸙⸛⸞-⸟⸪-⸰⺀-⺙⺛-⻳⼀-⿕⿰-⿻\u3000-\u3007〒-〓〠-\u302f〱-〿ぁ-ゖ\u3099-ゟァ-ヿㄅ-ㄭㄱ-ㆎ㆐-ㆷ㇀-㇣ㇰ-㈞㈠-㉃㉐-㋾㌀-䶵䷀-鿃ꀀ-ꒌ꒐-꓆ꔀ-ꘫꙀ-ꙟꙢ-\ua66f꙳\ua67c-ꚗ꜀-ꞌꟻ-꠫ꡀ-꡷\ua880-\ua8c4꣎-꣙꤀-\ua953꥟ꨀ-\uaa36ꩀ-\uaa4d꩐-꩙꩜-꩟가-힣豈-鶴侮-頻並-龎ﬀ-ﬆﬓ-ﬗיִ-זּטּ-לּמּנּ-סּףּ-פּצּ-ﮱﯓ-ﴽﵐ-ﶏﶒ-ﷇﷰ-﷽\ufe00-︖︙\ufe20-\ufe26︰︳-︴﹅-﹆﹉-﹒﹔-﹗﹟-﹢﹤-﹦﹨-﹫ﹰ-ﹴﹶ-ﻼ！-＇＊-，．-Ｚ＼＾-ｚ｜～｡､-ﾾￂ-ￇￊ-ￏￒ-ￗￚ-ￜ￠-￦￨-￮￼-�]|[\ud840-\ud868][\udc00-\udfff]|\ud800[\udc00-\udc0b\udc0d-\udc26\udc28-\udc3a\udc3c-\udc3d\udc3f-\udc4d\udc50-\udc5d\udc80-\udcfa\udd00-\udd02\udd07-\udd33\udd37-\udd8a\udd90-\udd9b\uddd0-\uddfd\ude80-\ude9c\udea0-\uded0\udf00-\udf1e\udf20-\udf23\udf30-\udf4a\udf80-\udf9d\udf9f-\udfc3\udfc8-\udfd5]|\ud801[\udc00-\udc9d\udca0-\udca9]|\ud802[\udc00-\udc05\udc08\udc0a-\udc35\udc37-\udc38\udc3c\udc3f\udd00-\udd19\udd1f-\udd39\udd3f\ude00-\ude03\ude05-\ude06\ude0c-\ude13\ude15-\ude17\ude19-\ude33\ude38-\ude3a\ude3f-\ude47\ude50-\ude58]|\ud808[\udc00-\udf6e]|\ud809[\udc00-\udc62\udc70-\udc73]|\ud834[\udc00-\udcf5\udd00-\udd26\udd29-\udd72\udd7b-\udddd\ude00-\ude45\udf00-\udf56\udf60-\udf71]|\ud835[\udc00-\udc54\udc56-\udc9c\udc9e-\udc9f\udca2\udca5-\udca6\udca9-\udcac\udcae-\udcb9\udcbb\udcbd-\udcc3\udcc5-\udd05\udd07-\udd0a\udd0d-\udd14\udd16-\udd1c\udd1e-\udd39\udd3b-\udd3e\udd40-\udd44\udd46\udd4a-\udd50\udd52-\udea5\udea8-\udfcb\udfce-\udfff]|\ud83c[\udc00-\udc2b\udc30-\udc93]|\ud869[\udc00-\uded6]|\ud87e[\udc00-\ude1d]|\udb40[\udd00-\uddef])+$/

        // https://apps.timwhitlock.info/js/regex with:
        // ' ', Zp, Zs
        // /(?:[...])+/
        var spaceRe = /(?:[ \u00a0\u1680\u180e\u2000-\u200a\u2029\u202f\u205f\u3000])+/

        function Watcher(api, popup, settleTimeout) {
            this.api = api;
            this.popup = popup;
            this.settleTimeout = settleTimeout || 700;
            this._settleTimer = null;
            this._pendingItem = null;
            this._range = null;

            document.addEventListener('selectionchange', function () {
                this.onSelectionChange();
            }.bind(this), true);

            document.addEventListener('scroll', function () {
                if (this._range) {
                    this.popup.reposition(this._range);
                }
            }.bind(this), true);

            window.addEventListener('resize', function () {
                if (this._range) {
                    this.popup.reposition(this._range);
                }
            }.bind(this), true);
        }

        Watcher.prototype.onSelectionChange = function () {
            // clear the existing settle timer, if any
            if (this._settleTimer) {
                window.clearTimeout(this._settleTimer);
                this._settleTimer = null;
            }

            // process the selection
            var sel = document.getSelection();
            this._range = sel.rangeCount != 0
                ? sel.getRangeAt(0)
                : null;
            var normalized = this._range.toString().trim().replace(normIRe, '').replace(normFRe, '')
            var eligible = this._range
                && normalized.length != 0
                && normalized.length < 50
                && normalized.split(spaceRe).length <= 4
                && normalized.match(checkRe);

            // hide the popup if the selection isn't eligible
            if (!eligible) {
                if (this._pendingItem) {
                    this._pendingItem.removeAll();
                    this._pendingItem = null;
                }
                this.popup.hide();
                return;
            }

            // show/update the loading popup
            this.popup.reposition(this._range);
            if (this._pendingItem) {
                this._pendingItem.removeAll();
            } else {
                this._pendingItem = this.popup.show();
            }
            this._pendingItem.status(normalized, 'Loading...');

            // after the selection settles, make the request
            var self = this;
            this._settleTimer = window.setTimeout(function () {
                Promise.all(self.api.define(normalized).map(function (rsp) {
                    return rsp.catch(function (ex) {
                        return ex.toString();
                    });
                })).then(function (ss) {
                    if (self._pendingItem) {
                        var x = self._pendingItem.removeAll();
                        for (var i = 0; i < ss.length; i++) {
                            x = i ? x.addNew() : x;
                            if (!Array.isArray(ss[i])) {
                                x.status(normalized, ss[i]);
                            } else if (ss[i].length == 0) {
                                x.status(normalized, 'No match for word.'); // TODO: maybe show the dictionary server if more than one?
                            } else {
                                for (var j = 0; j < ss[i].length; j++) {
                                    (j ? x.addNew() : x).word(ss[i][j]);
                                }
                            }
                        }
                    }
                }).catch(function (ex) {
                    if (self._pendingItem) {
                        self._pendingItem.removeAll().status(normalized, ex.toString());
                    }
                });
            }, self.settleTimeout);
        }

        return Watcher;
    }();

    var popup = new Popup(Popup.POPUP_POSITION_SELECTION_OPPOSITE);
    var api = API.fromLithiumSettings();
    var watcher = new Watcher(api, popup);

    hookThemeApplied(popup.onThemeApplied.bind(popup));

    function hookThemeApplied(hook) {
        if (typeof LithiumAnnotations !== 'undefined') {
            var orig = LithiumAnnotations.onThemeApplied;
            LithiumAnnotations.onThemeApplied = function (theme) {
                try {
                    hook(theme);
                } catch (ex) {
                    console.error(ex);
                }
                return orig(theme);
            };
        }
    }

    return watcher;
}();
