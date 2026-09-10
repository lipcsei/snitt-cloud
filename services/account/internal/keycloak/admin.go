// Package keycloak a Keycloak Admin REST API vékony kliense.
//
// Az identitást (e-mail, név, engedélyezett-e) a Keycloak birtokolja, ezért az
// admin felület onnan olvassa, nem másoljuk át Postgresbe. A hozzáférés egy
// bizalmas, service accountos klienssel történik (client_credentials).
package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// User a Keycloak felhasználó-reprezentáció minket érdeklő része.
type User struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Enabled          bool   `json:"enabled"`
	EmailVerified    bool   `json:"emailVerified"`
	CreatedTimestamp int64  `json:"createdTimestamp"`
}

// CreatedAt a Keycloak ezredmásodperces időbélyegét adja vissza időpontként.
func (u User) CreatedAt() time.Time {
	if u.CreatedTimestamp == 0 {
		return time.Time{}
	}
	return time.UnixMilli(u.CreatedTimestamp).UTC()
}

// FullName a megjeleníthető név a Keycloakban tárolt mezőkből.
func (u User) FullName() string {
	name := strings.TrimSpace(u.FirstName + " " + u.LastName)
	if name == "" {
		return u.Username
	}
	return name
}

// ErrNotFound akkor jön vissza, ha a Keycloak nem ismeri a kért felhasználót.
var ErrNotFound = fmt.Errorf("a felhasználó nem található a Keycloakban")

// AdminClient a service account tokent gyorsítótárazva hívja az Admin API-t.
type AdminClient struct {
	baseURL      string
	realm        string
	clientID     string
	clientSecret string
	hc           *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
}

// NewAdminClient beállítja a klienst. Hálózati hívást nem végez: az első
// tokenkérés az első tényleges lekérdezéskor történik, így a szolgáltatás
// akkor is elindul, ha a Keycloak még nem áll készen.
func NewAdminClient(baseURL, realm, clientID, clientSecret string) *AdminClient {
	return &AdminClient{
		baseURL:      strings.TrimRight(baseURL, "/"),
		realm:        realm,
		clientID:     clientID,
		clientSecret: clientSecret,
		hc:           &http.Client{Timeout: 10 * time.Second},
	}
}

// ListUsers lapozott felhasználólistát ad; a search a Keycloak beépített,
// felhasználónév/e-mail/vezetéknév/keresztnév mezőkön futó keresése.
func (c *AdminClient) ListUsers(ctx context.Context, search string, first, max int) ([]User, error) {
	q := url.Values{}
	q.Set("first", strconv.Itoa(first))
	q.Set("max", strconv.Itoa(max))
	// briefRepresentation nélkül a Keycloak minden felhasználóhoz lehúzza a
	// credential- és attribútum-adatokat is, amikre itt nincs szükség.
	q.Set("briefRepresentation", "true")
	if search != "" {
		q.Set("search", search)
	}

	var users []User
	if err := c.getJSON(ctx, "/users?"+q.Encode(), &users); err != nil {
		return nil, err
	}
	return users, nil
}

// CountUsers a keresésre illeszkedő felhasználók száma (lapozáshoz).
func (c *AdminClient) CountUsers(ctx context.Context, search string) (int, error) {
	q := url.Values{}
	if search != "" {
		q.Set("search", search)
	}
	path := "/users/count"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var n int
	if err := c.getJSON(ctx, path, &n); err != nil {
		return 0, err
	}
	return n, nil
}

// GetUser egyetlen felhasználó a Keycloak azonosítója (a token sub-ja) alapján.
func (c *AdminClient) GetUser(ctx context.Context, id string) (User, error) {
	var u User
	if err := c.getJSON(ctx, "/users/"+url.PathEscape(id), &u); err != nil {
		return User{}, err
	}
	return u, nil
}

// SetUserEnabled a Keycloak fiókot engedélyezi vagy letiltja. Letiltott
// fiókkal se belépni, se tokent frissíteni nem lehet – ez az egyetlen olyan
// admin művelet, ami a felhasználó azonosságát birtokló rendszert írja.
//
// Ehhez a service accountnak manage-users szerep kell a realm-management
// kliensen; e nélkül a Keycloak 403-at ad, és a hibaüzenet ezt meg is mondja.
func (c *AdminClient) SetUserEnabled(ctx context.Context, id string, enabled bool) error {
	// Szándékosan csak az enabled mezőt küldjük: a Keycloak a hiányzó
	// mezőket változatlanul hagyja, így egy párhuzamos szerkesztés nem
	// íródik felül.
	body := map[string]any{"enabled": enabled}
	return c.doJSON(ctx, http.MethodPut, "/users/"+url.PathEscape(id), body, nil)
}

func (c *AdminClient) getJSON(ctx context.Context, path string, out any) error {
	return c.doJSON(ctx, http.MethodGet, path, nil, out)
}

// doJSON a közös Keycloak admin API hívás: token, hibafordítás, JSON.
// Az in nil értéke törzs nélküli kérés, az out nil értéke eldobott válasz.
func (c *AdminClient) doJSON(ctx context.Context, method, path string, in, out any) error {
	tok, err := c.accessToken(ctx)
	if err != nil {
		return err
	}

	var payload io.Reader
	if in != nil {
		buf, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("keycloak kérés törzse: %w", err)
		}
		payload = bytes.NewReader(buf)
	}

	endpoint := c.baseURL + "/admin/realms/" + url.PathEscape(c.realm) + path
	req, err := http.NewRequestWithContext(ctx, method, endpoint, payload)
	if err != nil {
		return fmt.Errorf("keycloak kérés összeállítása: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("keycloak admin API hívás: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		// A gyorsítótárazott token elévülhetett vagy a service accountnak
		// nincs realm-management joga; a következő hívás új tokent kér.
		c.invalidate()
		return fmt.Errorf("keycloak admin API elutasította a kérést (%d) – ellenőrizd a service account realm-management szerepeit", resp.StatusCode)
	case resp.StatusCode >= 300:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("keycloak admin API hiba (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if out == nil {
		// A módosító hívások 204-et adnak, üres törzzsel.
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("keycloak válasz dekódolása: %w", err)
	}
	return nil
}

func (c *AdminClient) invalidate() {
	c.mu.Lock()
	c.token, c.expires = "", time.Time{}
	c.mu.Unlock()
}

func (c *AdminClient) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.expires) {
		return c.token, nil
	}

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
	}
	endpoint := c.baseURL + "/realms/" + url.PathEscape(c.realm) + "/protocol/openid-connect/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("token kérés összeállítása: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("service account token kérése: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("service account token elutasítva (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tr struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("token válasz dekódolása: %w", err)
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("a Keycloak üres access tokent adott")
	}

	// 30 másodperc ráhagyás, hogy a lejárat pillanatában ne fusson bele
	// egy már érvénytelen tokenbe.
	lifetime := time.Duration(tr.ExpiresIn) * time.Second
	if lifetime > 30*time.Second {
		lifetime -= 30 * time.Second
	}
	c.token, c.expires = tr.AccessToken, time.Now().Add(lifetime)
	return c.token, nil
}
