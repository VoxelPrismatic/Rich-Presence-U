#!/usr/bin/env bash
set -euo pipefail

bin="${1:?binary}"
app="${2:?app bundle name}"

rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
cp "$bin" "$app/Contents/MacOS/rich-presence-qt"
chmod +x "$app/Contents/MacOS/rich-presence-qt"
cp logo.png "$app/Contents/Resources/logo.png"

cat > "$app/Contents/Info.plist" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleExecutable</key>
	<string>rich-presence-qt</string>
	<key>CFBundleIconFile</key>
	<string>logo</string>
	<key>CFBundleIdentifier</key>
	<string>dev.voxelprismatic.richpresenceqt</string>
	<key>CFBundleName</key>
	<string>Rich Presence Qt</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>2.7.0</string>
	<key>LSMinimumSystemVersion</key>
	<string>12.0</string>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
EOF

macdeployqt="$(command -v macdeployqt || true)"
if [ -z "$macdeployqt" ]; then
	macdeployqt="$(brew --prefix qt)/bin/macdeployqt"
fi
"$macdeployqt" "$app" -always-overwrite

ditto -c -k --keepParent "$app" "${app}.zip"
