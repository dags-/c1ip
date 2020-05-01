package app

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	page = template.Must(template.ParseFiles("html/video/template.html"))

	home = template.Must(template.New("template.html").Funcs(template.FuncMap{
		"name": func(s string) string {
			return strings.TrimSuffix(s, ".mp4")
		},
		"date": func(t time.Time) string {
			return t.Format(time.Stamp)
		},
	}).ParseFiles("html/home/template.html"))
)

type Action struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func (a *App) serve(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			if a.auth.check(w, r) {
				a.list(w, r)
				return
			}
			return
		}

		if strings.LastIndexAny(r.URL.Path, "./") > 0 {
			if strings.HasSuffix(r.URL.Path, ".mp4") {
				http.ServeFile(w, r, filepath.Join("video", r.URL.Path))
			} else {
				http.ServeFile(w, r, filepath.Join("html", r.URL.Path))
			}
			return
		}

		e := page.Execute(w, a.addr+r.URL.Path)

		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	}

	if r.Method == http.MethodDelete {
		if a.auth.test(r) {
			path := filepath.Join("video", r.URL.Path+".mp4")
			if _, e := os.Stat(path); e == nil {
				e := os.Remove(path)
				if e != nil {
					http.Error(w, e.Error(), http.StatusInternalServerError)
				}
			}
		} else {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
		}
	}
}

func (a *App) list(w http.ResponseWriter, r *http.Request) {
	files, e := listFiles()
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}

	e = home.Execute(w, files)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

func (a *App) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		e := r.ParseMultipartForm(500 << 20)
		if e != nil {
			http.Error(w, e.Error(), http.StatusNotFound)
			return
		}

		in, _, e := r.FormFile("upload")
		if e != nil {
			http.Error(w, e.Error(), http.StatusNotFound)
			return
		}
		defer doClose("Close input", in)

		name := nextId()
		src := filepath.Join("temp", name)
		out, e := os.Create(src)
		if e != nil {
			http.Error(w, e.Error(), http.StatusNotFound)
			return
		}
		_, _ = io.Copy(out, in)
		doClose("Close output", out)

		temp := filepath.Join("temp", name+".mp4")
		dest := filepath.Join("video", name+".mp4")
		go convert(src, temp, dest, a.debug)

		http.Redirect(w, r, "/", 302)
		return
	}
}
