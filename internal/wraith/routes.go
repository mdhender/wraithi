// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"fmt"
	"github.com/mdhender/wraithi/internal/way"
	"net/http"
)

func (a *App) Routes() http.Handler {
	r := way.NewRouter()

	// public routes
	r.HandleFunc("GET", "/", a.getIndex)
	r.HandleFunc("GET", "/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})
	r.HandleFunc("GET", "/internalServerError", func(w http.ResponseWriter, r *http.Request) {
		a.internalError(w, r, fmt.Errorf("internal server error"))
	})
	r.HandleFunc("GET", "/notFound", a.notFound)
	r.HandleFunc("GET", "/version", a.getVersion)

	return r
}
