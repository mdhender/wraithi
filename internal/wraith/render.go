// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package wraith

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func (a *App) render(w http.ResponseWriter, r *http.Request, data any, names ...string) {
	var files []string
	for _, name := range names {
		files = append(files, filepath.Join(a.templates, name+".gohtml"))
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
