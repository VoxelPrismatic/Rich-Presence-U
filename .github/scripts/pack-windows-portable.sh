#!/usr/bin/env bash
# Build the portable Windows exe as one static binary.
# MSYS2's shared Qt is what the linked zip uses. This script points
# pkg-config at a private Qt6Widgets.pc for /ucrt64/qt6-static, imports
# the platform, style, SVG, and JPEG plugins, and links the MinGW
# runtime and Qt's image/text libraries into the exe.
set -euo pipefail

out="${1:?portable exe path}"
root="$(cd "$(dirname "$0")/../.." && pwd)"
prefix="${QT6_STATIC_PREFIX:-/ucrt64/qt6-static}"
lib="$prefix/lib"
plug="$prefix/share/qt6/plugins"

export PATH="/ucrt64/bin:${PATH:-}"

need=(
  "$lib/libQt6Widgets.a"
  "$lib/libQt6Gui.a"
  "$lib/libQt6Core.a"
  "$lib/libQt6Network.a"
  "$lib/libQt6Svg.a"
  "$lib/libQt6OpenGL.a"
  "$plug/platforms/libqwindows.a"
  "$plug/styles/libqmodernwindowsstyle.a"
  "$plug/iconengines/libqsvgicon.a"
  "$plug/imageformats/libqsvg.a"
  "$plug/imageformats/libqjpeg.a"
)
for f in "${need[@]}"; do
  if [ ! -f "$f" ]; then
    echo "static Qt file missing: $f" >&2
    echo "Install mingw-w64-ucrt-x86_64-qt6-static." >&2
    exit 1
  fi
done

objs=(
  "$lib/objects-Release/Widgets_resources_1/.qt/rcc/qrc_qstyle_init.cpp.obj"
  "$lib/objects-Release/Widgets_resources_2/.qt/rcc/qrc_qstyle1_init.cpp.obj"
  "$lib/objects-Release/Widgets_resources_3/.qt/rcc/qrc_qstyle_fusion_init.cpp.obj"
  "$lib/objects-Release/Widgets_resources_4/.qt/rcc/qrc_qmessagebox_init.cpp.obj"
  "$lib/objects-Release/Gui_resources_1/.qt/rcc/qrc_qpdf_init.cpp.obj"
  "$lib/objects-Release/Gui_resources_2/.qt/rcc/qrc_gui_shaders_init.cpp.obj"
  "$lib/objects-Release/QWindowsIntegrationPlugin_resources_1/.qt/rcc/qrc_openglblacklists_init.cpp.obj"
  "$lib/objects-Release/QWindowsIntegrationPlugin_resources_2/.qt/rcc/qrc_cursors_init.cpp.obj"
)
for f in "${objs[@]}"; do
  if [ ! -f "$f" ]; then
    echo "static Qt resource missing: $f" >&2
    exit 1
  fi
done

pcdir="$(mktemp -d)"
linkgo="$root/ui/static_link_windows.go"
trap 'rm -rf "$pcdir"; rm -f "$linkgo"' EXIT

# setup-go is a Windows binary. It does not translate MSYS paths, so
# /ucrt64/... never reaches gcc. cygpath -m gives D:/.../ucrt64/...
win() { cygpath -m "$1"; }
prefix_win="$(win "$prefix")"

# miqt asks pkg-config for Qt6Widgets twice and folds --static into that
# one call, so anything listed here is repeated. Archives and system
# libraries survive that. Resource objects do not, so those are written
# into one package's #cgo LDFLAGS below.
cat >"$pcdir/Qt6Widgets.pc" <<EOF
prefix=$prefix_win
libdir=\${prefix}/lib
includedir=\${prefix}/include/qt6

