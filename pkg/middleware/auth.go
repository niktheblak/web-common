package middleware

import (
	"net/http"
	"strings"

	"github.com/niktheblak/web-common/pkg/auth"
)

// Authenticator returns a HTTP handler that checks the request's Authorization header and proceeds or rejects the request.
//
// The token is read from a Bearer credential as defined by RFC 7235, whose scheme name is matched
// case-insensitively. A request carrying no Bearer credential at all is authenticated with an empty
// token rather than being rejected outright, so that permissive implementations such as
// auth.AlwaysAllow still let it through.
func Authenticator(handler http.Handler, authenticator auth.Authenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := authenticator.Authenticate(r.Context(), bearerToken(r))
		if err != nil {
			unauthorized(w)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// bearerToken returns the token from the request's Bearer Authorization header, or an empty string
// if the header is missing or holds credentials of some other scheme.
func bearerToken(r *http.Request) string {
	const scheme = "Bearer"
	h := r.Header.Get("Authorization")
	if len(h) <= len(scheme) || !strings.EqualFold(h[:len(scheme)], scheme) || h[len(scheme)] != ' ' {
		return ""
	}
	return strings.TrimSpace(h[len(scheme)+1:])
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
