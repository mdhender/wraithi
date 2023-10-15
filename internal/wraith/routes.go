// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"fmt"
	"github.com/mdhender/wraithi/internal/way"
	"net/http"
)

func (a *App) routes() http.Handler {
	r := way.NewRouter()

	// public routes
	r.HandleFunc("GET", "/", a.getIndex)
	r.HandleFunc("GET", "/guest", a.getGuest)
	r.HandleFunc("GET", "/welcome", a.getWelcome)

	r.HandleFunc("GET", "/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})
	r.HandleFunc("GET", "/internalServerError", func(w http.ResponseWriter, r *http.Request) {
		a.internalError(w, r, fmt.Errorf("internal server error"))
	})
	r.HandleFunc("GET", "/notFound", a.notFound)
	r.HandleFunc("GET", "/version", a.getVersion)

	// authorization routes
	r.HandleFunc("GET", "/auth/callback/:provider", a.getAuthCallback)
	r.HandleFunc("POST", "/auth/login", a.postAuthLogin)

	// protected routes
	r.Handle("GET", "/admin/users", a.adminOnly(a.getUsers()))
	r.Handle("GET", "/games", a.authOnly(a.getGames()))

	// not found is also our assets server
	r.NotFound = a.assetServer("", a.assets, false)

	return r
}
