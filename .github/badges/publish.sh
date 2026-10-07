#!/usr/bin/env bash
# Egy könyvtárnyi badge (SVG) kitolása a repó `badges` ágára; a README onnan hivatkozza őket:
#
#   GH_TOKEN=... ./publish.sh <könyvtár>
#
# Az ág árva és egyetlen commitból áll, minden futás felülírja (force push): a badge-eknek nincs
# megőrzendő története, így a repó sem hízik tőlük. A CI-ban a GITHUB_TOKEN-nel fut, `contents:
# write` joggal; az ilyen push nem indít újabb workflow-t.
set -euo pipefail

dir="${1:?Használat: publish.sh <könyvtár>}"
: "${GH_TOKEN:?GH_TOKEN kell a pushhoz}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY kell (owner/repo)}"

cd "$dir"
ls ./*.svg >/dev/null

rm -rf .git
git init -q -b badges
git add -- *.svg
git -c user.name="github-actions[bot]" -c user.email="41898282+github-actions[bot]@users.noreply.github.com" \
  commit -q -m "Badge-ek frissítése (${GITHUB_SHA:-kézi futás})"
# A token csak ennek az egy parancsnak a fejlécében él, nem kerül a git konfigurációba.
git -c "http.https://github.com/.extraheader=AUTHORIZATION: basic $(printf 'x-access-token:%s' "$GH_TOKEN" | base64 | tr -d '\n')" \
  push -q --force "https://github.com/${GITHUB_REPOSITORY}.git" badges
echo "badges ág frissítve: $(ls ./*.svg | tr '\n' ' ')"
