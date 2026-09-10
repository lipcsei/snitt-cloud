#!/usr/bin/env bash
# Visszaállítás egy mentésből.
#
# A mentés csak akkor ér valamit, ha a visszaállítás is le van írva - és ha
# kipróbáltad. Egy soha ki nem próbált mentés nem mentés, csak remény.
#
# Használat:
#   ./deploy/scripts/restore.sh /var/backups/snitt/keycloak-20260910-030000.sql.gz keycloak

set -euo pipefail

FILE="${1:?add meg a mentés fájlját}"
DB="${2:?add meg az adatbázis nevét (keycloak vagy snitt)}"
COMPOSE_FILE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/docker-compose.yml"

echo "FIGYELEM: a(z) '$DB' adatbázis jelenlegi tartalma felülíródik."
read -r -p "Folytatod? (igen/nem) " valasz
[ "$valasz" = "igen" ] || { echo "megszakítva"; exit 1; }

# A Keycloak és az account szolgáltatás ne írjon, amíg cserélünk alóluk.
docker compose -f "$COMPOSE_FILE" stop keycloak account >/dev/null

gunzip -c "$FILE" | docker compose -f "$COMPOSE_FILE" exec -T postgres \
  psql -U "${POSTGRES_USER:-snitt}" -d "$DB" -v ON_ERROR_STOP=1

docker compose -f "$COMPOSE_FILE" start keycloak account >/dev/null
echo "kész: $DB visszaállítva a(z) $FILE fájlból"
