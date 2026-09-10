// Package auth a Keycloak által kiállított JWT-k ellenőrzését és a
// hívó azonosítójának a request contextbe helyezését végzi.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// Identity a tokenből kiolvasott, Keycloak által birtokolt felhasználói adatok.
// A Subject a stabil kulcs, minden más mező tájékoztató jellegű.
type Identity struct {
	Subject           string      `json:"sub"`
	Email             string      `json:"email"`
	PreferredUsername string      `json:"preferred_username"`
	Name              string      `json:"name"`
	RealmAccess       RealmAccess `json:"realm_access"`
}

// RealmAccess a Keycloak realm-szintű szerepeit tartalmazó token claim.
type RealmAccess struct {
	Roles []string `json:"roles"`
}

// HasRealmRole megmondja, hogy a token hordozza-e az adott realm szerepet.
func (i Identity) HasRealmRole(role string) bool {
	for _, r := range i.RealmAccess.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// TokenVerifier absztrakció a JWT ellenőrzés fölött, hogy a HTTP réteg
// tesztelhető legyen élő Keycloak nélkül.
type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (Identity, error)
}

type ctxKey struct{}

// ErrUnauthenticated minden olyan hibát jelöl, amire 401 a helyes válasz.
var ErrUnauthenticated = errors.New("hitelesítés sikertelen")

// OIDCVerifier OIDC discovery alapján ellenőrzi a Keycloak tokeneket.
type OIDCVerifier struct {
	verifier *oidc.IDTokenVerifier
}

// NewOIDCVerifier lekéri a discovery dokumentumot és felépíti a JWKS-alapú
// ellenőrzőt. Hálózati hívás, ezért indulásnál hívandó.
//
// A discoveryURL eltérhet az issuertől: Dockerben a szolgáltatás a
// "keycloak" hoszton éri el a Keycloakot, a böngészőnek kiadott tokenek
// issuer claimje viszont "localhost". A tokenek ellenőrzése ilyenkor is az
// issuer paraméterhez igazodik.
func NewOIDCVerifier(ctx context.Context, issuer, discoveryURL, audience string) (*OIDCVerifier, error) {
	if discoveryURL == "" {
		discoveryURL = issuer
	}
	if discoveryURL != issuer {
		ctx = oidc.InsecureIssuerURLContext(ctx, issuer)
	}
	provider, err := oidc.NewProvider(ctx, discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery (%s): %w", discoveryURL, err)
	}
	cfg := &oidc.Config{ClientID: audience}
	if audience == "" {
		cfg.SkipClientIDCheck = true
	}
	return &OIDCVerifier{verifier: provider.Verifier(cfg)}, nil
}

// Verify ellenőrzi az aláírást, az issuert, a lejáratot és az audience-t.
func (v *OIDCVerifier) Verify(ctx context.Context, rawToken string) (Identity, error) {
	tok, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return Identity{}, fmt.Errorf("%w: %v", ErrUnauthenticated, err)
	}
	var id Identity
	if err := tok.Claims(&id); err != nil {
		return Identity{}, fmt.Errorf("%w: claimek kiolvasása: %v", ErrUnauthenticated, err)
	}
	if id.Subject == "" {
		// A go-oidc a Subject claimet külön is kiadja, de a struct-tag alapú
		// dekódolásnak is meg kell találnia; ha nincs, a token használhatatlan.
		id.Subject = tok.Subject
	}
	if id.Subject == "" {
		return Identity{}, fmt.Errorf("%w: hiányzó sub claim", ErrUnauthenticated)
	}
	return id, nil
}

// Middleware Bearer tokent vár, ellenőrizteti, majd az Identity-t a
// contextbe teszi. Hiba esetén 401-gyel rövidre zár.
func Middleware(v TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r)
			if !ok {
				writeUnauthorized(w, "hiányzó vagy hibás Authorization fejléc")
				return
			}
			id, err := v.Verify(r.Context(), raw)
			if err != nil {
				writeUnauthorized(w, "érvénytelen token")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithIdentity(r.Context(), id)))
		})
	}
}

// RequireRealmRole a Middleware fölé rétegződik: feltételezi, hogy a token
// már ellenőrzött, és csak a jogosultságot nézi. Külön második hitelesítési
// mechanizmus (admin API kulcs, alap-hitelesítés) szándékosan nincs.
func RequireRealmRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := FromContext(r.Context())
			if !ok {
				writeUnauthorized(w, "hiányzó identitás")
				return
			}
			if !id.HasRealmRole(role) {
				writeForbidden(w, "ehhez a művelethez "+role+" jogosultság kell")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// WithIdentity az Identity-t a contexthez fűzi.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext visszaadja a middleware által elhelyezett Identity-t.
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}

func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	const prefix = "bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}
	tok := strings.TrimSpace(h[len(prefix):])
	return tok, tok != ""
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	// A WWW-Authenticate fejléc nélkül a böngésző oldali kliensek nehezebben
	// különböztetik meg a lejárt tokent az egyéb 401-től.
	w.Header().Set("WWW-Authenticate", `Bearer realm="snitt"`)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeForbidden(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
