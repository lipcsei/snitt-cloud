package config

import (
	"strings"
	"testing"
)

const testSecret = "0123456789abcdef0123456789abcdef0123456789abcdef"

// prodEnv egy érvényes éles környezet; a tesztek ebből rontanak el egy-egy kulcsot.
func prodEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SENTRY_ENVIRONMENT", "production")
	t.Setenv("DATABASE_URL", "postgres://snitt:"+testSecret+"@postgres:5432/snitt?sslmode=disable")
	t.Setenv("KEYCLOAK_ISSUER", "https://auth.example.com/realms/snitt")
	t.Setenv("CORS_ORIGINS", "https://example.com,https://admin.example.com")
	t.Setenv("KEYCLOAK_ADMIN_CLIENT_SECRET", testSecret)
}

func TestLoadDevelopmentDefaults(t *testing.T) {
	t.Setenv("SENTRY_ENVIRONMENT", "")
	if _, err := Load(); err != nil {
		t.Fatalf("a fejlesztői alapértékekkel indulnia kell: %v", err)
	}
}

func TestLoadProduction(t *testing.T) {
	prodEnv(t)
	if _, err := Load(); err != nil {
		t.Fatalf("érvényes éles környezet: %v", err)
	}
}

func TestLoadProductionRefusesDevValues(t *testing.T) {
	cases := []struct{ name, key, value, want string }{
		{"hiányzó titok (fejlesztői alapérték)", "KEYCLOAK_ADMIN_CLIENT_SECRET", "", "KEYCLOAK_ADMIN_CLIENT_SECRET"},
		{"fejlesztői titok", "KEYCLOAK_ADMIN_CLIENT_SECRET", devAdminClientSecret, "KEYCLOAK_ADMIN_CLIENT_SECRET"},
		{"rövid titok", "KEYCLOAK_ADMIN_CLIENT_SECRET", "tooshort", "KEYCLOAK_ADMIN_CLIENT_SECRET"},
		{"hiányzó DSN (fejlesztői alapérték)", "DATABASE_URL", "", "DATABASE_URL"},
		{"fejlesztői adatbázis-jelszó", "DATABASE_URL", "postgres://snitt:snitt@postgres:5432/snitt", "DATABASE_URL"},
		{"http kibocsátó", "KEYCLOAK_ISSUER", "http://auth.example.com/realms/snitt", "KEYCLOAK_ISSUER"},
		{"csillag CORS", "CORS_ORIGINS", "*", "CORS_ORIGINS"},
		{"http CORS-eredet", "CORS_ORIGINS", "https://example.com,http://localhost:5174", "CORS_ORIGINS"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prodEnv(t)
			t.Setenv(tc.key, tc.value)
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("hibát vártunk %s említésével, ezt kaptuk: %v", tc.want, err)
			}
		})
	}
}

func TestStripeAIPricesAreOptional(t *testing.T) {
	t.Setenv("SENTRY_ENVIRONMENT", "")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec")
	t.Setenv("STRIPE_PRICE_MONTHLY", "price_m")
	t.Setenv("STRIPE_PRICE_LIFETIME", "price_l")
	t.Setenv("STRIPE_PRICE_AI_MONTHLY", "")
	t.Setenv("STRIPE_PRICE_AI_TOKENS", "")
	c, err := Load()
	if err != nil {
		t.Fatalf("az AI árak nélkül is indulnia kell: %v", err)
	}
	if c.AIEmbedModel != "voyage-4" || c.AIAnswerModel != "claude-opus-5-5" {
		t.Errorf("AI modellek alapértéke = %q / %q", c.AIEmbedModel, c.AIAnswerModel)
	}

	t.Setenv("STRIPE_PRICE_LIFETIME", "")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "STRIPE_PRICE_LIFETIME") {
		t.Errorf("a lifetime ára továbbra is kötelező: %v", err)
	}
}
