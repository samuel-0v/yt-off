package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestNetworkIPFromRequestPrefersForwardedHost(t *testing.T) {
	got := networkIPFromTestRequest(t, "192.168.1.77:5173", "192.168.1.10")
	if got != "192.168.1.77" {
		t.Fatalf("networkIPFromRequest() = %q, want 192.168.1.77", got)
	}
}

func TestNetworkIPFromRequestFallsBackToConfiguredIP(t *testing.T) {
	got := networkIPFromTestRequest(t, "localhost:5173", "192.168.1.10")
	if got != "192.168.1.10" {
		t.Fatalf("networkIPFromRequest() = %q, want 192.168.1.10", got)
	}
}

func networkIPFromTestRequest(t *testing.T, forwardedHost string, configuredIP string) string {
	t.Helper()

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString(networkIPFromRequest(c, configuredIP))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Forwarded-Host", forwardedHost)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	return string(body)
}
