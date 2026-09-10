package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// fakeVerifier a token szövegét kulcsként használja: így a tesztek élő
// Keycloak nélkül állítanak elő érvényes és érvénytelen tokeneket.
type fakeVerifier struct {
	valid map[string]auth.Identity
}

func (f fakeVerifier) Verify(_ context.Context, raw string) (auth.Identity, error) {
	if id, ok := f.valid[raw]; ok {
		return id, nil
	}
	return auth.Identity{}, auth.ErrUnauthenticated
}

type fakeStore struct {
	pingErr      error
	upsertErr    error
	updateErr    error
	entErr       error
	profile      store.Profile
	entitlements []store.Entitlement

	lastUpsertSubject string
	lastDisplayName   *string
	lastLocale        *string
	upsertCalls       int

	// Az admin végpontok állapota külön struktúrában, lásd admin_test.go.
	admin adminFake
}

func (f *fakeStore) Ping(context.Context) error { return f.pingErr }

func (f *fakeStore) UpsertProfile(_ context.Context, subject, email, username, name string) (store.Profile, error) {
	f.upsertCalls++
	f.lastUpsertSubject = subject
	if f.upsertErr != nil {
		return store.Profile{}, f.upsertErr
	}
	p := f.profile
	p.Subject, p.Email, p.Username = subject, email, username
	if p.DisplayName == "" {
		p.DisplayName = name
	}
	return p, nil
}

func (f *fakeStore) UpdateProfile(_ context.Context, subject string, displayName, locale *string) (store.Profile, error) {
	f.lastDisplayName, f.lastLocale = displayName, locale
	if f.updateErr != nil {
		return store.Profile{}, f.updateErr
	}
	p := f.profile
	p.Subject = subject
	if displayName != nil {
		p.DisplayName = *displayName
	}
	if locale != nil {
		p.Locale = *locale
	}
	return p, nil
}

func (f *fakeStore) ActiveEntitlements(context.Context, string) ([]store.Entitlement, error) {
	if f.entErr != nil {
		return nil, f.entErr
	}
	return f.entitlements, nil
}

var testIdentity = auth.Identity{
	Subject:           "sub-123",
	Email:             "user@example.com",
	PreferredUsername: "user",
	Name:              "Teszt Elek",
}

func newTestAPI(st Store) http.Handler {
	return newTestAPIWith(st, &fakeDirectory{})
}

func TestAuthRequired(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		authHeader string
		wantStatus int
	}{
		{"me hiányzó fejléc", http.MethodGet, "/api/v1/me", "", http.StatusUnauthorized},
		{"me rossz séma", http.MethodGet, "/api/v1/me", "Basic abcdef", http.StatusUnauthorized},
		{"me üres bearer", http.MethodGet, "/api/v1/me", "Bearer ", http.StatusUnauthorized},
		{"me érvénytelen token", http.MethodGet, "/api/v1/me", "Bearer nope", http.StatusUnauthorized},
		{"me érvényes token", http.MethodGet, "/api/v1/me", "Bearer good-token", http.StatusOK},
		{"me kisbetűs bearer", http.MethodGet, "/api/v1/me", "bearer good-token", http.StatusOK},
		{"entitlements hiányzó fejléc", http.MethodGet, "/api/v1/entitlements", "", http.StatusUnauthorized},
		{"entitlements érvényes token", http.MethodGet, "/api/v1/entitlements", "Bearer good-token", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAPI(&fakeStore{})
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestGetMe(t *testing.T) {
	st := &fakeStore{}
	h := newTestAPI(st)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d, várt 200", rec.Code)
	}
	var got store.Profile
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("válasz dekódolása: %v", err)
	}
	if got.Subject != testIdentity.Subject {
		t.Errorf("sub = %q, várt %q", got.Subject, testIdentity.Subject)
	}
	if got.Email != testIdentity.Email {
		t.Errorf("email = %q, várt %q", got.Email, testIdentity.Email)
	}
	if got.DisplayName != testIdentity.Name {
		t.Errorf("display_name = %q, várt %q", got.DisplayName, testIdentity.Name)
	}
	if st.upsertCalls != 1 {
		t.Errorf("upsert hívások = %d, várt 1", st.upsertCalls)
	}
}

func TestGetMeStoreError(t *testing.T) {
	h := newTestAPI(&fakeStore{upsertErr: errors.New("boom")})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("státusz = %d, várt 500", rec.Code)
	}
}

