//go:build ignore

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"os"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

var (
	LithiumAPK = "Lithium_0.24.5.apk"
	LithiumURL = "https://www.apkmirror.com/apk/faultexception/lithium-epub-reader/lithium-epub-reader-0-24-5-release/"
	LithiumSHA = "455cc8371a69ba0cd1f77906f799de751cf74c2974c4c86a60d986ed0070642e"
)

func main() {
	fmt.Printf("info: checking for existing apk\n")
	if s, err := shaFile(LithiumAPK); !errors.Is(err, fs.ErrNotExist) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to read existing apk %q: %v\n", LithiumAPK, err)
			os.Exit(1)
		}
		if s != LithiumSHA {
			fmt.Fprintf(os.Stderr, "error: existing apk %q has incorrect checksum %s (expected %s)", LithiumAPK, s, LithiumSHA)
			os.Exit(1)
		}
		fmt.Printf("info: verified apk %q\n", LithiumAPK)
		return
	}
	fmt.Printf("info: downloading apk from %s\n", LithiumURL)
	if b, err := fetch(LithiumURL); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to download apk: %v\n", err)
		os.Exit(1)
	} else if s := sha(b); s != LithiumSHA {
		fmt.Fprintf(os.Stderr, "error: downloaded apk has incorrect checksum %s (expected %s)\n", s, LithiumSHA)
		os.Exit(1)
	} else if err := os.WriteFile(LithiumAPK, b, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to save apk: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("info: downloaded and verified apk %q\n", LithiumAPK)
}

func sha(data []byte) string {
	s := sha256.Sum256(data)
	return hex.EncodeToString(s[:])
}

func shaFile(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	s := h.Sum(nil)
	return hex.EncodeToString(s[:]), nil
}

func fetch(url string) ([]byte, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar: %w", err)
	}

	cl := &http.Client{
		Transport: headerTransport{
			Header: http.Header{
				textproto.CanonicalMIMEHeaderKey("Accept"):          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
				textproto.CanonicalMIMEHeaderKey("Accept-Language"): {"en-US;q=0.7,en;q=0.3"},
				textproto.CanonicalMIMEHeaderKey("User-Agent"):      {"Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0"},
			},
		},
		Jar: jar,
	}

	referer, downloadPageURL, err := func() (string, string, error) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return "", "", fmt.Errorf("make request: %w", err)
		}
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "same-origin")

		resp, err := cl.Do(req)
		if err != nil {
			return "", "", fmt.Errorf("make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", "", fmt.Errorf("response status %d (%s)", resp.StatusCode, resp.Status)
		}

		doc, err := html.Parse(resp.Body)
		if err != nil {
			return "", "", fmt.Errorf("parse html: %w", err)
		}

		if m := cascadia.QueryAll(doc, cascadia.MustCompile("a.downloadButton")); len(m) != 1 {
			return "", "", fmt.Errorf("find download button: expected exactly one match, got %d", len(m))
		} else {
			for _, a := range m[0].Attr {
				if a.Key == "href" {
					u, err := resp.Request.URL.Parse(a.Val)
					if err != nil {
						return "", "", fmt.Errorf("find download button: failed to resolve href %q against page url %q: %w", a.Val, resp.Request.URL.String(), err)
					}
					return resp.Request.URL.String(), u.String(), nil
				}
			}
			return "", "", fmt.Errorf("find download button: no href attribute")
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("scrape download page url from %q: %w", url, err)
	}
	fmt.Printf("info: got download page %q\n", downloadPageURL)

	referer, downloadURL, err := func() (string, string, error) {
		req, err := http.NewRequest(http.MethodGet, downloadPageURL, nil)
		if err != nil {
			return "", "", fmt.Errorf("make request: %w", err)
		}
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Header.Set("Referer", referer)

		resp, err := cl.Do(req)
		if err != nil {
			return "", "", fmt.Errorf("make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", "", fmt.Errorf("response status %d (%s)", resp.StatusCode, resp.Status)
		}

		doc, err := html.Parse(resp.Body)
		if err != nil {
			return "", "", fmt.Errorf("parse html: %w", err)
		}

		if m := cascadia.QueryAll(doc, cascadia.MustCompile("p > span > a:contains('here')")); len(m) != 1 {
			return "", "", fmt.Errorf("find download link: expected exactly one match, got %d", len(m))
		} else {
			for _, a := range m[0].Attr {
				if a.Key == "href" {
					u, err := resp.Request.URL.Parse(a.Val)
					if err != nil {
						return "", "", fmt.Errorf("find download link: failed to resolve href %q against page url %q: %w", a.Val, resp.Request.URL.String(), err)
					}
					return resp.Request.URL.String(), u.String(), nil
				}
			}
			return "", "", fmt.Errorf("find download link: no href attribute")
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("scrape download link url from %q: %w", downloadPageURL, err)
	}
	fmt.Printf("info: got download link %q\n", downloadURL)

	buf, err := func() ([]byte, error) {
		req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
		if err != nil {
			return nil, fmt.Errorf("make request: %w", err)
		}
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Header.Set("Referer", referer)

		resp, err := cl.Do(req)
		if err != nil {
			return nil, fmt.Errorf("make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("response status %d (%s)", resp.StatusCode, resp.Status)
		}

		if a, e := resp.Header.Get("Content-Type"), "application/vnd.android.package-archive"; a != e {
			return nil, fmt.Errorf("got content type %q, expected %q", a, e)
		}
		return io.ReadAll(resp.Body)
	}()
	if err != nil {
		return nil, fmt.Errorf("download apk from %q: %w", downloadURL, err)
	}
	fmt.Printf("info: got %d bytes\n", len(buf))

	return buf, nil
}

type headerTransport struct {
	Transport http.RoundTripper
	Header    http.Header
}

func (t headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Header {
		req.Header[k] = v
	}
	if t.Transport != nil {
		return t.Transport.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}
