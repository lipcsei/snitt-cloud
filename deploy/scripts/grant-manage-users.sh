#!/usr/bin/env bash
# A snitt-admin-api service accountnak megadja a realm-management
# "manage-users" szerepét.
#
# MIÉRT KELL: az admin felület fiók letiltása/engedélyezése művelete a
# Keycloakba ír. A view-users önmagában csak olvasásra jó, írásra 403 jön,
# amit az account szolgáltatás 502-ként ad tovább.
#
# MIÉRT KÜLÖN SCRIPT: a keycloak/realm-export.json már tartalmazza ezt a
# szerepet, de a realm import CSAK AKKOR fut le, ha a realm még nem létezik.
# Egy már működő Keycloakon tehát az export módosítása önmagában nem hat -
# ez a script pótolja azt az egy lépést. Új telepítésnél nincs rá szükség.
#
# Használat (a repo gyökeréből):
#   ./deploy/scripts/grant-manage-users.sh
#
# A Keycloak bootstrap admin jelszavát a KEYCLOAK_ADMIN_PASSWORD környezeti
# változóból veszi; ha nincs beállítva, a kcadm interaktívan bekéri.
# A script kétszer futtatva sem árt: a Keycloak a már meglévő szerepet
# egyszerűen nem adja hozzá újra.

set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-deploy/docker-compose.yml}"
REALM="${KEYCLOAK_REALM:-snitt}"
ADMIN_USER="${KEYCLOAK_ADMIN:-admin}"
CLIENT_ID="${KEYCLOAK_ADMIN_CLIENT_ID:-snitt-admin-api}"
KC_URL="${KEYCLOAK_INTERNAL_URL:-http://localhost:8081}"

kc() { docker compose -f "$COMPOSE_FILE" exec -T keycloak /opt/keycloak/bin/kcadm.sh "$@"; }

echo "Bejelentkezés a Keycloak admin felületére ($KC_URL, felhasználó: $ADMIN_USER)…"
if [[ -n "${KEYCLOAK_ADMIN_PASSWORD:-}" ]]; then
  kc config credentials --server "$KC_URL" --realm master \
    --user "$ADMIN_USER" --password "$KEYCLOAK_ADMIN_PASSWORD"
else
  # Jelszó nélkül a kcadm maga kérdez rá - a jelszó így nem kerül a
  # shell előzményébe és a folyamatlistába sem.
  docker compose -f "$COMPOSE_FILE" exec keycloak /opt/keycloak/bin/kcadm.sh \
    config credentials --server "$KC_URL" --realm master --user "$ADMIN_USER"
fi

echo "A(z) $CLIENT_ID service accountjának manage-users szerep adása…"
kc add-roles -r "$REALM" \
  --uusername "service-account-${CLIENT_ID}" \
  --cclientid realm-management \
  --rolename manage-users

echo
echo "Kész. Ellenőrzés - a service account realm-management szerepei:"
kc get-roles -r "$REALM" \
  --uusername "service-account-${CLIENT_ID}" \
  --cclientid realm-management --fields name

echo
echo "Az account szolgáltatás a tokent gyorsítótárazza; a szerep néhány percen"
echo "belül, vagy a konténer újraindítása után lép életbe:"
echo "  docker compose -f $COMPOSE_FILE restart account"
