package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/getsentry/sentry-go"

	"github.com/lipcsei/commons/errtrack"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// A Sentry hubja folyamat-szintű, ezért az ezt használó tesztek nem futhatnak párhuzamosan
// (a csomag tesztjei egyébként sem használnak t.Parallel-t).

// installErrtrack bekapcsolja a hibajelentést a teszt idejére egy memóriabeli szállítóval, hálózat
// nélkül; a teszt végén kikapcsolja.
func installErrtrack(t *testing.T) *sentry.MockTransport {
	t.Helper()
	tr := &sentry.MockTransport{}
	flush := errtrack.Setup(errtrack.Config{
		DSN:         "https://public@glitchtip.test/1",
		Environment: "test",
		Transport:   tr,
	}, nil)
	t.Cleanup(flush)
	return tr
}

// panicStore az UpsertProfile-ban (GET /api/v1/me) pánikol, minden más a fakeStore-é. A value nil
// esetén valódi runtime hiba (nil pointer) a pánik, különben maga az érték.
type panicStore struct {
	*fakeStore
	value any
}

func (p panicStore) UpsertProfile(context.Context, string, string, string, string) (store.Profile, error) {
	if p.value != nil {
		panic(p.value)
	}
	var missing *store.Profile
	return *missing, nil
}

func TestErrtrackDisabledIsNoop(t *testing.T) {
	// SENTRY_DSN nélkül nincs kliens, és a szolgáltatás úgy viselkedik, mintha a csomag nem is
	// létezne: a hibaválasz és a pánik-viselkedés változatlan.
	flush := errtrack.Setup(errtrack.Config{}, nil)
	defer flush()
	if sentry.CurrentHub().Client() != nil {
		t.Fatal("üres DSN mellett is van Sentry-kliens")
	}

	h := newTestAPI(&fakeStore{upsertErr: errors.New("boom")})
	if rec := do(t, h, http.MethodGet, "/api/v1/me", "good-token", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("státusz = %d, várt 500", rec.Code)
	}

	defer func() {
		if r := recover(); r != "kaboom" {
			t.Errorf("recover() = %v, várt az eredeti pánik", r)
		}
	}()
	do(t, newTestAPI(panicStore{fakeStore: &fakeStore{}, value: "kaboom"}), http.MethodGet, "/api/v1/me", "good-token", "")
	t.Error("a pánik nem jutott el a hívóig")
}

func TestErrtrackReportsEach5xxOnce(t *testing.T) {
	tests := []struct {
		name       string
		st         Store
		dir        UserDirectory
		method     string
		path       string
		token      string
		body       string
		wantStatus int
		wantRoute  string
		wantErr    string
	}{
		{
			name:       "adatbázis-hiba: 500",
			st:         &fakeStore{upsertErr: errors.New("db: connection refused")},
			dir:        &fakeDirectory{},
			method:     http.MethodGet,
			path:       "/api/v1/me",
			token:      "good-token",
			wantStatus: http.StatusInternalServerError,
			wantRoute:  "GET /api/v1/me",
			wantErr:    "db: connection refused",
		},
		{
			name:       "Keycloak nem elérhető: 502",
			st:         &fakeStore{},
			dir:        &fakeDirectory{listErr: errors.New("keycloak: timeout")},
			method:     http.MethodGet,
			path:       "/api/v1/admin/users",
			token:      "admin-token",
			wantStatus: http.StatusBadGateway,
			wantRoute:  "GET /api/v1/admin/users",
			wantErr:    "keycloak: timeout",
		},
		{
			name:       "adatbázis-hiba a közös hibaágon (writeStoreError): 500",
			st:         &fakeStore{admin: adminFake{err: errors.New("db: deadlock")}},
			dir:        &fakeDirectory{},
			method:     http.MethodPost,
			path:       "/api/v1/admin/invoices/i1/pay",
			token:      "admin-token",
			body:       `{}`,
			wantStatus: http.StatusInternalServerError,
			wantRoute:  "POST /api/v1/admin/invoices/{id}/pay",
			wantErr:    "db: deadlock",
		},
		{
			name:       "healthz, adatbázis nem elérhető: 503",
			st:         &fakeStore{pingErr: errors.New("db: down")},
			dir:        &fakeDirectory{},
			method:     http.MethodGet,
			path:       "/healthz",
			wantStatus: http.StatusServiceUnavailable,
			wantRoute:  "GET /healthz",
			wantErr:    "db: down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := installErrtrack(t)
			rec := do(t, newTestAPIWith(tt.st, tt.dir), tt.method, tt.path, tt.token, tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			events := tr.Events()
			if len(events) != 1 {
				t.Fatalf("események = %d, várt pontosan 1", len(events))
			}
			ev := events[0]
			if len(ev.Exception) == 0 || ev.Exception[len(ev.Exception)-1].Value != tt.wantErr {
				t.Errorf("kivétel = %+v, várt érték: %q", ev.Exception, tt.wantErr)
			}
			if got := ev.Tags["http.route"]; got != tt.wantRoute {
				t.Errorf("http.route = %q, várt %q", got, tt.wantRoute)
			}
			if got, want := ev.Tags["http.status_code"], strconv.Itoa(tt.wantStatus); got != want {
				t.Errorf("http.status_code = %q, várt %q", got, want)
			}
			if ev.Environment != "test" {
				t.Errorf("environment = %q, várt test", ev.Environment)
			}
			if ev.Level != sentry.LevelError {
				t.Errorf("level = %q, várt error", ev.Level)
			}
		})
	}
}

