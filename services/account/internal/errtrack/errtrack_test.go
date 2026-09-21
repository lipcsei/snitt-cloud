package errtrack

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getsentry/sentry-go"
)

// A Sentry hubja folyamat-szintű, ezért ezek a tesztek nem futhatnak párhuzamosan.

func TestSetupEmptyDSNIsDisabled(t *testing.T) {
	var logs bytes.Buffer
	flush := Setup(Config{Environment: "test"}, slog.New(slog.NewTextHandler(&logs, nil)))
	if flush == nil {
		t.Fatal("a flush függvény nil")
	}
	defer flush()

	if enabled() {
		t.Fatal("üres DSN mellett bekapcsolt a hibajelentés")
	}
	flush() // kikapcsolt módban nem csinál semmit, és nem pánikol
	if strings.Contains(logs.String(), "bekapcsolva") {
		t.Errorf("üres DSN mellett bekapcsolást naplózott: %s", logs.String())
	}
}

func TestDisabledMiddlewareAndReportAreTransparent(t *testing.T) {
	flush := Setup(Config{}, nil)
	defer flush()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sentry.GetHubFromContext(r.Context()) != nil {
			t.Error("kikapcsolt módban is került hub a kérés contextjébe")
		}
		Report(r, errors.New("boom"), http.StatusInternalServerError) // nem csinálhat semmit
		w.Header().Set("X-Teszt", "igen")
		w.WriteHeader(http.StatusTeapot)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	Middleware(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot || rec.Header().Get("X-Teszt") != "igen" {
		t.Errorf("a kikapcsolt middleware módosította a választ: %d %v", rec.Code, rec.Header())
	}
}

func TestSetupInvalidDSNStaysDisabled(t *testing.T) {
	const bad = "ez-nem-dsn-titkoskulcs"
	var logs bytes.Buffer
	flush := Setup(Config{DSN: bad}, slog.New(slog.NewTextHandler(&logs, nil)))
	defer flush()

	if enabled() {
		t.Fatal("érvénytelen DSN mellett bekapcsolt a hibajelentés")
	}
	if !strings.Contains(logs.String(), "érvénytelen") {
		t.Errorf("az érvénytelen DSN nem okozott figyelmeztetést: %q", logs.String())
	}
	if strings.Contains(logs.String(), bad) {
		t.Errorf("a napló a nyers DSN-t tartalmazza: %q", logs.String())
	}
}

func TestReportSkipsNilError(t *testing.T) {
	tr := &sentry.MockTransport{}
	flush := Setup(Config{DSN: "https://public@glitchtip.test/1", Transport: tr}, nil)
	defer flush()

	Report(httptest.NewRequest(http.MethodGet, "/x", nil), nil, http.StatusInternalServerError)
	if n := len(tr.Events()); n != 0 {
		t.Errorf("események = %d, várt 0", n)
	}
}

func TestFlushDisablesReporting(t *testing.T) {
	tr := &sentry.MockTransport{}
	flush := Setup(Config{DSN: "https://public@glitchtip.test/1", Transport: tr}, nil)
	if !enabled() {
		t.Fatal("érvényes DSN mellett nem kapcsolt be a hibajelentés")
	}
	flush()
	if enabled() {
		t.Error("a flush után is be van kapcsolva a hibajelentés")
	}
}
