package app

import (
	"crypto/sha512"
	"encoding/base64"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sync"
	"time"
)

type Auth struct {
	lock     *sync.RWMutex
	secure   bool
	salt     string
	token    string
	sessions map[string]time.Time
}

func auth(user, pass, salt string, debug bool) *Auth {
	return &Auth{
		lock:     &sync.RWMutex{},
		secure:   !debug && user != "" && pass != "",
		salt:     salt,
		token:    hash(user, pass, salt),
		sessions: map[string]time.Time{},
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

		sessionToken := nextToken()
		timestamp := time.Now().Add(time.Hour * 24 * 30)
		a.storeToken(sessionToken, timestamp)

		http.SetCookie(w, &http.Cookie{
			Secure:  a.secure,
			Path:    "/",
			Name:    "session",
			Value:   sessionToken,
			Expires: timestamp,
		})

		http.Redirect(w, r, target, 302)
	}
}

func (a *Auth) check(w http.ResponseWriter, r *http.Request) bool {
	token, e := r.Cookie("session")
	if e == nil && a.validateToken(token.Value) {
		return true
	} else {
		http.Redirect(w, r, "/login", 302)
		return false
	}
}

func (a *Auth) test(r *http.Request) bool {
	token, e := r.Cookie("auth")
	if e == nil && a.validateToken(token.Value) {
		return true
	}
	return false
}

func (a *Auth) storeToken(token string, timestamp time.Time) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.sessions[token] = timestamp
}

func (a *Auth) validateToken(token string) bool {
	a.lock.RLock()
	timestamp, ok := a.sessions[token]
	a.lock.RUnlock()

	if !ok {
		return false
	}

	if !time.Now().Before(timestamp) {
		a.lock.Lock()
		delete(a.sessions, token)
		a.lock.Unlock()
		return false
	}

	return true
}

func (a *Auth) manageSessions() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	const empty = ""
	var size = 0
	var queue []string

	for range ticker.C {
		a.lock.Lock()
		queue, size = collectExpired(a.sessions, queue)
		for i := 0; i < size; i++ {
			delete(a.sessions, queue[i])
			queue[i] = empty
		}
		a.lock.Unlock()
		log.Printf("Purged %v sessions\n", size)
	}
}

func collectExpired(sessions map[string]time.Time, queue []string) ([]string, int) {
	i := 0
	now := time.Now()
	length := len(queue)
	for k, v := range sessions {
		if now.After(v) {
			if i < length {
				queue[i] = k
			} else {
				queue = append(queue, k)
			}
			i++
		}
	}
	return queue, i
}

func (a *Auth) wrap(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.check(w, r) {
			handler(w, r)
		}
	}
}

func nextToken() string {
	return uuid.New().String()
}