func TestErrtrackIgnoresNon5xx(t *testing.T) {
	// A 4xx kliens-hiba, és a 2xx-szel visszatérő, de naplóba hibát író művelet sem jelentendő.
	tests := []struct {
		name       string
		st         *fakeStore
		method     string
		path       string
		token      string
		body       string
		wantStatus int
	}{
		{"ismeretlen útvonal: 404", &fakeStore{}, http.MethodGet, "/nincs-ilyen", "", "", http.StatusNotFound},
		{"nincs token: 401", &fakeStore{}, http.MethodGet, "/api/v1/me", "", "", http.StatusUnauthorized},
		{"nem admin: 403", &fakeStore{}, http.MethodGet, "/api/v1/admin/users", "good-token", "", http.StatusForbidden},
		{"hibás törzs: 400", &fakeStore{}, http.MethodPatch, "/api/v1/me", "good-token", `{"email":"x"}`, http.StatusBadRequest},
		{"nem létező rekord (writeStoreError): 404", &fakeStore{}, http.MethodPost, "/api/v1/admin/invoices/nincs/pay", "admin-token", `{}`, http.StatusNotFound},
		{"sikeres kérés: 200", &fakeStore{}, http.MethodGet, "/api/v1/me", "good-token", "", http.StatusOK},
		{
			"a naplóírás bukik, a művelet nem: 200",
			&fakeStore{admin: adminFake{invoices: openInvoice(), auditErr: errors.New("a napló nem elérhető")}},
			http.MethodPost, "/api/v1/admin/invoices/i1/pay", "admin-token", `{}`, http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := installErrtrack(t)
			rec := do(t, newTestAPI(tt.st), tt.method, tt.path, tt.token, tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if n := len(tr.Events()); n != 0 {
				t.Errorf("események = %d, várt 0", n)
			}
		})
	}
}

func TestErrtrackIgnoresCanceledRequest(t *testing.T) {
	// Ha a kliens megszakítja a kérést, az adatbázis-hívás context.Canceled-del tér vissza, és a
	// handler 500-at ír - ez nem a szolgáltatás hibája, nem jelentendő.
	tr := installErrtrack(t)
	h := newTestAPI(&fakeStore{upsertErr: context.Canceled})
	if rec := do(t, h, http.MethodGet, "/api/v1/me", "good-token", ""); rec.Code != http.StatusInternalServerError {
		t.Fatalf("státusz = %d, várt 500", rec.Code)
	}
	if n := len(tr.Events()); n != 0 {
		t.Errorf("események = %d, várt 0", n)
	}
}

