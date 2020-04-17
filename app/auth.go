package app

import (
	"net/http"
	"time"
)

type Auth struct {
	user string
	pass string
}

func (a *Auth) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "html/login/index.html")
		return
	}

	if r.Method == http.MethodPost {
		e := r.ParseMultipartForm(1 << 20)
		if e != nil {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}

		user := r.FormValue("user")
		pass := r.FormValue("pass")
		target := r.FormValue("target")
		if user != a.user && pass != a.pass {
			http.Error(w, "incorrect login", http.StatusNetworkAuthenticationRequired)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Secure:  true,
			Path:    "/",
			Name:    "auth",
			Value:   "true",
			Expires: time.Now().Add(time.Hour * 24 * 30),
		})

		http.Redirect(w, r, target, 302)
	}
}

func (a *Auth) check(w http.ResponseWriter, r *http.Request) bool {
	token, e := r.Cookie("auth")
	if e == nil && token.Value == "true" {
		return true
	} else {
		http.Redirect(w, r, "/login", 302)
		return false
	}
}

func (a *Auth) wrap(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.check(w, r) {
			handler(w, r)
		}
	}
}
