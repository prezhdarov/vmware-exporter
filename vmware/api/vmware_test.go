package vmware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestGETWithHeaders(t *testing.T) {
	restoreVMwareFlags(t)

	*vmwTLS = false
	*vmwInterval = 20

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want %q", r.Method, http.MethodGet)
		}

		if got := r.Header.Get("X-Test-Header"); got != "test-value" {
			t.Fatalf("X-Test-Header = %q, want %q", got, "test-value")
		}

		w.Header().Set("Cookie", "vmware-session=fake-session")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	statusCode, cookie, body, err := request(
		http.MethodGet,
		server.URL,
		map[string]string{"X-Test-Header": "test-value"},
		false,
	)
	if err != nil {
		t.Fatalf("request() returned error: %v", err)
	}

	if statusCode != http.StatusOK {
		t.Fatalf("statusCode = %d, want %d", statusCode, http.StatusOK)
	}

	if cookie != "vmware-session=fake-session" {
		t.Fatalf("cookie = %q, want %q", cookie, "vmware-session=fake-session")
	}

	if string(body) != "ok" {
		t.Fatalf("body = %q, want %q", string(body), "ok")
	}
}

func TestRequestPOSTWithBasicAuthWhenLoginIsTrue(t *testing.T) {
	restoreVMwareFlags(t)

	*vmwUser = "test-user"
	*vmwPasswd = "test-password"
	*vmwTLS = false
	*vmwInterval = 20

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want %q", r.Method, http.MethodPost)
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			t.Fatal("expected BasicAuth to be set")
		}

		if username != "test-user" {
			t.Fatalf("username = %q, want %q", username, "test-user")
		}

		if password != "test-password" {
			t.Fatalf("password = %q, want %q", password, "test-password")
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))
	defer server.Close()

	statusCode, _, body, err := request(
		http.MethodPost,
		server.URL,
		map[string]string{},
		true,
	)
	if err != nil {
		t.Fatalf("request() returned error: %v", err)
	}

	if statusCode != http.StatusCreated {
		t.Fatalf("statusCode = %d, want %d", statusCode, http.StatusCreated)
	}

	if string(body) != "created" {
		t.Fatalf("body = %q, want %q", string(body), "created")
	}
}

func TestRequestReturnsErrorForInvalidURL(t *testing.T) {
	restoreVMwareFlags(t)

	*vmwTLS = false
	*vmwInterval = 20

	statusCode, cookie, body, err := request(
		http.MethodGet,
		":// invalid-url",
		map[string]string{},
		false,
	)

	if err == nil {
		t.Fatal("expected request() to return an error")
	}

	if statusCode != 0 {
		t.Fatalf("statusCode = %d, want 0", statusCode)
	}

	if cookie != "" {
		t.Fatalf("cookie = %q, want empty string", cookie)
	}

	if body != nil {
		t.Fatalf("body = %q, want nil", string(body))
	}
}

func TestVMwareGetReturnsResponseBody(t *testing.T) {
	restoreVMwareFlags(t)

	*vmwSchema = "http"
	*vmwTLS = false
	*vmwInterval = 20

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/test" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/api/test")
		}

		if got := r.Header.Get("X-Session"); got != "fake-session" {
			t.Fatalf("X-Session = %q, want %q", got, "fake-session")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	target := strings.TrimPrefix(server.URL, "http://")

	loginData := map[string]interface{}{
		"target": target,
		"headers": map[string]string{
			"X-Session": "fake-session",
		},
	}

	extraConfig := map[string]interface{}{
		"api": "/api/test",
	}

	vm := NewAPI()

	got, err := vm.Get(loginData, extraConfig, logger)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	body, ok := got.(*[]byte)
	if !ok {
		t.Fatalf("Get() returned %T, want *[]byte", got)
	}

	if string(*body) != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", string(*body), `{"status":"ok"}`)
	}
}

func TestLogoutDoesNothingAndReturnsNil(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	vm := NewAPI()

	err := vm.Logout(map[string]interface{}{}, logger)
	if err != nil {
		t.Fatalf("Logout() returned error: %v", err)
	}
}

func restoreVMwareFlags(t *testing.T) {
	t.Helper()

	oldUser := *vmwUser
	oldPassword := *vmwPasswd
	oldVCenter := *vCenter
	oldSchema := *vmwSchema
	oldTLS := *vmwTLS
	oldInterval := *vmwInterval
	oldGranularity := *vmGranularity

	t.Cleanup(func() {
		*vmwUser = oldUser
		*vmwPasswd = oldPassword
		*vCenter = oldVCenter
		*vmwSchema = oldSchema
		*vmwTLS = oldTLS
		*vmwInterval = oldInterval
		*vmGranularity = oldGranularity
	})
}