func TestErrtrackPanicIsReportedOnceAndRepanicked(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string // a pánik szövege az eseményen
	}{
		{"szöveges pánik", "kaboom", "kaboom"},
		{"runtime hiba (nil pointer)", nil, "nil pointer dereference"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := installErrtrack(t)
			h := newTestAPI(panicStore{fakeStore: &fakeStore{}, value: tt.value})

			// Valódi net/http szerveren: a szerver saját pánik-kezelése (Repanic miatt) ugyanúgy
			// lezárja a kapcsolatot és tovább szolgál, mint a hibajelentés nélküli szolgáltatás.
			srv := httptest.NewUnstartedServer(h)
			srv.Config.ErrorLog = log.New(io.Discard, "", 0)
			srv.Start()
			defer srv.Close()

			req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/me", nil)
			req.Header.Set("Authorization", "Bearer good-token")
			if resp, err := http.DefaultClient.Do(req); err == nil {
				resp.Body.Close()
				t.Fatalf("a pánikoló kérés választ kapott (%d), a net/http-nak le kellett volna zárnia a kapcsolatot", resp.StatusCode)
			}

			// A szerver él: a következő kérés (healthz) rendben kiszolgálódik.
			resp, err := http.Get(srv.URL + "/healthz")
			if err != nil {
				t.Fatalf("a pánik után a szerver nem szolgál ki: %v", err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("healthz státusz = %d, várt 200", resp.StatusCode)
			}

			events := tr.Events()
			if len(events) != 1 {
				t.Fatalf("események = %d, várt pontosan 1 (pánikonként egy)", len(events))
			}
			ev := events[0]
			if ev.Level != sentry.LevelFatal {
				t.Errorf("level = %q, várt fatal", ev.Level)
			}
			if !strings.Contains(ev.Message+exceptionValues(ev), tt.want) {
				t.Errorf("az esemény nem tartalmazza a pánik szövegét (%q): %+v", tt.want, ev)
			}
			// "Pontos stacktrace": a pánikoló függvény látszik benne.
			if !hasFrame(ev, "panicStore") {
				t.Errorf("a stacktrace nem tartalmazza a pánikoló panicStore.UpsertProfile keretet")
			}
		})
	}
}

func TestErrtrackEventsCarryNoPersonalData(t *testing.T) {
	tr := installErrtrack(t)
	h := newTestAPIWith(&fakeStore{}, &fakeDirectory{listErr: errors.New("keycloak: timeout")})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?q=anna@example.com", strings.NewReader(`{"titok":"torzs-adat"}`))
	req.Header.Set("Authorization", "Bearer admin-token")
	req.Header.Set("Cookie", "session=suti-adat")
	req.Header.Set("X-Forwarded-For", "203.0.113.7")
	req.RemoteAddr = "198.51.100.9:4711"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("státusz = %d, várt 502", rec.Code)
	}

	events := tr.Events()
	if len(events) != 1 {
		t.Fatalf("események = %d, várt 1", len(events))
	}
	raw, err := json.Marshal(events[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"admin-token", "anna@example.com", "torzs-adat", "suti-adat", "203.0.113.7", "198.51.100.9"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("az esemény személyes adatot tartalmaz: %q\n%s", secret, raw)
		}
	}
	if u := events[0].User; !u.IsEmpty() {
		t.Errorf("az eseményen felhasználó-adat van: %+v", u)
	}
	if events[0].Request == nil || events[0].Request.Method != http.MethodGet {
		t.Errorf("a kérés metaadata (metódus) hiányzik: %+v", events[0].Request)
	}
}

func exceptionValues(ev *sentry.Event) string {
	var b strings.Builder
	for _, e := range ev.Exception {
		b.WriteString(e.Value)
	}
	return b.String()
}

// hasFrame: szerepel-e az esemény valamelyik stacktrace-ében (kivételé vagy szálé) olyan keret,
// aminek a függvényneve tartalmazza fn-t.
func hasFrame(ev *sentry.Event, fn string) bool {
	var traces []*sentry.Stacktrace
	for _, e := range ev.Exception {
		traces = append(traces, e.Stacktrace)
	}
	for _, th := range ev.Threads {
		traces = append(traces, th.Stacktrace)
	}
	for _, st := range traces {
		if st == nil {
			continue
		}
		for _, f := range st.Frames {
			if strings.Contains(f.Function, fn) {
				return true
			}
		}
	}
	return false
}
