// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

// Payload is the data for a rendered page.
type Payload struct {
	Site    SiteData
	Page    PageData
	Content any
	Footer  FooterData
}

// SiteData is data for the site.
type SiteData struct {
	Title     string
	Copyright struct {
		Author string
		Year   string
	}
	NavBar      NavBarData
	UseCDN      bool // if true, serve css and js from CDN
	UseOutliner bool // if true, add outlines for debugging
	Version     string
}

// PageData is the data for a page.
type PageData struct {
	Title   string
	NavBar  NavBarData
	Content any // generic data
	Footer  FooterData
}

// NavBarData is the data for a page's navigation bar.
type NavBarData struct {
	Links []LinkData
}

// LinkData is link data
type LinkData struct {
	Text   string
	Url    string
	Method string
	Target string
}

// FooterData is the data for a page's footer.
type FooterData struct{}

// templateHandler implements a handler for loading, compiling, and serving a template.
// from Mat Ryer's Go Programming Blueprints.
type templateHandler struct {
	sync.Mutex
	once    sync.Once
	files   []string           // template files to load
	headers [][]string         // response headers
	t       *template.Template // represents a single template
}

func (t *templateHandler) AddFiles(root string, files ...string) error {
	t.Lock()
	defer t.Unlock()
	// log.Printf("[templates] root   %q\n", root)
	// load requested templates
	for _, file := range files {
		// log.Printf("[templates] adding %q\n", filepath.Join(root, file+".gohtml"))
		t.files = append(t.files, filepath.Join(root, file+".gohtml"))
	}
	_, err := template.ParseFiles(t.files...)
	if err != nil {
		log.Printf("[templates] error %v\n", err)
	}
	return err
}

func (t *templateHandler) render(w http.ResponseWriter, r *http.Request, data any) {
	buf := &bytes.Buffer{}
	var err error
	t.t, err = template.ParseFiles(t.files...)
	if err != nil {
		log.Printf("%s %s: render: parse: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if err = t.t.ExecuteTemplate(buf, "layout", data); err != nil {
		log.Printf("%s %s: render: execute: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	for _, kvp := range t.headers {
		key, value := kvp[0], kvp[1]
		w.Header().Set(key, value)
	}
	_, _ = w.Write(buf.Bytes())
}

func (a *App) render(w http.ResponseWriter, r *http.Request, t *templateHandler, data any) {
	//if p, ok := data.(Payload); ok {
	//	log.Printf("%s %s: render: content %+v\n", r.Method, r.URL, p.Content)
	//}

	w.Header().Set("Wraith-Version", a.version)

	var err error
	t.t, err = template.ParseFiles(t.files...)
	if err != nil {
		log.Printf("%s %s: render: parse: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := t.t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("%s %s: render: execute: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (a *App) oldRender(w http.ResponseWriter, r *http.Request, data any, names ...string) {
	if pld, ok := data.(Payload); ok {
		pld.Site = a.templates.site
		data = pld
	}
	var files []string
	// load default templates
	for _, name := range []string{"layout", "head", "site_header_default", "site_navbar_default", "site_footer_default"} {
		files = append(files, filepath.Join(a.templates.path, name+".gohtml"))
	}
	// load requested templates
	for _, name := range names {
		files = append(files, filepath.Join(a.templates.path, name+".gohtml"))
	}

	w.Header().Set("Wraith-Version", a.version)

	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("%s %s: render: parse: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("%s %s: render: execute: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (a *App) newTemplate(files ...string) (*templateHandler, error) {
	t := &templateHandler{}
	t.headers = append(t.headers, []string{"Wraith-Version", a.version})
	for _, file := range files {
		t.files = append(t.files, filepath.Join(a.templates.path, file+".gohtml"))
	}
	if _, err := template.ParseFiles(t.files...); err != nil {
		log.Printf("[templates] parse: %v\n", err)
		return nil, err
	}
	return t, nil
}
