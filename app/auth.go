package app

import (
	"crypto/sha512"
	"encoding/base64"
	"net/http"
	"time"
)

type Auth struct {
	secure bool
	salt   string
	token  string
}

func auth(user, pass, salt string) *Auth {
	return &Auth{
		secure: user != "" && pass != "",
		salt:   salt,
		token:  hash(user, pass, salt),
	}
}

func hash(user, pass, salt string) string {
	hash := sha512.New()
	hash.Write([]byte(user))
	hash.Write([]byte(pass))
	hash.Write([]byte(salt))
	result := hash.Sum(nil)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(result)
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
		token := hash(user, pass, a.salt)
		if token != a.token {
			http.Error(w, "incorrect login", http.StatusNetworkAuthenticationRequired)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Secure:  a.secure,
			Path:    "/",
			Name:    "auth",
			Value:   token,
			Expires: time.Now().Add(time.Hour * 24 * 30),
		})

		http.Redirect(w, r, target, 302)
	}
}

func (a *Auth) check(w http.ResponseWriter, r *http.Request) bool {
	token, e := r.Cookie("auth")
	if e == nil && token.Value == a.token {
		return true
	} else {
		http.Redirect(w, r, "/login", 302)
		return false
	}
}

func (a *Auth) test(r *http.Request) bool {
	token, e := r.Cookie("auth")
	if e == nil && token.Value == a.token {
		return true
	}
	return false
}

func (a *Auth) wrap(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.check(w, r) {
			handler(w, r)
		}
	}
}
