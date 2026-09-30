#!/usr/bin/env bash
# Bundle a MinGW-built portable exe with Qt and the GCC runtime.
# windeployqt copies Qt but often skips libgcc when it cannot see g++
# (typical in CI). Qt and this binary both import that runtime, so a
# missing libgcc_s_seh-1.dll is the dialog Windows shows on launch.
set -euo pipefail

exe="${1:?portable exe}"
dest="$(cd "$(dirname "$exe")" && pwd)"
exe="$dest/$(basename "$exe")"

export PATH="/ucrt64/bin:${PATH:-}"

windeployqt6 --compiler-runtime "$exe"

# Companions of libgcc. Shipping libgcc alone just changes the dialog
# to libstdc++ or libwinpthread.
runtime=(
  libgcc_s_seh-1.dll
  libstdc++-6.dll
  libwinpthread-1.dll
)
for dll in "${runtime[@]}"; do
  src="/ucrt64/bin/$dll"
  if [ ! -f "$src" ]; then
    echo "MinGW runtime not found: $src" >&2
    exit 1
  fi
  cp -f "$src" "$dest/$dll"
done

if [ ! -f "$dest/libgcc_s_seh-1.dll" ]; then
  echo "libgcc was not bundled next to $exe" >&2
  exit 1
fi
