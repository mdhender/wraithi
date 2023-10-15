// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"fmt"
	"github.com/mdhender/wraithi/internal/authn"
	"github.com/mdhender/wraithi/internal/way"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// assetServer tries to serve assets from the web root.
// if not found, forwards to the normal not found handler.
func (a *App) assetServer(prefix, root string, spa bool) http.HandlerFunc {
	log.Printf("[assets] serving %q\n", root)
	log.Println("[assets] initializing")
	defer log.Println("[assets] initialized")

	log.Printf("[assets] strip: %q\n", prefix)
	log.Printf("[assets]  root: %q\n", root)

	if sb, err := os.Stat(root); err != nil || !sb.IsDir() {
		if err == nil {
			err = fmt.Errorf("%q: is not a directory", root)
		}
		log.Printf("[assets] assets: %v\n", err)
		return func(w http.ResponseWriter, r *http.Request) {
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
	}

	spaIndex := filepath.Join(root, "index.html")
	if sb, err := os.Stat(spaIndex); err != nil || !sb.Mode().IsRegular() {
		if err == nil {
			err = fmt.Errorf("index.html is not a file")
		}
		log.Printf("[assets] spa: %v\n", err)
		spaIndex = ""
	} else {
		log.Printf("[assets] spa: %q\n", spaIndex)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		file := filepath.Join(root, path.Clean("/"+strings.TrimPrefix(r.URL.Path, prefix)))
		if a.flags.log.assets {
			log.Printf("[assets] %q\n", file)
		}

		sb, err := os.Stat(file)
		if err != nil {
			// try serving root index file for SPA routing instead
			if a.flags.spa && spaIndex != "" {
				if rdr, err := os.Open(spaIndex); err == nil {
					defer rdr.Close()
					http.ServeContent(w, r, file, sb.ModTime(), rdr)
					return
				}
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// we never want to give a directory listing, so change raw directory request to fetch the index.html instead.
		if sb.IsDir() {
			file = filepath.Join(file, "index.html")
			sb, err = os.Stat(file)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
		}

		// only serve regular files (this avoids serving a directory named index.html)
		if !sb.Mode().IsRegular() {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// pretty sure that we have a regular file at this point.
		rdr, err := os.Open(file)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		defer rdr.Close()

		http.ServeContent(w, r, file, sb.ModTime(), rdr)
	}
}

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

func (a *App) getGames() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.render(w, r, nil, "layout", "navbar", "index")
	}
}

func (a *App) getGuest(w http.ResponseWriter, r *http.Request) {
	payload := Payload{Site: a.templates.site}
	payload.Page.Title = "Guest"
	a.render(w, r, payload, "layout", "navbar", "guest")
}

func (a *App) getIndex(w http.ResponseWriter, r *http.Request) {
	payload := Payload{Site: a.templates.site}
	payload.Page.Title = "Welcome"
	a.render(w, r, payload, "layout", "navbar", "index")
}

func (a *App) getUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.render(w, r, nil, "layout", "navbar", "index")
	}
}

func (a *App) getVersion(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Version string
	}{
		Version: a.version,
	}
	a.render(w, r, payload, "layout", "navbar", "version")
}

func (a *App) getWelcome(w http.ResponseWriter, r *http.Request) {
	payload := Payload{Site: a.templates.site}
	payload.Page.Title = "Welcome"
	a.render(w, r, payload, "layout", "navbar", "welcome")
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
