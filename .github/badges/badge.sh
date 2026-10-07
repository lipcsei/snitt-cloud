#!/usr/bin/env bash
# Egy badge SVG-je a szabványos kimenetre, külső szolgáltatás nélkül (a privát repókat a
# shields.io nem látja):
#
#   ./badge.sh <címke> <érték> [szín] > coverage-go.svg
#
# A szín hex kód (#4c1) vagy `auto`: ekkor az érték elején álló számból (százalék) lesz, a
# lefedettség szokásos zöld-sárga-piros skáláján. Érték nélkül (üres) szürke "n/a" készül, hogy egy
# kimaradt mérés ne hagyjon törött képet a README-ben.
set -euo pipefail

label="${1:?Használat: badge.sh <címke> <érték> [szín]}"
value="${2:-}"
color="${3:-auto}"

if [ -z "$value" ]; then
  value="n/a"
  color="#9f9f9f"
fi

if [ "$color" = "auto" ]; then
  num=""
  if [[ "$value" =~ ^([0-9]+) ]]; then num="${BASH_REMATCH[1]}"; fi
  if [ -z "$num" ]; then color="#9f9f9f"
  elif [ "$num" -ge 80 ]; then color="#4c1"
  elif [ "$num" -ge 60 ]; then color="#a4a61d"
  elif [ "$num" -ge 40 ]; then color="#dfb317"
  elif [ "$num" -ge 20 ]; then color="#fe7d37"
  else color="#e05d44"
  fi
fi

# A szöveg szélessége betűcsoportonkénti becsléssel (Verdana 11px): pontos mérés betűkészlet nélkül
# nincs, a textLength viszont a becsült szélességre húzza a szöveget, így nem lóg ki a dobozból.
width() {
  perl -CA -e '
    my $w = 0;
    for (split //, $ARGV[0]) {
      $w += /[ilj.,:;|!\x27]/ ? 3.1 : /[ ftrI\/()-]/ ? 4.3 : /[mwMW%@]/ ? 11 : /[A-Z]/ ? 7.8 : /[0-9]/ ? 7 : 6.6;
    }
    printf "%d\n", $w + 10.5;' "$1"
}

esc() { sed -e 's/&/\&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' -e 's/"/\&quot;/g' <<<"$1"; }

lw="$(width "$label")"
vw="$(width "$value")"
total=$((lw + vw))
label_x=$((lw * 5))
value_x=$((lw * 10 + vw * 5))
label_len=$(((lw - 10) * 10))
value_len=$(((vw - 10) * 10))
label="$(esc "$label")"
value="$(esc "$value")"

cat <<SVG
<svg xmlns="http://www.w3.org/2000/svg" width="$total" height="20" role="img" aria-label="$label: $value">
  <title>$label: $value</title>
  <linearGradient id="s" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r"><rect width="$total" height="20" rx="3" fill="#fff"/></clipPath>
  <g clip-path="url(#r)">
    <rect width="$lw" height="20" fill="#555"/>
    <rect x="$lw" width="$vw" height="20" fill="$color"/>
    <rect width="$total" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="110" text-rendering="geometricPrecision">
    <text x="$label_x" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="$label_len">$label</text>
    <text x="$label_x" y="140" transform="scale(.1)" textLength="$label_len">$label</text>
    <text x="$value_x" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="$value_len">$value</text>
    <text x="$value_x" y="140" transform="scale(.1)" textLength="$value_len">$value</text>
  </g>
</svg>
SVG
