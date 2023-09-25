// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"net/http"
)

func (a *App) getIndex(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, nil, "layout", "navbar", "index")
}

func (a *App) getVersion(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Version string
	}{
		Version: a.version,
	}
	a.render(w, r, payload, "layout", "navbar", "version")
}

func (a *App) internalError(w http.ResponseWriter, r *http.Request, err error) {
	// log.Printf("%s %s: %v\n", r.Method, r.URL, err)
	// http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

	payload := struct {
		Method string
		URL    string
		Error  error
	}{
		Method: r.Method,
		URL:    r.URL.Path,
		Error:  err,
	}
	a.render(w, r, payload, "layout", "navbar", "internal_error")
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Method string
		URL    string
	}{
		Method: r.Method,
		URL:    r.URL.Path,
	}
	a.render(w, r, payload, "layout", "navbar", "not_found")
}
