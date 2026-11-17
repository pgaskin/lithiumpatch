import sys
from fontTools.ttLib import TTFont

def remove_kern_table(font_path):
    print(f"Processing {font_path}")
    try:
        font = TTFont(font_path)
        if 'kern' in font:
            del font['kern']
            print("  Removed 'kern' table")
            font.save(font_path)
        else:
            print("  No 'kern' table found")
    except Exception as e:
        print(f"  Error: {e}")

if __name__ == "__main__":
    for path in sys.argv[1:]:
        remove_kern_table(path)
