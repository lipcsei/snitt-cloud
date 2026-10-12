// Package config a szolgáltatás környezeti változókból olvasott beállításait tartalmazza.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config a futáshoz szükséges összes beállítás. Minden mező környezeti
// változóból jön, a fejlesztői stackhez illő alapértelmezésekkel.
type Config struct {
	HTTPAddr         string
	DatabaseURL      string
	KeycloakIssuer   string
	KeycloakDiscoURL string
	KeycloakAudience string
	CORSOrigins      []string

	// Az admin felület a Keycloak Admin REST API-n keresztül olvassa a
	// felhasználókat, egy service accountos, bizalmas klienssel.
	KeycloakBaseURL   string
	KeycloakRealm     string
	AdminClientID     string
	AdminClientSecret string
	AdminRole         string
	DefaultCurrency   string

	// Opcionális hibajelentés egy Sentry-kompatibilis szerverre (a saját
	// GlitchTipünkre). Üres SentryDSN = teljesen kikapcsolva.
	SentryDSN         string
	SentryEnvironment string
	SentryRelease     string
}

// Load beolvassa a konfigurációt a környezetből és validálja.
func Load() (Config, error) {
	c := Config{
		HTTPAddr:       env("HTTP_ADDR", ":8090"),
		DatabaseURL:    env("DATABASE_URL", "postgres://snitt:snitt@localhost:5432/snitt?sslmode=disable"),
		KeycloakIssuer: env("KEYCLOAK_ISSUER", "http://localhost:8081/realms/snitt"),
		// Vesszővel elválasztva több elfogadott audience is megadható: a
		// landing, az admin felület és később a desktop app is a saját
		// client id-jével kap tokent ugyanehhez az API-hoz.
		KeycloakAudience: env("KEYCLOAK_AUDIENCE", "snitt-landing,snitt-admin"),
		CORSOrigins:      splitList(env("CORS_ORIGINS", "http://localhost:5174,http://localhost:5175")),

		KeycloakBaseURL:   strings.TrimRight(env("KEYCLOAK_BASE_URL", "http://localhost:8081"), "/"),
		KeycloakRealm:     env("KEYCLOAK_REALM", "snitt"),
		AdminClientID:     env("KEYCLOAK_ADMIN_CLIENT_ID", "snitt-admin-api"),
		AdminClientSecret: env("KEYCLOAK_ADMIN_CLIENT_SECRET", "snitt-admin-api-dev-secret"),
		AdminRole:         env("ADMIN_ROLE", "admin"),
		DefaultCurrency:   strings.ToUpper(env("DEFAULT_CURRENCY", "HUF")),

		SentryDSN:         env("SENTRY_DSN", ""),
		SentryEnvironment: env("SENTRY_ENVIRONMENT", "development"),
		SentryRelease:     env("SENTRY_RELEASE", ""),
	}
	// Containeren belül a Keycloak más hosztnéven érhető el ("keycloak"), mint
	// ahogy a tokenek issuer claimje szól ("localhost"). Alapból a kettő azonos.
	c.KeycloakDiscoURL = env("KEYCLOAK_DISCOVERY_URL", c.KeycloakIssuer)

	if c.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL kötelező")
	}
	if c.KeycloakIssuer == "" {
		return Config{}, fmt.Errorf("KEYCLOAK_ISSUER kötelező")
	}
	if c.AdminRole == "" {
		return Config{}, fmt.Errorf("ADMIN_ROLE nem lehet üres")
	}
	if len(c.DefaultCurrency) != 3 {
		return Config{}, fmt.Errorf("DEFAULT_CURRENCY három betűs ISO kód kell legyen")
	}
	if c.SentryEnvironment == "production" {
		if err := c.checkProduction(); err != nil {
			return Config{}, err
		}
	}
	return c, nil
}

// A fejlesztői alapértékek ebben a PUBLIKUS repóban bárki számára olvashatók.
const (
	devAdminClientSecret = "snitt-admin-api-dev-secret"
	devDatabaseUserinfo  = "snitt:snitt@"
)

// checkProduction megakadályozza, hogy az éles szolgáltatás egy kimaradt
// környezeti változó miatt csendben a fejlesztői alapértékkel induljon el. A
// service account manage-users joggal bír, a titka tehát a realm összes
// felhasználójához hozzáférést ad.
func (c Config) checkProduction() error {
	if c.AdminClientSecret == devAdminClientSecret || len(c.AdminClientSecret) < 32 {
		return fmt.Errorf("KEYCLOAK_ADMIN_CLIENT_SECRET élesben legalább 32 karakteres, generált titok kell legyen, nem a fejlesztői érték")
	}
	if strings.Contains(c.DatabaseURL, devDatabaseUserinfo) {
		return fmt.Errorf("DATABASE_URL élesben nem használhatja a fejlesztői adatbázis-jelszót")
	}
	if !strings.HasPrefix(c.KeycloakIssuer, "https://") {
		return fmt.Errorf("KEYCLOAK_ISSUER élesben https:// cím kell legyen")
	}
	for _, o := range c.CORSOrigins {
		if !strings.HasPrefix(o, "https://") {
			return fmt.Errorf("CORS_ORIGINS élesben csak https:// eredeteket tartalmazhat (kapott: %q)", o)
		}
	}
	return nil
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
