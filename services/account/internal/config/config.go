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
}

// Load beolvassa a konfigurációt a környezetből és validálja.
func Load() (Config, error) {
	c := Config{
		HTTPAddr:         env("HTTP_ADDR", ":8090"),
		DatabaseURL:      env("DATABASE_URL", "postgres://snitt:snitt@localhost:5432/snitt?sslmode=disable"),
		KeycloakIssuer:   env("KEYCLOAK_ISSUER", "http://localhost:8081/realms/snitt"),
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
	return c, nil
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
