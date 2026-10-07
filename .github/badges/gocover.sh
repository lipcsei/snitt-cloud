#!/usr/bin/env bash
# Go lefedettség egész százalékban egy vagy több coverprofile-ból, a generált fájlok nélkül:
#
#   go test ./... -coverpkg=./... -coverprofile=coverage.out
#   ./gocover.sh coverage.out            # -> 73
#
# A generált kódot (sqlc, oapi-codegen, mockok: "// Code generated ... DO NOT EDIT." fejléc) nem
# számolja: azt nem mi írjuk és nem is teszteljük, a nagy tömege viszont elnyomná a saját kód
# arányát. A profilt abból a modulból kell futtatni, amelyikről szól (a go list ott oldja fel a
# csomagok könyvtárát). Utasítás-szinten számol, ahogy a `go tool cover -func` összesítője. A
# -coverpkg=./... azért kell, hogy egy csomag tesztje a többi csomagban bejárt kódot is fedettnek
# számolja (a service integrációs tesztjei a store-t is végigjárják).
set -euo pipefail

[ "$#" -ge 1 ] || { echo "Használat: gocover.sh <coverprofile>..." >&2; exit 2; }

# importútvonal -> könyvtár, hogy a profil fájlneveiből (importútvonal/fájl.go) valódi útvonal legyen.
dirs="$(go list -f '{{.ImportPath}} {{.Dir}}' ./...)"

generated="$(
  cat "$@" | grep -v '^mode:' | cut -d: -f1 | sort -u | while read -r file; do
    pkg="${file%/*}"
    dir="$(awk -v p="$pkg" '$1 == p {print $2; exit}' <<<"$dirs")"
    [ -n "$dir" ] || continue
    if head -n 5 "$dir/${file##*/}" 2>/dev/null | grep -qE '^// Code generated .* DO NOT EDIT\.$'; then
      echo "$file"
    fi
  done
)"

# Egy blokk több profilban (vagy -coverpkg mellett több csomag tesztjéből) is szerepelhet: akkor
# fedett, ha bármelyikben lefutott. A generált fájlok listája a szabványos bemeneten megy át (a BSD
# awk -v változója nem tűr sortörést), a profilok fájlként.
awk '
  FILENAME == "-" { if ($0 != "") skip[$0] = 1; next }
  /^mode:/ { next }
  {
    split($1, loc, ":")
    if (loc[1] in skip) next
    stmts[$1] = $2
    if ($3 > 0) hit[$1] = 1
  }
  END {
    for (b in stmts) { total += stmts[b]; if (b in hit) covered += stmts[b] }
    if (total == 0) { print 0; exit }
    printf "%d\n", covered * 100 / total
  }' - "$@" <<<"$generated"
