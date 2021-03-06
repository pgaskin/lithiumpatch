# lithiumpatch
Adds additional functionality to the Lithium EPUB Reader Android app.

## Features

- Custom icon color.
- Custom cover aspect ratio.
- Optional cover-only grid view.
- Debuggable reader webview.
- Dictionary.
- Custom fonts.
- Smaller minimum font size.
- Additional information in the reader footer.
- Series metadata support.

## Usage

1. Download the Lithium 0.24.1 APK from [here](https://www.apkmirror.com/apk/faultexception/lithium-epub-reader/lithium-epub-reader-0-24-1-release/lithium-epub-reader-0-24-1-android-apk-download/download/) or extract it from your device.
2. Install JRE 1.8 or newer.
3. Install Go 1.18 or newer.
4. Install zipalign (part of the Android build tools).
5. Run `go run . /path/to/Lithium_0.24.1.apk` from the root of the repository. Use `--help` to see additional options including using a custom keystore, setting the tool paths, and adding fonts from an external directory.
6. If you haven't already done so, create a new Google APIs project with access to the Drive API for the signing key's signature to enable sync.

```
usage: lithiumpatch [options] APK_PATH

options:
  -k, --keystore string              Path to keystore for signing (will be created if does not exist) (default "keys/default.jks")
      --keystore-alias string        Keystore alias (default "default")
      --keystore-passphrase string   Keystore passphrase (default "default")
  -o, --output string                Output APK path (default: {basename}.patched.resigned.apk)
  -d, --diff string                  Write diff to the specified file (default: disabled)
      --add-fonts strings            Add extra TTF fonts from a directory (Regular/Roman, Bold, Italic, and BoldItalic variants should be provided) (can be specified multiple times)
      --apktool string               Path to apktool.jar (2.4.1 - 2.6.1) (default "lib/apktool-2.6.1.jar")
      --apksigner string             Path to apksigner.jar (0.8 or later) (default "lib/apksigner-0.8.jar")
      --zipalign string              zipalign executable (will search PATH) (default "zipalign")
      --keytool string               keytool executable (will search PATH) (default "keytool")
  -q, --quiet                        Do not show the diff
      --help                         Show this help text
```
