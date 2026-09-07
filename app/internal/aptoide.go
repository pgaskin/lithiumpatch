package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
)

// FetchAPK_Aptoide fetches an APK from an Aptoide app/getMeta API URL.
func FetchAPK_Aptoide(apiURL string) ([]byte, error) {
	downloadURL, err := func() (string, error) {
		req, err := http.NewRequest(http.MethodGet, apiURL, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to make request to aptoide api: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("aptoide response status %d (%s)", resp.StatusCode, resp.Status)
		}

		if mt, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type")); mt != "application/json" {
			return "", fmt.Errorf("aptoide returned non-json response (got %q)", mt)
		}

		var obj struct {
			Info struct {
				Status string `json:"status"`
			} `json:"info"`
			Data struct {
				File struct {
					VerName string `json:"vername"`
					VerCode int    `json:"vercode"`
					Path    string `json:"path"`
				} `json:"file"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
			return "", fmt.Errorf("failed to parse aptoide response: %w", err)
		}
		if obj.Info.Status != "OK" {
			return "", fmt.Errorf("aptoide returned status %q", obj.Info.Status)
		}
		if obj.Data.File.Path == "" {
			return "", fmt.Errorf("aptoide response does not have a file path")
		}
		fmt.Printf("info: got version %s (%d)\n", obj.Data.File.VerName, obj.Data.File.VerCode)

		return obj.Data.File.Path, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("get aptoide url: %w", err)
	}
	fmt.Printf("info: got aptoide url %q\n", downloadURL)

	buf, err := func() ([]byte, error) {
		resp, err := http.Get(downloadURL)
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