Name: Qt6Widgets
Description: Static Qt6 Widgets
Version: 6
Libs: -L\${libdir} -Wl,--start-group -lQt6Svg -lQt6OpenGL -lQt6Widgets -lQt6Gui -lQt6Network -lQt6Core -Wl,--end-group -Wl,-Bstatic -Wl,--start-group -lpng -lpng16 -ljpeg -lz -lharfbuzz -lfreetype -lbz2 -lbrotlidec -lbrotlicommon -lglib-2.0 -lintl -liconv -lffi -lpcre2-16 -lpcre2-8 -lgraphite2 -lb2 -lssl -lcrypto -lwinpthread -Wl,--end-group -Wl,-Bdynamic -lsynchronization -lmpr -luserenv -lauthz -lkernel32 -lnetapi32 -lntdll -lruntimeobject -lversion -lwinmm -lws2_32 -ld3d11 -ldxgi -ldxguid -ld3d12 -ld3d9 -ladvapi32 -lgdi32 -lole32 -loleaut32 -lshell32 -luser32 -luuid -lusp10 -lrpcrt4 -ldwrite -ld2d1 -ldwmapi -luxtheme -limm32 -lsetupapi -lshlwapi -lwinspool -lwtsapi32 -lshcore -lcomdlg32 -ldnsapi -liphlpapi -lsecur32 -lwinhttp -lcrypt32 -latomic
Cflags: -I\${includedir}/QtWidgets -I\${includedir}/QtGui -I\${includedir}/QtCore -I\${includedir}/QtSvg -I\${includedir} -I\${prefix}/share/qt6/mkspecs/win32-g++ -DQT_WIDGETS_LIB -DQT_GUI_LIB -DQT_SVG_LIB -DQT_CORE_LIB -DQT_STATICPLUGIN
EOF

plugins=(
  "$plug/platforms/libqwindows.a"
  "$plug/styles/libqmodernwindowsstyle.a"
  "$plug/iconengines/libqsvgicon.a"
  "$plug/imageformats/libqsvg.a"
  "$plug/imageformats/libqjpeg.a"
)
{
  printf '%s\n' '//go:build windows && windowsqtstatic' '' 'package ui' '' '/*' '#cgo LDFLAGS: -Wl,--start-group'
  for f in "${objs[@]}" "${plugins[@]}"; do
    printf '#cgo LDFLAGS: %s\n' "$(win "$f")"
  done
  printf '%s\n' '#cgo LDFLAGS: -Wl,--end-group' '*/' 'import "C"'
} >"$linkgo"

inc="$(win "$prefix/include/qt6")"
spec="$(win "$prefix/share/qt6/mkspecs/win32-g++")"
includes="-I${inc}/QtWidgets -I${inc}/QtGui -I${inc}/QtCore -I${inc}/QtSvg -I${inc} -I${spec}"
export PKG_CONFIG_PATH="$(win "$pcdir")"
export CGO_ENABLED=1
# Replace, rather than append, so the shared Qt include path cannot leak in.
export CGO_CPPFLAGS="$includes"
export CGO_CXXFLAGS="-std=c++17 -DQT_STATICPLUGIN $includes"
# -static makes ld prefer .a over .dll.a. Qt DLLs import the shared GCC
# runtime, so this flag is only safe once those DLLs are gone.
export CGO_LDFLAGS="-static -static-libgcc -static-libstdc++"
# Go's linker allow-list matches ".o" before ".obj", so a path ending in
# ".cpp.obj" is rejected. Permit the Windows resource objects and plugin
# archives this script writes into #cgo LDFLAGS.
export CGO_LDFLAGS_ALLOW='^[A-Za-z]:/.*\.(obj|a)$'
out_win="$(win "$out")"

mkdir -p "$(dirname "$out")"
(
  cd "$root"
  go build -tags "portable,windowsqtstatic" -ldflags "-s -w -H windowsgui" -o "$out_win" .
)

# These are the DLLs the shared portable folder used to ship. A hit means
# the exe still needs that file beside it.
banned='Qt6(Widgets|Svg|Network|Gui|Core)\.dll|libwinpthread-1\.dll|libstdc\+\+-6\.dll|libgcc_s_seh-1\.dll|libfreetype-6\.dll|libharfbuzz-0\.dll|libpng16-16\.dll|libmd4c\.dll'
imports="$(objdump -p "$out_win" | sed -n 's/^[[:space:]]*DLL Name: //p')"
echo "DLL imports:"
printf '%s\n' "$imports"
if printf '%s\n' "$imports" | grep -Eiq "$banned"; then
  echo "portable exe still imports a library that should be linked in" >&2
  exit 1
fi
