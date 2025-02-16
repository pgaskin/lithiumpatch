# lithiumpatch

Adds additional functionality to the Lithium EPUB Reader Android app.

## Features

- Requires Android Oreo (8) or higher.
- Latest WebView (at least ~73) required for full functionality.
- Custom icon color.
- Monochrome adaptive icon support.
- Dynamic grid view cover width and custom aspect ratio.
- Optional cover-only grid view.
- Debuggable reader webview.
- Dictionary (works offline).
- Custom fonts.
- Additional font script support (e.g., Thai).
- Smaller minimum font size.
- Additional information in the reader footer.
- Series metadata support.
- Series section in library drawer.
- Increased number of visible actions for the reader toolbar.
- Support for inverted portrait/landscape rotation.
- Expand display settings popup by default.
- Support for hyphenation.
- Additional built-in themes.
- Material You colors on Android 12+.
- Full-bleed background in fullscreen mode on devices with a notch.
- Option to disable page turn animations (e.g., for e-ink screens).
- Option to invert the color of images or the entire page.

## Usage

1. Install JRE 1.8 or newer.
2. Install Go 1.21 or newer.
3. Install zipalign (part of the Android build tools).
4. Optionally run `go generate ./dict/edgedict` to download additional dictionaries.
5. Optionally download additional fonts into the `fonts` directory to add additional fonts (to limit them to a single language, put them in a subdirectory named `latin`/`cyrillic`/`greek`/`thai`).
6. Run `go generate ./app` from the root of the repository to download the APK. If this does not work, you can manually download the Lithium 0.24.5 APK from [here](https://www.apkmirror.com/apk/faultexception/lithium-epub-reader/lithium-epub-reader-0-24-5-release/lithium-epub-reader-0-24-5-android-apk-download/) or extract it from your device.
7. Run `go run . app/Lithium_0.24.5.apk` from the root of the repository. Use `--help` to see additional options including using a custom keystore, setting the tool paths, and adding fonts from an external directory.
8. For Google Drive support, specify a custom keystore with `--keystore whatever.jks`, and create a new Google APIs project with access to the Drive API for the signing key's signature to enable sync.

```
usage: lithiumpatch [options] APK_PATH

options:
  -k, --keystore string              Path to keystore for signing (will be created if does not exist) (default "keystore.jks")
      --keystore-alias string        Keystore alias (default "default")
      --keystore-passphrase string   Keystore passphrase (default "default")
  -o, --output string                Output APK path (default: {basename}.patched.resigned.apk)
  -d, --diff string                  Write diff to the specified file (default: disabled)
      --add-fonts strings            Add extra TTF fonts from a directory (Regular/Roman, Bold, Italic, and BoldItalic variants should be provided) (can be specified multiple times)
      --apktool string               Path to apktool.jar (2.8.1) (default "lib/apktool-2.8.1.jar")
      --apksigner string             Path to apksigner.jar (0.9 or later) (default "lib/apksigner-0.9.jar")
      --zipalign string              zipalign executable (will search PATH) (default "zipalign")
      --keytool string               keytool executable (will search PATH) (default "keytool")
  -q, --quiet                        Do not show the diff
      --help                         Show this help text
```

**Note:** If you get an error from apktool about `No resource identifier found for attribute 'preserveLegacyExternalStorage'`, run `java -jar lib/apktool-2.8.1.jar empty-framework-dir`.
