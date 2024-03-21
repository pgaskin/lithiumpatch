//go:build ignore

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/pgaskin/lithiumpatch/app"
	"github.com/pgaskin/lithiumpatch/app/internal"
)

func main() {
	fmt.Printf("info: checking for existing apk\n")
	if s, err := shaFile(app.LithiumAPK); !errors.Is(err, fs.ErrNotExist) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to read existing apk %q: %v\n", app.LithiumAPK, err)
			os.Exit(1)
		}
		if s != app.LithiumSHA {
			fmt.Fprintf(os.Stderr, "error: existing apk %q has incorrect checksum %s (expected %s)", app.LithiumAPK, s, app.LithiumSHA)
			os.Exit(1)
		}
		fmt.Printf("info: verified apk %q\n", app.LithiumAPK)
		return
	}
	if b, err := fetch(); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to download apk: %v\n", err)
		os.Exit(1)
	} else if s := sha(b); s != app.LithiumSHA {
		fmt.Fprintf(os.Stderr, "error: downloaded apk has incorrect checksum %s (expected %s)\n", s, app.LithiumSHA)
		os.Exit(1)
	} else if err := os.WriteFile(app.LithiumAPK, b, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to save apk: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("info: downloaded and verified apk %q\n", app.LithiumAPK)
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

func fetch() ([]byte, error) {
	fmt.Printf("info: downloading apk from apkmirror %s\n", app.LithiumURL_APKM)
	if b, err := internal.FetchAPK_APKM(app.LithiumURL_APKM); err != nil {
		fmt.Fprintf(os.Stderr, "warn: failed to download apk from apkmirror (error: %v)\n", err)
	} else {
		return b, nil
	}

	fmt.Printf("info: downloading apk from internet archive %s\n", app.LithiumURL_IA)
	if b, err := internal.FetchAPK_IA(app.LithiumURL_IA); err != nil {
		fmt.Fprintf(os.Stderr, "warn: failed to download apk from internet archive (error: %v)\n", err)
	} else {
		return b, nil
	}

	return nil, fmt.Errorf("all sources failed")
}
