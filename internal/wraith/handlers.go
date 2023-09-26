// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"fmt"
	"github.com/mdhender/wraithi/internal/authn"
	"github.com/mdhender/wraithi/internal/way"
	"log"
	"net/http"
	"strings"
)

func (a *App) getAuthCallback(w http.ResponseWriter, r *http.Request) {
	var provider authn.Provider
	name := way.Param(r.Context(), "provider")
	for _, p := range a.authn {
		if name == p.Code() {
			provider = p
			break
		}
	}
	if provider == nil {
		a.notFound(w, r)
		return
	}

	authorization, err := provider.ProcessCallback(r)
	if err != nil {
		a.internalError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/users/%s", authorization.Id), http.StatusTemporaryRedirect)
}

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

func (a *App) postAuthLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s: entered\n", r.Method, r.URL)

	// convert provider to lower case to compare against the code value.
	name := strings.ToLower(r.FormValue("provider"))
	log.Printf("%s %s: name %q\n", r.Method, r.URL, name)

	var provider authn.Provider
	for _, p := range a.authn {
		log.Printf("%s %s: p %v\n", r.Method, r.URL, p)
		if name == p.Code() {
			provider = p
			break
		}
	}
	if provider == nil {
		log.Printf("%s %s: provider is nil\n", r.Method, r.URL)
		a.notFound(w, r)
		return
	}
	url, err := provider.LoginURL()
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	log.Printf("%s %s: url %q\n", r.Method, r.URL, url)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
