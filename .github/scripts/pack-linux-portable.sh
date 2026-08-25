#!/usr/bin/env bash
set -euo pipefail

bin="${1:?binary}"
out="${2:?output name}"

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT
appdir="$workdir/AppDir"
mkdir -p "$appdir/usr/bin" "$appdir/usr/share/applications" "$appdir/usr/share/icons/hicolor/256x256/apps"

cp "$bin" "$appdir/usr/bin/rich-presence-qt"
chmod +x "$appdir/usr/bin/rich-presence-qt"
cp logo.png "$appdir/usr/share/icons/hicolor/256x256/apps/rich-presence-qt.png"

cat > "$appdir/usr/share/applications/rich-presence-qt.desktop" << 'EOF'
[Desktop Entry]
Type=Application
Name=Rich Presence Qt
Exec=rich-presence-qt
Icon=rich-presence-qt
Categories=Game;Network;
Terminal=false
EOF

curl -fsSL -o "$workdir/linuxdeploy.AppImage" \
  https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
curl -fsSL -o "$workdir/linuxdeploy-plugin-qt.AppImage" \
  https://github.com/linuxdeploy/linuxdeploy-plugin-qt/releases/download/continuous/linuxdeploy-plugin-qt-x86_64.AppImage
chmod +x "$workdir"/linuxdeploy*.AppImage

export APPIMAGE_EXTRACT_AND_RUN=1
export QMAKE="$(command -v qmake6 || command -v qmake)"
export EXTRA_QT_PLUGINS="svg;iconengines;imageformats"

"$workdir/linuxdeploy.AppImage" \
  --appdir "$appdir" \
  --executable "$appdir/usr/bin/rich-presence-qt" \
  --desktop-file "$appdir/usr/share/applications/rich-presence-qt.desktop" \
  --icon-file "$appdir/usr/share/icons/hicolor/256x256/apps/rich-presence-qt.png" \
  --plugin qt \
  --output appimage

shopt -s nullglob
imgs=("$PWD"/*.AppImage)
if [ ${#imgs[@]} -eq 0 ]; then
  echo "linuxdeploy produced no AppImage" >&2
  ls -la "$PWD"
  exit 1
fi
mv "${imgs[0]}" "$out"
chmod +x "$out"
