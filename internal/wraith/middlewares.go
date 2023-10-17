// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"context"
	"github.com/mdhender/wraithi/internal/sessions"
	"log"
	"net/http"
)

func (a *App) mustHaveBoosted(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Hx-Request") != "true" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// HTTP middleware setting a value on the request context
func (a *App) mustHaveRole(roles ...string) func(http.Handler) http.Handler {
	nfh := a.notFound()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := a.currentUser(r)
			if !user.HasRole(roles...) {
				log.Printf("%s %s: missing roles %v\n", r.Method, r.URL, roles)
				nfh(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// withUserMiddleware injects a User into the request's Context.
// It searches for a session token in the Bearer Token header first,
// and then in a session cookie. If a valid token is found, the User
// returned will have the appropriated roles added. If not, then the
// "guest" User and its roles are used.
func (a *App) withUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u User
		// try to fetch the user from the request
		if token := sessions.FromRequest(r, a.cookies.name); token == "" {
			u.handle = "guest"
			u.roles = []string{"guest"}
		} else {
			u.roles = []string{"authenticated", "admin"}
		}
		log.Printf("%s %s: user %q\n", r.Method, r.URL, u.handle)
		ctx := context.WithValue(r.Context(), userContextKey("user"), u)
		// serve with the user in the context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
