package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

const testIssuer = "https://auth.example.test/realms/snitt"

// testRealm saját kulccsal aláírt tokeneket ad, és a hozzá tartozó, a valódi Verify kódot futtató
// ellenőrzőt - Keycloak nélkül.
type testRealm struct {
	signer   jose.Signer
	verifier *OIDCVerifier
}

func newTestRealm(t *testing.T, audiences ...string) testRealm {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		t.Fatal(err)
	}
	keys := &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&key.PublicKey}}
	return testRealm{
		signer: signer,
		verifier: &OIDCVerifier{
			verifier:  oidc.NewVerifier(testIssuer, keys, &oidc.Config{SkipClientIDCheck: true}),
			audiences: audiences,
		},
	}
}

func (r testRealm) token(t *testing.T, typ string, audience ...string) string {
	t.Helper()
	extra := map[string]any{"preferred_username": "anna"}
	if typ != "" {
		extra["typ"] = typ
	}
	raw, err := jwt.Signed(r.signer).Claims(jwt.Claims{
		Issuer: testIssuer, Subject: "user-1", Audience: audience,
		Expiry: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
	}).Claims(extra).Serialize()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestVerifyAcceptsAnAccessTokenForAnAllowedAudience(t *testing.T) {
	r := newTestRealm(t, "snitt-landing", "snitt-admin")
	id, err := r.verifier.Verify(context.Background(), r.token(t, "Bearer", "snitt-admin"))
	if err != nil {
		t.Fatalf("érvényes tokent utasított el: %v", err)
	}
	if id.Subject != "user-1" || id.PreferredUsername != "anna" {
		t.Fatalf("identity = %+v", id)
	}
}

// Az ID token audience-e maga a kliens, tehát az audience-ellenőrzésen ÁTMEGY - a típusa buktatja le.
func TestVerifyRefusesTokensThatAreNotAccessTokens(t *testing.T) {
	r := newTestRealm(t, "snitt-landing", "snitt-admin")
	for name, typ := range map[string]string{"ID token": "ID", "refresh token": "Refresh", "típus nélkül": ""} {
		t.Run(name, func(t *testing.T) {
			_, err := r.verifier.Verify(context.Background(), r.token(t, typ, "snitt-landing"))
			if !errors.Is(err, ErrUnauthenticated) {
				t.Fatalf("err = %v, ErrUnauthenticated kellene", err)
			}
		})
	}
}

func TestVerifyRefusesAForeignAudience(t *testing.T) {
	r := newTestRealm(t, "snitt-landing", "snitt-admin")
	_, err := r.verifier.Verify(context.Background(), r.token(t, "Bearer", "account"))
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("err = %v, ErrUnauthenticated kellene", err)
	}
}
