package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/niktheblak/web-common/pkg/auth"
)

const testToken = "test_token_2dc9a"

func TestAuthenticator(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "OK"); err != nil {
			t.Error(err)
		}
	})
	tests := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{"Authenticated", "Bearer " + testToken, http.StatusOK},
		{"Lowercase scheme", "bearer " + testToken, http.StatusOK},
		{"Uppercase scheme", "BEARER " + testToken, http.StatusOK},
		{"Extra whitespace around token", "Bearer   " + testToken + "  ", http.StatusOK},
		{"Unauthenticated", "", http.StatusUnauthorized},
		{"Invalid token", "Bearer other_token_7a3b1", http.StatusUnauthorized},
		{"Token without scheme", testToken, http.StatusUnauthorized},
		{"Wrong scheme", "Basic " + testToken, http.StatusUnauthorized},
		{"Scheme without token", "Bearer", http.StatusUnauthorized},
		{"Empty token", "Bearer ", http.StatusUnauthorized},
		{"Scheme as a prefix of another word", "BearerToken " + testToken, http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := Authenticator(handler, auth.Static(testToken))
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			a.ServeHTTP(w, req)
			resp := w.Result()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if tt.wantStatus == http.StatusUnauthorized {
				assert.Equal(t, "Bearer", resp.Header.Get("WWW-Authenticate"))
			}
		})
	}
}

// A permissive Authenticator must still see requests that carry no credentials at all.
func TestAuthenticatorAlwaysAllow(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	a := Authenticator(handler, auth.AlwaysAllow())
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}
