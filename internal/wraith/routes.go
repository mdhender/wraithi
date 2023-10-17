// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"fmt"
	//"github.com/go-chi/chi/v5"
	//"github.com/go-chi/chi/v5/middleware"
	"github.com/mdhender/wraithi/internal/way"
	"net/http"
)

func (a *App) routes() http.Handler {
	//r := chi.NewRouter()
	//r.Use(middleware.Logger)
	//
	//// public routes
	//r.Group(func(r chi.Router) {
	//	r.Get("/", a.getIndex())
	//	r.Get("/welcome", a.getWelcome())
	//})
	//
	//// guest routes
	//r.Group(func(r chi.Router) {
	//	r.Use(a.mustHaveRole("guest"))
	//	r.Get("/guest", a.getGuest())
	//})
	//
	//r.NotFound(a.assetServer("", a.assets, false))
	//
	//if r != nil {
	//	return r
	//}

	wayRouter := way.NewRouter()

	// public routes
	wayRouter.HandleFunc("GET", "/", a.getIndex())
	wayRouter.HandleFunc("GET", "/guest", a.getGuest())
	wayRouter.HandleFunc("GET", "/signin", a.getSignIn())
	wayRouter.HandleFunc("POST", "/signin", a.postSignIn())
	wayRouter.HandleFunc("GET", "/signout", a.getSignOut())
	wayRouter.HandleFunc("GET", "/welcome", a.getWelcome())

	wayRouter.HandleFunc("GET", "/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})
	wayRouter.HandleFunc("GET", "/internalServerError", func(w http.ResponseWriter, r *http.Request) {
		a.internalError(w, r, fmt.Errorf("internal server error"))
	})
	wayRouter.HandleFunc("GET", "/notFound", a.notFound())
	wayRouter.HandleFunc("GET", "/version", a.getVersion())

	// authorization routes
	wayRouter.HandleFunc("GET", "/auth/callback/:provider", a.getAuthCallback())
	wayRouter.HandleFunc("POST", "/auth/login", a.postAuthLogin())

	// protected routes
	wayRouter.Handle("GET", "/games", a.authOnly(a.getGames()))
	wayRouter.Handle("GET", "/users", a.authOnly(a.getUsers()))
	wayRouter.Handle("GET", "/users/:id", a.authOnly(a.getUsersId()))

	// not found is also our assets server
	wayRouter.NotFound = a.assetServer("", a.assets, false)

	return wayRouter
}
