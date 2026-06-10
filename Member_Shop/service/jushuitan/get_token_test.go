package jushuitan

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTokenUsesRefreshURLForTestEnvironment(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("app_key"); got != "test-app-key" {
			t.Fatalf("expected test app key, got %q", got)
		}
		if got := r.Form.Get("code"); got != "test-auth-code" {
			t.Fatalf("expected test auth code, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"success","data":{"access_token":"fresh-test-token","expires_in":3600}}`))
	}))
	defer server.Close()

	t.Setenv("APP_DOTENV_PATH", t.TempDir()+"/missing.env")
	t.Setenv("JST_ENV_STAGE", "test")
	t.Setenv("JST_APP_KEY_TEST", "test-app-key")
	t.Setenv("JST_APP_SECRET_TEST", "test-app-secret")
	t.Setenv("JST_AUTH_CODE_TEST", "test-auth-code")
	t.Setenv("JST_GET_TOKEN_URL_TEST", server.URL)
	t.Setenv("JST_ACCESS_TOKEN_TEST", "stale-hardcoded-token")

	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken returned error: %v", err)
	}
	if token != "fresh-test-token" {
		t.Fatalf("expected refreshed token, got %q", token)
	}
	if requestCount != 1 {
		t.Fatalf("expected one refresh request, got %d", requestCount)
	}
}
