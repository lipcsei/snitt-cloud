package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/keycloak"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

func TestGrantEntitlement(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantStatus  int
		wantExpires bool
		wantWarning bool
	}{
		{
			name: "csomagon kívüli kulcs, lejárat nélkül",
			body: `{"feature_key":"beta-access"}`, wantStatus: http.StatusCreated,
		},
		{
			name: "futamidővel", body: `{"feature_key":"beta-access","days":30}`,
			wantStatus: http.StatusCreated, wantExpires: true,
		},
		{
			// A csomag-kulcsot ki lehet adni, de az admin figyelmeztetést kap:
			// a következő előfizetés-művelet felülírja.
			name:       "csomag által vezérelt kulcs figyelmeztet",
			body:       `{"feature_key":"ai-semantic-search"}`,
			wantStatus: http.StatusCreated, wantWarning: true,
		},
		{"üres kulcs", `{"feature_key":""}`, http.StatusBadRequest, false, false},
		{"nagybetűs kulcs", `{"feature_key":"Beta"}`, http.StatusBadRequest, false, false},
		{"szóköz a kulcsban", `{"feature_key":"beta access"}`, http.StatusBadRequest, false, false},
		{"nulla nap", `{"feature_key":"beta","days":0}`, http.StatusBadRequest, false, false},
		{"túl hosszú futamidő", `{"feature_key":"beta","days":99999}`, http.StatusBadRequest, false, false},
		{"ismeretlen mező", `{"feature_key":"beta","orokre":true}`, http.StatusBadRequest, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &fakeStore{}
			h := newTestAPIWith(st, &fakeDirectory{})

			rec := do(t, h, http.MethodPost, "/api/v1/admin/users/u1/entitlements", "admin-token", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusCreated {
				if len(st.admin.audit) != 0 {
					t.Errorf("elutasított kérés naplózott: %+v", st.admin.audit)
				}
				return
			}

			var resp struct {
				Entitlement store.Entitlement `json:"entitlement"`
				Warning     string            `json:"warning"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("válasz: %v", err)
			}
			if (resp.Entitlement.ExpiresAt != nil) != tt.wantExpires {
				t.Errorf("lejárat = %v, várt lejárat: %v", resp.Entitlement.ExpiresAt, tt.wantExpires)
			}
			if (resp.Warning != "") != tt.wantWarning {
				t.Errorf("figyelmeztetés = %q, várt: %v", resp.Warning, tt.wantWarning)
			}
			if len(st.admin.audit) != 1 || st.admin.audit[0].Action != store.AuditEntitlementGrant {
				t.Errorf("napló = %+v", st.admin.audit)
			}
		})
	}
}

func TestRevokeEntitlement(t *testing.T) {
	st := &fakeStore{admin: adminFake{entitlements: map[string][]store.Entitlement{
		"u1": {{FeatureKey: "beta-access", GrantedAt: time.Now().UTC()}},
	}}}
	h := newTestAPIWith(st, &fakeDirectory{})

	rec := do(t, h, http.MethodDelete, "/api/v1/admin/users/u1/entitlements/beta-access", "admin-token", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d (törzs: %s)", rec.Code, rec.Body.String())
	}
	if len(st.admin.entitlements["u1"]) != 0 {
		t.Errorf("a jogosultság megmaradt: %+v", st.admin.entitlements["u1"])
	}
	if len(st.admin.audit) != 1 || st.admin.audit[0].Action != store.AuditEntitlementRevoke {
		t.Fatalf("napló = %+v", st.admin.audit)
	}

	// Ami nincs kiadva, azt nem lehet visszavonni: 404, és nincs napló róla.
	rec = do(t, h, http.MethodDelete, "/api/v1/admin/users/u1/entitlements/beta-access", "admin-token", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("második törlés = %d, várt 404 (törzs: %s)", rec.Code, rec.Body.String())
	}
	if len(st.admin.audit) != 1 {
		t.Errorf("a sikertelen törlés naplózott: %+v", st.admin.audit)
	}
}

func TestSetUserEnabled(t *testing.T) {
	users := []keycloak.User{{ID: "u1", Username: "anna", Email: "anna@example.com", Enabled: true}}

	t.Run("letiltás", func(t *testing.T) {
		dir := &fakeDirectory{users: append([]keycloak.User(nil), users...)}
		st := &fakeStore{}
		h := newTestAPIWith(st, dir)

		rec := do(t, h, http.MethodPatch, "/api/v1/admin/users/u1", "admin-token", `{"enabled":false}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("státusz = %d (törzs: %s)", rec.Code, rec.Body.String())
		}
		if dir.users[0].Enabled {
			t.Error("a fiók a Keycloakban engedélyezve maradt")
		}
		if len(st.admin.audit) != 1 || st.admin.audit[0].Action != store.AuditUserDisable {
			t.Fatalf("napló = %+v", st.admin.audit)
		}
	})

	t.Run("már ebben az állapotban van", func(t *testing.T) {
		dir := &fakeDirectory{users: append([]keycloak.User(nil), users...)}
		st := &fakeStore{}
		h := newTestAPIWith(st, dir)

		rec := do(t, h, http.MethodPatch, "/api/v1/admin/users/u1", "admin-token", `{"enabled":true}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("státusz = %d (törzs: %s)", rec.Code, rec.Body.String())
		}
		if dir.enabledCalls != 0 {
			t.Error("felesleges Keycloak írás történt")
		}
		// Nincs változás, tehát félrevezető naplósor sincs.
		if len(st.admin.audit) != 0 {
			t.Errorf("napló = %+v", st.admin.audit)
		}
	})

	t.Run("nincs ilyen felhasználó", func(t *testing.T) {
		h := newTestAPIWith(&fakeStore{}, &fakeDirectory{})
		rec := do(t, h, http.MethodPatch, "/api/v1/admin/users/nincs", "admin-token", `{"enabled":false}`)
		if rec.Code != http.StatusNotFound {
			t.Errorf("státusz = %d, várt 404 (törzs: %s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("hiányzó enabled mező", func(t *testing.T) {
		dir := &fakeDirectory{users: append([]keycloak.User(nil), users...)}
		h := newTestAPIWith(&fakeStore{}, dir)
		rec := do(t, h, http.MethodPatch, "/api/v1/admin/users/u1", "admin-token", `{}`)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("státusz = %d, várt 400 (törzs: %s)", rec.Code, rec.Body.String())
		}
	})

	// A Keycloak írása bukhat (pl. hiányzó manage-users szerep): ilyenkor a
	// hiba látszik, és nem keletkezik napló egy meg nem történt változásról.
	t.Run("keycloak elutasítja", func(t *testing.T) {
		dir := &fakeDirectory{
			users:         append([]keycloak.User(nil), users...),
			setEnabledErr: errors.New("keycloak admin API elutasította a kérést (403)"),
		}
		st := &fakeStore{}
		h := newTestAPIWith(st, dir)

		rec := do(t, h, http.MethodPatch, "/api/v1/admin/users/u1", "admin-token", `{"enabled":false}`)
		if rec.Code != http.StatusBadGateway {
			t.Fatalf("státusz = %d, várt 502 (törzs: %s)", rec.Code, rec.Body.String())
		}
		if len(st.admin.audit) != 0 {
			t.Errorf("napló = %+v", st.admin.audit)
		}
	})
}
