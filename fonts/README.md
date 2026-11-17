# Fonts Directory

This directory contains custom fonts to be embedded in the patched Lithium APK.

## Supported Formats

- **TTF** (TrueType Font) - ✓ Supported
- **OTF** (OpenType Font with TrueType outlines) - ✓ Supported
- **OTF** (OpenType Font with CFF/PostScript outlines) - ✗ Not supported
- **WOFF2** - ✗ Not supported

## Font Requirements

Each font family requires at minimum a **Regular** (or **Roman**) variant.
Optional variants include **Bold**, **Italic**, and **BoldItalic**.

## Incompatible Files in This Directory

The following files cannot be used and are ignored during build:

### CFF-based OTF files (incompatible with truetype parser):
- `Adobe Jenson Pro Italic.otf`
- `Adobe Jenson Pro Regular.otf`
- `Bastarda.otf`
- `GreatVictorian-Standard.otf`
- `RusticRoadway.otf`
- `qt-bookmann/*.otf` (QTBookmann)

### WOFF2 files (not supported):
- `Anthropic-serif.woff2` (TTF version available: `Anthropic-serif.ttf`)

## Current Status

**33 complete font families** are available and embedded in the APK (including EB Garamond, Leander, Aquifer, Essays1743, Roman Antique, Subway Haze, Bastarda, Lumos, and Peaky Blinders).
