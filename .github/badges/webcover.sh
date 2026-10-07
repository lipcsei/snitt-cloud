#!/usr/bin/env bash
# Frontend-lefedettség egész százalékban a vitest/istanbul json-summary riportjából:
#
#   npx vitest run --coverage --coverage.reporter=json-summary
#   ./webcover.sh coverage/coverage-summary.json     # -> 81
#
# Utasítás-szinten számol, ahogy a gocover.sh, hogy a két szám ugyanazt jelentse.
set -euo pipefail

summary="${1:?Használat: webcover.sh <coverage-summary.json>}"
node -e '
  const pct = JSON.parse(require("fs").readFileSync(process.argv[1], "utf8")).total.statements.pct;
  console.log(typeof pct === "number" ? Math.floor(pct) : 0);' "$summary"
