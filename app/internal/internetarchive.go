package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

// FetchAPK_IA fetches an APK from an internet archive snapshot URL.
func FetchAPK_IA(filterURL string) ([]byte, error) {
	archiveURL, err := func() (string, error) {
		req, err := http.NewRequest(http.MethodGet, "https://archive.org/wayback/available?url="+url.QueryEscape(filterURL), nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to make request to internet archive api: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("internet archive response status %d (%s)", resp.StatusCode, resp.Status)
		}

		if mt, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type")); mt != "application/json" {
			return "", fmt.Errorf("internet archive returned non-json response (got %q)", mt)
		}

		var obj struct {
			URL               string `json:"url"`
			ArchivedSnapshots struct {
				Closest struct {
					Status    string `json:"status"`
					Available bool   `json:"available"`
					URL       string `json:"url"`
					Timestamp string `json:"timestamp"`
				} `json:"closest"`
			} `json:"archived_snapshots"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
			return "", fmt.Errorf("failed to parse internet archive response: %w", err)
		}
		if !obj.ArchivedSnapshots.Closest.Available {
			return "", fmt.Errorf("no snapshots available")
		}
		if obj.ArchivedSnapshots.Closest.Status != "200" {
			return "", fmt.Errorf("latest snapshot was not successful")
		}
		if obj.ArchivedSnapshots.Closest.URL == "" {
			return "", fmt.Errorf("latest snapshot does not have a url")
		}

		// make the link direct
		u, err := url.Parse(obj.ArchivedSnapshots.Closest.URL)
		if err != nil {
			return "", fmt.Errorf("failed to parse snapshot url %q: %w", obj.ArchivedSnapshots.Closest.URL, err)
		}
		if spl := strings.Split(u.Path, "/"); len(spl) < 3 || spl[0] != "" || spl[1] != "web" || strings.HasSuffix(spl[2], "_") {
			return "", fmt.Errorf("failed to parse snapshot url %q: does not match expected format", obj.ArchivedSnapshots.Closest.URL)
		} else {
			spl[2] += "im_"
			u.Path = strings.Join(spl, "/")
		}
		return u.String(), nil
	}()
	if err != nil {
		return nil, fmt.Errorf("get internet archive url: %w", err)
	}
	fmt.Printf("info: got archive url %q\n", archiveURL)

	buf, err := func() ([]byte, error) {
		resp, err := http.Get(archiveURL)
		if err != nil {
			return nil, fmt.Errorf("make request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("response status %d (%s)", resp.StatusCode, resp.Status)
		}

		if a, e := resp.Header.Get("Content-Type"), "application/vnd.android.package-archive"; a != e {
			return nil, fmt.Errorf("got content type %q, expected %q", a, e)
		}
		return io.ReadAll(resp.Body)
	}()
	if err != nil {
		return nil, fmt.Errorf("download apk from %q: %w", archiveURL, err)
	}
	fmt.Printf("info: got %d bytes\n", len(buf))

	return buf, nil
}
