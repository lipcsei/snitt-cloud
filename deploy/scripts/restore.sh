#!/usr/bin/env bash
# Visszaállítás egy mentésből - a "snitt" (account szolgáltatás) adatbázis.
#
# A Keycloak adatbázisának visszaállításához a megosztott
# https://github.com/lipcsei/sso repó saját scripts/restore.sh-ja kell -
# az nem itt van.
#
# A mentés csak akkor ér valamit, ha a visszaállítás is le van írva - és ha
# kipróbáltad. Egy soha ki nem próbált mentés nem mentés, csak remény.
#
# Használat:
#   ./deploy/scripts/restore.sh /var/backups/snitt/snitt-20260910-030000.sql.gz

set -euo pipefail

FILE="${1:?add meg a mentés fájlját}"
COMPOSE_FILE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/docker-compose.yml"
DB="${POSTGRES_DB:-snitt}"

echo "FIGYELEM: a(z) '$DB' adatbázis jelenlegi tartalma felülíródik."
read -r -p "Folytatod? (igen/nem) " valasz
[ "$valasz" = "igen" ] || { echo "megszakítva"; exit 1; }

# Az account szolgáltatás ne írjon, amíg cserélünk alóla.
docker compose -f "$COMPOSE_FILE" stop account >/dev/null

gunzip -c "$FILE" | docker compose -f "$COMPOSE_FILE" exec -T postgres \
  psql -U "${POSTGRES_USER:-snitt}" -d "$DB" -v ON_ERROR_STOP=1

docker compose -f "$COMPOSE_FILE" start account >/dev/null
echo "kész: $DB visszaállítva a(z) $FILE fájlból"