func TestPatchMe(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		wantStatus      int
		wantDisplayName string
		wantLocale      string
	}{
		{
			name:            "név és locale",
			body:            `{"display_name":"Új Név","locale":"hu-HU"}`,
			wantStatus:      http.StatusOK,
			wantDisplayName: "Új Név",
			wantLocale:      "hu-HU",
		},
		{
			name:            "csak név",
			body:            `{"display_name":"  Csak Név  "}`,
			wantStatus:      http.StatusOK,
			wantDisplayName: "Csak Név",
		},
		{
			name:       "csak locale",
			body:       `{"locale":"en"}`,
			wantStatus: http.StatusOK,
			wantLocale: "en",
		},
		{"üres név elutasítva", `{"display_name":"   "}`, http.StatusBadRequest, "", ""},
		{"túl hosszú név", `{"display_name":"` + strings.Repeat("a", 101) + `"}`, http.StatusBadRequest, "", ""},
		{"rossz locale", `{"locale":"magyarul"}`, http.StatusBadRequest, "", ""},
		{"email módosítás tiltva", `{"email":"hacker@example.com"}`, http.StatusBadRequest, "", ""},
		{"sub módosítás tiltva", `{"sub":"other-sub"}`, http.StatusBadRequest, "", ""},
		{"üres objektum", `{}`, http.StatusBadRequest, "", ""},
		{"hibás json", `{`, http.StatusBadRequest, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &fakeStore{}
			h := newTestAPI(st)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/me", strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer good-token")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var got store.Profile
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("válasz dekódolása: %v", err)
			}
			if tt.wantDisplayName != "" && got.DisplayName != tt.wantDisplayName {
				t.Errorf("display_name = %q, várt %q", got.DisplayName, tt.wantDisplayName)
			}
			if tt.wantLocale != "" && got.Locale != tt.wantLocale {
				t.Errorf("locale = %q, várt %q", got.Locale, tt.wantLocale)
			}
			// A nem küldött mezőknek nil-ként kell a store-ig eljutniuk,
			// hogy a COALESCE megőrizze a korábbi értéket.
			if tt.wantDisplayName == "" && st.lastDisplayName != nil {
				t.Errorf("display_name nem nil volt: %q", *st.lastDisplayName)
			}
			if tt.wantLocale == "" && st.lastLocale != nil {
				t.Errorf("locale nem nil volt: %q", *st.lastLocale)
			}
			if got.Subject != testIdentity.Subject {
				t.Errorf("sub = %q, várt %q", got.Subject, testIdentity.Subject)
			}
		})
	}
}

func TestEntitlements(t *testing.T) {
	expires := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		store      *fakeStore
		wantStatus int
		wantKeys   []string
	}{
		{
			name:       "üres lista normál válasz",
			store:      &fakeStore{entitlements: nil},
			wantStatus: http.StatusOK,
			wantKeys:   []string{},
		},
		{
			name: "aktív jogosultságok",
			store: &fakeStore{entitlements: []store.Entitlement{
				{FeatureKey: "cloud-sync", GrantedAt: time.Now()},
				{FeatureKey: "premium-export", GrantedAt: time.Now(), ExpiresAt: &expires},
			}},
			wantStatus: http.StatusOK,
			wantKeys:   []string{"cloud-sync", "premium-export"},
		},
		{
			name:       "store hiba",
			store:      &fakeStore{entErr: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAPI(tt.store)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/entitlements", nil)
			req.Header.Set("Authorization", "Bearer good-token")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d", rec.Code, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var got entitlementsResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("válasz dekódolása: %v", err)
			}
			if got.Entitlements == nil {
				t.Fatal("entitlements null volt, üres tömb helyett")
			}
			if len(got.Entitlements) != len(tt.wantKeys) {
				t.Fatalf("elemszám = %d, várt %d", len(got.Entitlements), len(tt.wantKeys))
			}
			for i, k := range tt.wantKeys {
				if got.Entitlements[i].FeatureKey != k {
					t.Errorf("[%d] feature_key = %q, várt %q", i, got.Entitlements[i].FeatureKey, k)
				}
			}
		})
	}
}

func TestHealthz(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
	}{
		{"adatbázis elérhető", nil, http.StatusOK},
		{"adatbázis nem elérhető", errors.New("connection refused"), http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAPI(&fakeStore{pingErr: tt.pingErr})
			// Szándékosan Authorization fejléc nélkül: a healthz nyílt végpont.
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		origin      string
		preflight   bool
		wantStatus  int
		wantAllowed string
	}{
		{"engedélyezett origin preflight", http.MethodOptions, "http://localhost:5174", true, http.StatusNoContent, "http://localhost:5174"},
		{"idegen origin preflight", http.MethodOptions, "http://evil.example", true, http.StatusNoContent, ""},
		{"engedélyezett origin sima kérés", http.MethodGet, "http://localhost:5174", false, http.StatusOK, "http://localhost:5174"},
		{"origin nélkül", http.MethodGet, "", false, http.StatusOK, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAPI(&fakeStore{})
			req := httptest.NewRequest(tt.method, "/healthz", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.preflight {
				req.Header.Set("Access-Control-Request-Method", "GET")
				req.Header.Set("Access-Control-Request-Headers", "authorization")
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantAllowed {
				t.Errorf("Allow-Origin = %q, várt %q", got, tt.wantAllowed)
			}
			if tt.wantAllowed != "" {
				if h := rec.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(h, "Authorization") {
					t.Errorf("Allow-Headers nem tartalmazza az Authorization-t: %q", h)
				}
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newTestAPI(&fakeStore{})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("státusz = %d, várt 405", rec.Code)
	}
}

func TestValidLocale(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"hu", true},
		{"hu-HU", true},
		{"en", true},
		{"en-GB", true},
		{"h", false},
		{"hu-HU-x", false},
		{"hu_HU", false},
		{"12", false},
		{"magyarnyelv", false},
	}
	for _, tt := range tests {
		if got := validLocale(tt.in); got != tt.want {
			t.Errorf("validLocale(%q) = %v, várt %v", tt.in, got, tt.want)
		}
	}
}
