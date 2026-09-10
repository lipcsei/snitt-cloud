// A snitt account szolgáltatás belépési pontja.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/billing"
	"github.com/lipcsei/snitt-cloud/services/account/internal/config"
	"github.com/lipcsei/snitt-cloud/services/account/internal/httpapi"
	"github.com/lipcsei/snitt-cloud/services/account/internal/keycloak"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("a szolgáltatás leállt hibával", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startupCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	st, err := store.New(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	if err := st.Migrate(startupCtx); err != nil {
		return err
	}
	log.Info("séma alkalmazva")

	// A Keycloak a compose stackben lassabban indul, mint ez a szolgáltatás,
	// ezért a discoveryt újrapróbáljuk, ahelyett hogy azonnal elszállnánk.
	verifier, err := discoverWithRetry(startupCtx, log, cfg)
	if err != nil {
		return err
	}
	log.Info("OIDC verifier kész",
		"issuer", cfg.KeycloakIssuer,
		"discovery_url", cfg.KeycloakDiscoURL,
		"audience", cfg.KeycloakAudience)

	// Az admin felület a Keycloak Admin API-n keresztül olvassa a
	// felhasználókat. A kliens lustán kér tokent, ezért ez nem hálózati hívás.
	directory := keycloak.NewAdminClient(cfg.KeycloakBaseURL, cfg.KeycloakRealm,
		cfg.AdminClientID, cfg.AdminClientSecret)
	log.Info("keycloak admin kliens kész",
		"base_url", cfg.KeycloakBaseURL, "realm", cfg.KeycloakRealm, "client_id", cfg.AdminClientID)

	api := httpapi.New(httpapi.Options{
		Store:     st,
		Verifier:  verifier,
		Directory: directory,
		// Valódi fizetési szolgáltató még nincs bekötve: minden előfizetés és
		// számla kézi admin döntés. Lásd internal/billing.
		Billing:         billing.Manual{},
		CORSOrigins:     cfg.CORSOrigins,
		AdminRole:       cfg.AdminRole,
		DefaultCurrency: cfg.DefaultCurrency,
		Log:             log,
	})
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("HTTP szerver indul", "addr", cfg.HTTPAddr, "cors_origins", cfg.CORSOrigins)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("leállítási jelzés, futó kérések bevárása")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()
	return srv.Shutdown(shutdownCtx)
}

func discoverWithRetry(ctx context.Context, log *slog.Logger, cfg config.Config) (auth.TokenVerifier, error) {
	const maxAttempts = 12
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		v, err := auth.NewOIDCVerifier(ctx, cfg.KeycloakIssuer, cfg.KeycloakDiscoURL, cfg.KeycloakAudience)
		if err == nil {
			return v, nil
		}
		lastErr = err
		log.Warn("OIDC discovery sikertelen, újrapróbálás", "attempt", attempt, "error", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	return nil, lastErr
}
