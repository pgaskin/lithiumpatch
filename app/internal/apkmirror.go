package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

// FetchAPK_APKM fetches a single non-split APK from an APKMirror URL.
func FetchAPK_APKM(url string) ([]byte, error) {
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

		if a, e, e1 := resp.Header.Get("Content-Type"), "application/vnd.android.package-archive", "application/octet-stream"; a != e && a != e1 {
			return nil, fmt.Errorf("got content type %q, expected %q or %q", a, e, e1)
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
