#!/usr/bin/env bash
# Build the macOS app bundle with Qt inside it.
# A .app is a directory. macdeployqt copies the frameworks and plugins
# the binary needs and rewrites them to @executable_path, which is what
# makes the bundle portable.
set -euo pipefail

bin="${1:?binary}"
app="${2:?app bundle name}"

rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
cp "$bin" "$app/Contents/MacOS/rich-presence-qt"
chmod +x "$app/Contents/MacOS/rich-presence-qt"
cp logo.png "$app/Contents/Resources/logo.png"

ver="$(sed -n 's/^const VERSION = "\(.*\)"/\1/p' svc/version.go | head -n1)"
if [ -z "$ver" ]; then
	echo "could not read VERSION from svc/version.go" >&2
	exit 1
fi

cat > "$app/Contents/Info.plist" << EOF
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
	<string>${ver}</string>
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
qtprefix="$(brew --prefix qt)"
plugins=""
for dir in "$qtprefix/share/qt/plugins" "$qtprefix/plugins"; do
	if [ -d "$dir" ]; then
		plugins="$dir"
		break
	fi
done
if [ -z "$plugins" ]; then
	echo "Qt plugins directory not found under $qtprefix" >&2
	exit 1
fi

"$macdeployqt" "$app" -always-overwrite

# Covers are JPEG. Icons are SVG. Copy any of those plugins macdeployqt
# left behind, then run it again so their framework dependencies are
# rewritten into the bundle.
needed=(
	imageformats/libqjpeg.dylib
	imageformats/libqsvg.dylib
	iconengines/libqsvgicon.dylib
)
for rel in "${needed[@]}"; do
	dest="$app/Contents/PlugIns/$rel"
	if [ -e "$dest" ]; then
		continue
	fi
	src="$plugins/$rel"
	if [ ! -f "$src" ]; then
		echo "Qt plugin not found: $src" >&2
		exit 1
	fi
	mkdir -p "$(dirname "$dest")"
	cp "$src" "$dest"
done
"$macdeployqt" "$app" -always-overwrite

require=(
	Contents/Frameworks/QtCore.framework/QtCore
	Contents/Frameworks/QtGui.framework/QtGui
	Contents/Frameworks/QtWidgets.framework/QtWidgets
	Contents/Frameworks/QtSvg.framework/QtSvg
	Contents/PlugIns/platforms/libqcocoa.dylib
	Contents/PlugIns/imageformats/libqjpeg.dylib
	Contents/PlugIns/imageformats/libqsvg.dylib
	Contents/PlugIns/iconengines/libqsvgicon.dylib
)
for rel in "${require[@]}"; do
	if [ ! -e "$app/$rel" ]; then
		echo "bundle is missing $rel" >&2
		exit 1
	fi
done

# A Homebrew path here means the bundle still needs the build machine.
if otool -L "$app/Contents/MacOS/rich-presence-qt" "$app/Contents/PlugIns/platforms/libqcocoa.dylib" "$app/Contents/PlugIns/iconengines/libqsvgicon.dylib" | grep -E '/opt/homebrew|/usr/local/opt'; then
	echo "bundle still references a Homebrew library" >&2
	exit 1
fi

# GitHub can attach a file, not a directory. ditto keeps the framework
# symlinks. The zip name matches the bytes.
ditto -c -k --keepParent "$app" "${app}.zip"
rm -rf "$app"
