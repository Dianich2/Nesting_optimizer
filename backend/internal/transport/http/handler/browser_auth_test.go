package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httperror "server_nesting_optimizer/internal/transport/http/errors"
	userusecase "server_nesting_optimizer/internal/usecase/user"

	"github.com/gofiber/fiber/v3"
)

const testBrowserOrigin = "http://localhost:5173"

func browserTestApp() *fiber.App {
	return fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
}

func TestBrowserAuthRejectsForeignOrigin(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		register func(*fiber.App, *BrowserAuthHandler)
	}{
		{"login", "/login", func(app *fiber.App, h *BrowserAuthHandler) { app.Post("/login", h.Login) }},
		{"refresh", "/refresh", func(app *fiber.App, h *BrowserAuthHandler) { app.Post("/refresh", h.Refresh) }},
		{"logout", "/logout", func(app *fiber.App, h *BrowserAuthHandler) { app.Post("/logout", h.Logout) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := browserTestApp()
			h := &BrowserAuthHandler{allowedOrigin: testBrowserOrigin}
			tt.register(app, h)

			for _, origin := range []string{"http://untrusted.example", ""} {
				req := httptest.NewRequest(http.MethodPost, tt.path, nil)
				if origin != "" {
					req.Header.Set("Origin", origin)
				}

				resp, err := app.Test(req)
				if err != nil {
					t.Fatal(err)
				}

				var body struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				decodeErr := json.NewDecoder(resp.Body).Decode(&body)
				_ = resp.Body.Close()

				if decodeErr != nil {
					t.Fatal(decodeErr)
				}
				if resp.StatusCode != http.StatusForbidden || body.Error.Code != "forbidden_error" {
					t.Fatalf("origin %q: status=%d code=%q, want 403/forbidden_error", origin, resp.StatusCode, body.Error.Code)
				}
			}
		})
	}
}

func TestBrowserRefreshWithoutCookie(t *testing.T) {
	app := browserTestApp()
	h := &BrowserAuthHandler{
		allowedOrigin:  testBrowserOrigin,
		refreshUseCase: userusecase.NewRefreshUseCase(nil, nil, nil),
	}
	app.Post("/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.Header.Set("Origin", testBrowserOrigin)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400", resp.StatusCode)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "validation_error" {
		t.Fatalf("code=%q, want validation_error", body.Error.Code)
	}
}

func TestBrowserRefreshCookieAttributes(t *testing.T) {
	app := browserTestApp()
	h := &BrowserAuthHandler{cookieSecure: false, cookieMaxAge: 3600}

	app.Get("/set", func(c fiber.Ctx) error {
		h.setRefreshCookie(c, "test-token")
		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Get("/clear", func(c fiber.Ctx) error {
		h.clearRefreshCookie(c)
		return c.SendStatus(fiber.StatusNoContent)
	})

	for _, tt := range []struct {
		path       string
		wantValue  string
		wantMaxAge int
	}{
		{"/set", "test-token", 3600},
		{"/clear", "", -1},
	} {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, tt.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			_ = resp.Body.Close()

			cookies := resp.Cookies()
			if len(cookies) != 1 {
				t.Fatalf("got %d cookies, want 1", len(cookies))
			}
			cookie := cookies[0]
			if cookie.Name != browserCookieName || cookie.Value != tt.wantValue ||
				cookie.Path != browserCookiePath || cookie.MaxAge != tt.wantMaxAge ||
				!cookie.HttpOnly || cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
				t.Fatalf("unexpected cookie attributes: name=%q path=%q maxAge=%d httpOnly=%v secure=%v sameSite=%v",
					cookie.Name, cookie.Path, cookie.MaxAge, cookie.HttpOnly, cookie.Secure, cookie.SameSite)
			}
		})
	}
}
