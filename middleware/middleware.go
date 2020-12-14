package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/alexanderkarlis/sw-dnsbl/auth"
)

type key int

const (
	contextTokenKey key = iota
)

// Middleware is the callback for the mux http handler to
// process all the request and see if the a tokem is present
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")

			if tokenString == "" {
				next.ServeHTTP(w, r)
				return
			}

			if len(tokenString) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Missing Authorization Header"))
				return
			}

			tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
			_, err := auth.ValidateToken(tokenString)

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Error verifying JWT token: " + err.Error()))
				return
			}

			ctx := context.WithValue(r.Context(), contextTokenKey, tokenString)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetTokenFromContext func gets the token from the context after successful login through auth mutation
func GetTokenFromContext(ctx context.Context) string {
	tokenString, _ := ctx.Value(contextTokenKey).(string)
	return tokenString
}
