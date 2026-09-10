package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/billing"
	"github.com/lipcsei/snitt-cloud/services/account/internal/keycloak"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// A jogosultság-kulcs szándékosan szabad szöveg: a csomagokon kívüli,
// kézzel adott kulcsokat (béta-hozzáférés, egyedi megállapodás) éppen ez
// teszi lehetővé. A minta csak azt zárja ki, ami biztosan elgépelés.
var featureKeyRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9._-]{0,62}[a-z0-9])?$`)

// maxEntitlementDays a kézi jogosultság leghosszabb futamideje. Nem
// korlátozás kedvéért van: a "soha nem jár le" külön, szándékos választás,
// nem elgépelt évszám mellékhatása.
const maxEntitlementDays = 3650

// ---------- jogosultságok ----------

type grantEntitlementRequest struct {
	FeatureKey string `json:"feature_key"`
	// Days nil értéke: a jogosultság nem jár le.
	Days *int   `json:"days"`
	Note string `json:"note"`
}

func (a *API) handleGrantEntitlement(w http.ResponseWriter, r *http.Request) {
	subject := r.PathValue("subject")

	var req grantEntitlementRequest
	if !decodeBody(w, r, &req) {
		return
	}
	key := strings.TrimSpace(req.FeatureKey)
	if !featureKeyRE.MatchString(key) {
		writeError(w, http.StatusBadRequest,
			"a jogosultság kulcsa csak kisbetűt, számot, pontot, kötőjelet és aláhúzást tartalmazhat")
		return
	}

	var expires *time.Time
	if req.Days != nil {
		if *req.Days < 1 || *req.Days > maxEntitlementDays {
			writeError(w, http.StatusBadRequest,
				fmt.Sprintf("a futamidő 1 és %d nap között lehet; lejárat nélküli jogosultsághoz hagyd üresen", maxEntitlementDays))
			return
		}
		t := time.Now().UTC().Add(time.Duration(*req.Days) * 24 * time.Hour)
		expires = &t
	}

	ent, err := a.store.GrantEntitlement(r.Context(), subject, key, expires)
	if err != nil {
		a.writeStoreError(w, r, err, "a jogosultság nem adható ki")
		return
	}

	until := "lejárat nélkül"
	if expires != nil {
		until = "eddig: " + expires.Format(dateLayout)
	}
	a.audit(r, store.NewAuditEntry{
		Action:     store.AuditEntitlementGrant,
		TargetType: "entitlement",
		TargetID:   key,
		Subject:    subject,
		Summary:    fmt.Sprintf("Jogosultság kiadva: %s (%s).", key, until),
		Detail: map[string]any{
			"feature_key": key, "expires_at": expires, "note": strings.TrimSpace(req.Note),
		},
	})

	resp := map[string]any{"entitlement": ent}
	// A csomagok által vezérelt kulcsokat a következő előfizetés-művelet
	// felülírja. Ezt az adminnak tudnia kell, mielőtt erre épít.
	if slices.Contains(billing.ManagedFeatureKeys(), key) {
		resp["warning"] = "Ez a kulcs csomaghoz tartozik: a felhasználó következő előfizetés-műveletekor (kiadás, csomagváltás, lemondás) felülíródik."
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (a *API) handleRevokeEntitlement(w http.ResponseWriter, r *http.Request) {
	subject, key := r.PathValue("subject"), r.PathValue("key")

	ent, err := a.store.RevokeEntitlement(r.Context(), subject, key)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ez a jogosultság nincs kiadva ennek a felhasználónak")
		return
	}
	if err != nil {
		a.writeStoreError(w, r, err, "a jogosultság nem vonható vissza")
		return
	}

	a.audit(r, store.NewAuditEntry{
		Action:     store.AuditEntitlementRevoke,
		TargetType: "entitlement",
		TargetID:   key,
		Subject:    subject,
		Summary:    fmt.Sprintf("Jogosultság visszavonva: %s.", key),
		Detail: map[string]any{
			"feature_key": key, "granted_at": ent.GrantedAt, "expires_at": ent.ExpiresAt,
		},
	})
	writeJSON(w, http.StatusOK, map[string]any{"entitlement": ent})
}

// ---------- fiók engedélyezése / letiltása ----------

type patchUserRequest struct {
	Enabled *bool  `json:"enabled"`
	Note    string `json:"note"`
}

func (a *API) handlePatchUser(w http.ResponseWriter, r *http.Request) {
	subject := r.PathValue("subject")

	var req patchUserRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Enabled == nil {
		writeError(w, http.StatusBadRequest, "az enabled mező kötelező")
		return
	}

	// Előbb megnézzük, mi az állapot: így a napló a tényleges változást írja,
	// és a nem létező felhasználó 404-et kap, nem a Keycloak nyers hibáját.
	ku, err := a.directory.GetUser(r.Context(), subject)
	if errors.Is(err, keycloak.ErrNotFound) {
		writeError(w, http.StatusNotFound, "nincs ilyen felhasználó")
		return
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "keycloak felhasználó lekérdezés sikertelen", "error", err)
		writeError(w, http.StatusBadGateway, "a Keycloak nem elérhető: "+err.Error())
		return
	}
	if ku.Enabled == *req.Enabled {
		// Nincs mit tenni: ne írjunk félrevezető naplósort sem.
		writeJSON(w, http.StatusOK, map[string]any{"enabled": ku.Enabled, "changed": false})
		return
	}

	if err := a.directory.SetUserEnabled(r.Context(), subject, *req.Enabled); err != nil {
		a.log.ErrorContext(r.Context(), "keycloak fiókállapot írása sikertelen", "error", err, "subject", subject)
		writeError(w, http.StatusBadGateway, "a Keycloak nem fogadta el a módosítást: "+err.Error())
		return
	}

	action, summary := store.AuditUserEnable, "Fiók engedélyezve."
	if !*req.Enabled {
		action, summary = store.AuditUserDisable, "Fiók letiltva: a felhasználó nem tud belépni."
	}
	a.audit(r, store.NewAuditEntry{
		Action:     action,
		TargetType: "user",
		TargetID:   subject,
		Subject:    subject,
		Summary:    summary,
		Detail: map[string]any{
			"enabled": *req.Enabled, "username": ku.Username, "email": ku.Email,
			"note": strings.TrimSpace(req.Note),
		},
	})
	writeJSON(w, http.StatusOK, map[string]any{"enabled": *req.Enabled, "changed": true})
}
