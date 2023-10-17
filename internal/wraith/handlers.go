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
	nfh := a.notFound()

	tInternalError := &templateHandler{}
	if err := tInternalError.AddFiles(a.templates.path, "layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "internal_error"); err != nil {
		panic(fmt.Sprintf("[app] assetServer: %v", err))
	}

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
			a.render(w, r, tInternalError, payload)
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
			nfh(w, r)
			return
		}

		// we never want to give a directory listing, so change raw directory request to fetch the index.html instead.
		if sb.IsDir() {
			file = filepath.Join(file, "index.html")
			sb, err = os.Stat(file)
			if err != nil {
				nfh(w, r)
				return
			}
		}

		// only serve regular files (this avoids serving a directory named index.html)
		if !sb.Mode().IsRegular() {
			nfh(w, r)
			return
		}

		// pretty sure that we have a regular file at this point.
		rdr, err := os.Open(file)
		if err != nil {
			nfh(w, r)
			return
		}
		defer rdr.Close()

		http.ServeContent(w, r, file, sb.ModTime(), rdr)
	}
}

func (a *App) getAuthCallback() http.HandlerFunc {
	nfh := a.notFound()
	return func(w http.ResponseWriter, r *http.Request) {
		var provider authn.Provider
		name := way.Param(r.Context(), "provider")
		for _, p := range a.authn {
			if name == p.Code() {
				provider = p
				break
			}
		}
		if provider == nil {
			nfh(w, r)
			return
		}

		authorization, err := provider.ProcessCallback(r)
		if err != nil {
			a.internalError(w, r, err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/users/%s", authorization.Id), http.StatusTemporaryRedirect)
	}
}

func (a *App) getGames() http.HandlerFunc {
	t := &templateHandler{}
	if err := t.AddFiles(a.templates.path, "layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "games"); err != nil {
		panic(fmt.Sprintf("[app] getGames: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		payload := Payload{Site: a.templates.site}
		payload.Page.Title = "Games"
		a.render(w, r, t, payload)
	}
}

func (a *App) getGuest() http.HandlerFunc {
	t := &templateHandler{}
	if err := t.AddFiles(a.templates.path, "layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "guest"); err != nil {
		panic(fmt.Sprintf("[app] getGuest: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		payload := Payload{Site: a.templates.site}
		payload.Page.Title = "Guest"
		a.render(w, r, t, payload)
	}
}

func (a *App) getIndex() func(w http.ResponseWriter, r *http.Request) {
	t, err := a.newTemplate("layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "index")
	if err != nil {
		panic(fmt.Sprintf("[app] getIndex: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := a.currentUser(r)
		log.Printf("%s %s: user %+v\n", r.Method, r.URL, user)
		if user.IsAuthenticated() {
			http.Redirect(w, r, fmt.Sprintf("/users/%s", user.Id()), http.StatusSeeOther)
		}
		payload := Payload{Site: a.templates.site}
		payload.Site.NavBar = NavBarData{Links: []LinkData{
			{Text: "Documentation", Url: "/docs"},
			{Text: "Sign Up", Url: "/signup"},
			{Text: "Sign In", Url: "/signin"},
		}}
		t.render(w, r, payload)
	}
}

func (a *App) getSignIn() func(w http.ResponseWriter, r *http.Request) {
	t, err := a.newTemplate("layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "signin")
	if err != nil {
		panic(fmt.Sprintf("[app] getSignIn: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     a.cookies.name,
			Path:     "/",
			HttpOnly: a.cookies.httpOnly,
			Secure:   a.cookies.secure,
		})

		payload := Payload{Site: a.templates.site}
		payload.Site.NavBar = NavBarData{Links: []LinkData{
			{Text: "Home", Url: "/"},
			{Text: "Documentation", Url: "/docs"},
		}}
		t.render(w, r, payload)
	}
}

func (a *App) getSignOut() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     a.cookies.name,
			Path:     "/",
			HttpOnly: a.cookies.httpOnly,
			Secure:   a.cookies.secure,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (a *App) getUsers() http.HandlerFunc {
	t, err := a.newTemplate("layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "users")
	if err != nil {
		panic(fmt.Sprintf("[app] getUsers: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := a.currentUser(r)
		log.Printf("%s %s: user %+v\n", r.Method, r.URL, user)
		if !user.IsAuthenticated() {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else if !user.IsAdmin() {
			http.Redirect(w, r, fmt.Sprintf("/users/%s", user.Id()), http.StatusSeeOther)
			return
		}
		payload := Payload{Site: a.templates.site}
		payload.Site.NavBar = NavBarData{Links: []LinkData{
			{Text: "Documentation", Url: "/docs"},
			{Text: "Sign Out", Url: "/signout"},
		}}
		t.render(w, r, payload)
	}
}

func (a *App) getUsersId() http.HandlerFunc {
	t, err := a.newTemplate("layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "users")
	if err != nil {
		panic(fmt.Sprintf("[app] getUsersId: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := a.currentUser(r)
		log.Printf("%s %s: user %+v\n", r.Method, r.URL, user)
		if !user.IsAuthenticated() {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		payload := Payload{Site: a.templates.site}
		payload.Site.NavBar = NavBarData{Links: []LinkData{
			{Text: "Documentation", Url: "/docs"},
			{Text: "Sign Out", Url: "/signout"},
		}}
		t.render(w, r, payload)
	}
}

func (a *App) getVersion() http.HandlerFunc {
	t := &templateHandler{}
	if err := t.AddFiles(a.templates.path, "layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "version"); err != nil {
		panic(fmt.Sprintf("[app] getVersion: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		payload := struct {
			Version string
		}{
			Version: a.version,
		}
		a.render(w, r, t, payload)
	}
}

func (a *App) getWelcome() http.HandlerFunc {
	t, err := a.newTemplate("layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "welcome")
	if err != nil {
		panic(fmt.Sprintf("[app] getWelcome: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		payload := Payload{Site: a.templates.site}
		payload.Page.Title = "Welcome"
		payload.Site.NavBar = NavBarData{Links: []LinkData{
			{Text: "Documentation", Url: "/docs"},
		}}
		a.render(w, r, t, payload)
	}
}

func (a *App) internalError(w http.ResponseWriter, r *http.Request, err error) {
	// log.Printf("%s %s: %v\n", r.Method, r.URL, err)
	// http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

	t := &templateHandler{}
	if err := t.AddFiles(a.templates.path, "layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "internal_error"); err != nil {
		panic(fmt.Sprintf("[app] internalError: %v", err))
	}

	payload := struct {
		Method string
		URL    string
		Error  error
	}{
		Method: r.Method,
		URL:    r.URL.Path,
		Error:  err,
	}
	a.render(w, r, t, payload)
}

func (a *App) notFound() http.HandlerFunc {
	t, err := a.newTemplate("layout", "head", "site_header_default", "site_navbar_default", "site_footer_default", "not_found")
	if err != nil {
		panic(fmt.Sprintf("[app] notFound: %v", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s: not found\n", r.Method, r.URL)
		payload := Payload{Site: a.templates.site}
		payload.Site.NavBar = NavBarData{Links: []LinkData{
			{Text: "Home", Url: "/"},
		}}
		payload.Page.Title = "Not Found"
		payload.Content = struct {
			Method string
			URL    string
		}{
			Method: r.Method,
			URL:    r.URL.Path,
		}
		w.WriteHeader(http.StatusNotFound)
		t.render(w, r, payload)
	}
}

func (a *App) postAuthLogin() http.HandlerFunc {
	nfh := a.notFound()
	return func(w http.ResponseWriter, r *http.Request) {
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
			nfh(w, r)
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
}

func (a *App) postSignIn() func(w http.ResponseWriter, r *http.Request) {
	type input struct {
		Email    string
		Password string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Hx-Request") != "true" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		// _ = r.ParseForm()
		// log.Printf("%s %s: form %+v\n", r.Method, r.URL, r.Form)
		input := input{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}
		// log.Printf("%s %s: input %+v\n", r.Method, r.URL, input)
		var user User
		if input.Email == "user@example.com" && input.Password == "password" {
			user.id, user.handle, user.roles = "1", "admin", []string{"authenticated", "admin"}
		}
		if user.id == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     a.cookies.name,
			Path:     "/",
			Value:    "authenticated",
			HttpOnly: a.cookies.httpOnly,
			Secure:   a.cookies.secure,
		})
		http.Redirect(w, r, fmt.Sprintf("/users/%s", user.id), http.StatusSeeOther)
	}
}
