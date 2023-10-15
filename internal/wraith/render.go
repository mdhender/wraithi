// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
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
	Text string
	Url  string
}

// FooterData is the data for a page's footer.
type FooterData struct{}

func (a *App) render(w http.ResponseWriter, r *http.Request, data any, names ...string) {
	if pld, ok := data.(Payload); ok {
		pld.Site = a.templates.site
		data = pld
	}
	var files []string
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
