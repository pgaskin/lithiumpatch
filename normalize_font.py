#!/usr/bin/env python3
"""Helper script to inspect, normalize subfamily/family names, and clean TTF fonts for lithiumpatch."""

import argparse
import sys
from fontTools.ttLib import TTFont


def normalize_font(font_path, family=None, subfamily=None, drop_kern=False):
    print(f"Processing {font_path}")
    try:
        font = TTFont(font_path)

        if drop_kern and "kern" in font:
            del font["kern"]
            print("  Removed 'kern' table")

        if subfamily or family:
            for r in font["name"].names:
                if subfamily and r.nameID == 2:
                    r.string = subfamily.encode("utf-16be" if r.platformID != 1 else "ascii")
                if family and r.nameID in (1, 4):
                    r.string = family.encode("utf-16be" if r.platformID != 1 else "ascii")
            if subfamily:
                print(f"  Set subfamily to '{subfamily}'")
            if family:
                print(f"  Set family to '{family}'")

        font.save(font_path)
        print("  Saved successfully")
    except Exception as e:
        print(f"  Error: {e}")


def main():
    parser = argparse.ArgumentParser(description="Normalize TTF font metadata for lithiumpatch")
    parser.add_argument("fonts", nargs="+", help="Path to TTF font files")
    parser.add_argument("--subfamily", default="Regular", help="Subfamily name to set (default: Regular)")
    parser.add_argument("--family", default=None, help="Optional family name to override")
    parser.add_argument("--drop-kern", action="store_true", help="Remove 'kern' table")
    args = parser.parse_args()

    for path in args.fonts:
        normalize_font(path, family=args.family, subfamily=args.subfamily, drop_kern=args.drop_kern)


if __name__ == "__main__":
    main()
