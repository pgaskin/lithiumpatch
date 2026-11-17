#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"

# Add common Go and Android build-tools paths to PATH just in case
export PATH=$HOME/.cache/codex-tools/build-tools/android-16:$HOME/Android/Sdk/build-tools/36.1.0:$HOME/sdk/go/bin:$PATH:/usr/local/go/bin:$HOME/go/bin

echo "Checking for Go..."
if ! command -v go &> /dev/null; then
    echo "Error: 'go' command not found. Please ensure Go is installed and in your PATH."
    exit 1
fi

echo "Generating assets..."
go generate ./app

echo "Building patched APK..."
INPUT_APK="app/$(sed -n 's/.*LithiumAPK[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' app/app.go)"
OUTPUT_APK="${INPUT_APK%.apk}.patched.resigned.apk"

go build -o lithiumpatch .
./lithiumpatch "$INPUT_APK"

echo "Installing to device..."

if [ ! -f "$OUTPUT_APK" ]; then
    echo "Error: Output APK $OUTPUT_APK not found."
    exit 1
fi

adb install -r "$OUTPUT_APK"
echo "Launching app..."
adb shell am start -W -n com.faultexception.reader/com.faultexception.reader.MainActivity
echo "Success! App installed and launched."
