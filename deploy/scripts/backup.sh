#!/usr/bin/env bash
# A Snitt felhő oldalának mentése.
#
# MIT ment: a Postgres két adatbázisát - a Keycloakét (felhasználók, jelszó-
# lenyomatok, kliensbeállítások) és az account szolgáltatásét (profilok,
# előfizetések, számlák).
#
# MIÉRT pont ezt: ez az EGYETLEN adat az egész rendszerben, ami nem
# újratermelhető. A landing statikus, a konténerképek újraépíthetők, a
# felhasználók videói és snittjei pedig a SAJÁT gépükön vannak - azokhoz
# sosem nyúlunk. Ha a VPS holnap elveszik, csak ez a néhány megabájt hiányzik.
#
# Használat (a repo gyökeréből vagy cronból):
#   ./deploy/scripts/backup.sh /celkonyvtar
#
# Cron példa - hajnali 3-kor, naponta:
#   0 3 * * * /opt/snitt/deploy/scripts/backup.sh /var/backups/snitt >> /var/log/snitt-backup.log 2>&1

set -euo pipefail

DEST="${1:-/var/backups/snitt}"
COMPOSE_FILE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/docker-compose.yml"
KEEP_DAYS="${SNITT_BACKUP_KEEP_DAYS:-30}"
STAMP="$(date +%Y%m%d-%H%M%S)"

mkdir -p "$DEST"

dump() {
  local db="$1"
  local out="$DEST/${db}-${STAMP}.sql.gz"
  echo "mentés: $db -> $out"
  # A -Fp (sima SQL) szándékos: bármelyik Postgres verzióval visszaolvasható,
  # és a tartalma megnézhető anélkül, hogy vissza kellene állítani.
  docker compose -f "$COMPOSE_FILE" exec -T postgres \
    pg_dump -U "${POSTGRES_USER:-snitt}" --clean --if-exists "$db" | gzip -9 > "$out.part"
  # Csak a KÉSZ fájlt nevezzük át: egy megszakadt mentés ne tűnjön jónak.
  mv "$out.part" "$out"
  echo "  méret: $(du -h "$out" | cut -f1)"
}

dump "${KEYCLOAK_DB:-keycloak}"
dump "${POSTGRES_DB:-snitt}"

# A régi mentések takarítása. Ha ez elmarad, a lemez telik meg - és akkor a
# mentés is elhasal, pont amikor a legjobban kellene.
find "$DEST" -name '*.sql.gz' -type f -mtime +"$KEEP_DAYS" -print -delete

echo "kész: $(ls -1 "$DEST"/*.sql.gz 2>/dev/null | wc -l | tr -d ' ') mentés a célkönyvtárban"
