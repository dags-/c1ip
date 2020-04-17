package app

import (
	"log"
	"net"
	"net/http"
)

type Config struct {
	User  string
	Pass  string
	Addr  string
	Debug bool
}

type App struct {
	auth  *Auth
	addr  string
	debug bool
}

func New(c *Config) *App {
	return &App{
		auth: &Auth{
			user: c.User,
			pass: c.Pass,
		},
		addr:  c.Addr,
		debug: c.Debug,
	}
}

func (a *App) Serve(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.serve)
	mux.HandleFunc("/login", a.auth.login)
	mux.HandleFunc("/upload", a.auth.wrap(a.upload))
	log.Println("http://" + l.Addr().String())
	return http.Serve(l, mux)
}
