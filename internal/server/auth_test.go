package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"spacedchess/internal/config"
	"spacedchess/internal/store"
)

func deleteUser(t *testing.T, username string) {
	t.Helper()

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Errorf("cleanup connect: %v", err)
		return
	}
	defer conn.Close(context.Background())

	if _, err := conn.Exec(context.Background(), `DELETE FROM users WHERE username = $1`, username); err != nil {
		t.Errorf("cleanup user %s: %v", username, err)
	}
}

// Exercises the handlers against the seeded dev database, so it needs the
// compose Postgres and the seeded 'nairwolf' user.
func testServer(t *testing.T) http.Handler {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	s, err := store.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(s.Close)
	return NewServer(s, config.Config{})
}

func do(t *testing.T, h http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func sessionCookieOf(t *testing.T, w *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, c := range w.Result().Cookies() {
		if c.Name == sessionCookie {
			return c
		}
	}
	t.Fatal("response set no session cookie")
	return nil
}

func login(t *testing.T, h http.Handler) *http.Cookie {
	t.Helper()

	w := do(t, h, "POST", "/auth/login", `{"username":"nairwolf","password":"password"}`, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("login = %d, want 200: %s", w.Code, w.Body)
	}
	return sessionCookieOf(t, w)
}

func TestLoginAndMe(t *testing.T) {
	h := testServer(t)
	c := login(t, h)

	w := do(t, h, "GET", "/auth/me", "", c)
	if w.Code != http.StatusOK {
		t.Fatalf("me = %d, want 200: %s", w.Code, w.Body)
	}
	if !strings.Contains(w.Body.String(), `"username":"nairwolf"`) {
		t.Errorf("me body = %s", w.Body)
	}

	// Clean up the session this test created.
	do(t, h, "POST", "/auth/logout", "", c)
}

func TestLoginRejectsBadPassword(t *testing.T) {
	h := testServer(t)

	w := do(t, h, "POST", "/auth/login", `{"username":"nairwolf","password":"wrong"}`, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("login = %d, want 401: %s", w.Code, w.Body)
	}
}

func TestLoginRejectsUnknownUser(t *testing.T) {
	h := testServer(t)

	w := do(t, h, "POST", "/auth/login", `{"username":"nobody_here","password":"password"}`, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("login = %d, want 401: %s", w.Code, w.Body)
	}
}

func TestMeRequiresSession(t *testing.T) {
	h := testServer(t)

	tests := []struct {
		name   string
		cookie *http.Cookie
	}{
		{name: "no cookie"},
		{name: "expired session", cookie: &http.Cookie{Name: sessionCookie, Value: "dev_session_expired"}},
		{name: "unknown token", cookie: &http.Cookie{Name: sessionCookie, Value: "nope"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := do(t, h, "GET", "/auth/me", "", tt.cookie)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("me = %d, want 401: %s", w.Code, w.Body)
			}
		})
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	h := testServer(t)
	c := login(t, h)

	w := do(t, h, "POST", "/auth/logout", "", c)
	if w.Code != http.StatusNoContent {
		t.Fatalf("logout = %d, want 204: %s", w.Code, w.Body)
	}

	if w := do(t, h, "GET", "/auth/me", "", c); w.Code != http.StatusUnauthorized {
		t.Errorf("me after logout = %d, want 401", w.Code)
	}
}

func TestLogoutWithoutCookie(t *testing.T) {
	h := testServer(t)

	w := do(t, h, "POST", "/auth/logout", "", nil)
	if w.Code != http.StatusNoContent {
		t.Errorf("logout = %d, want 204: %s", w.Code, w.Body)
	}
}

func TestRegisterThenLogin(t *testing.T) {
	h := testServer(t)
	name := fmt.Sprintf("test_%d", time.Now().UnixNano())
	body := fmt.Sprintf(`{"username":%q,"password":"password"}`, name)

	w := do(t, h, "POST", "/auth/register", body, nil)
	if w.Code != http.StatusCreated {
		t.Fatalf("register = %d, want 201: %s", w.Code, w.Body)
	}
	t.Cleanup(func() { deleteUser(t, name) })

	// Registering must log you in — a 201 with no usable cookie leaves the
	// client thinking it is signed in until the next page load.
	c := sessionCookieOf(t, w)
	if w := do(t, h, "GET", "/auth/me", "", c); w.Code != http.StatusOK {
		t.Errorf("me after register = %d, want 200: %s", w.Code, w.Body)
	}

	if w := do(t, h, "POST", "/auth/login", body, nil); w.Code != http.StatusOK {
		t.Errorf("login = %d, want 200: %s", w.Code, w.Body)
	}

	// A second account with the same name is the one thing open registration
	// must still refuse.
	if w := do(t, h, "POST", "/auth/register", body, nil); w.Code != http.StatusConflict {
		t.Errorf("duplicate register = %d, want 409: %s", w.Code, w.Body)
	}
}

func TestRegisterValidatesInput(t *testing.T) {
	h := testServer(t)

	tests := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: `{`},
		{name: "empty username", body: `{"username":"","password":"password"}`},
		{name: "short password", body: `{"username":"someone","password":"short"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := do(t, h, "POST", "/auth/register", tt.body, nil)
			if w.Code != http.StatusBadRequest {
				t.Errorf("register = %d, want 400: %s", w.Code, w.Body)
			}
		})
	}
}
