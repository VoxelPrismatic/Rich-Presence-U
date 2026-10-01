#!/usr/bin/env bash
# Embed logo.png as the Windows exe icon. Explorer, the taskbar, and
# shortcuts read that resource from the exe itself.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
export PATH="/ucrt64/bin:${PATH:-}"

if ! command -v windres >/dev/null 2>&1; then
	echo "windres is required to embed the exe icon" >&2
	exit 1
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

ico="$work/logo.ico"
syso="$root/rsrc_windows_amd64.syso"
if command -v cygpath >/dev/null 2>&1; then
	ico="$(cygpath -m "$ico")"
	syso="$(cygpath -m "$syso")"
fi
(
	cd "$root"
	go run .github/scripts/winicon.go -png logo.png -ico "$ico"
)
printf '%s\n' '1 ICON "logo.ico"' >"$work/logo.rc"
# Run beside the rc file so windres finds logo.ico. Bash rewrites the
# absolute output path for the Windows windres binary.
(cd "$work" && windres -O coff -i logo.rc -o "$syso")
