package app

import (
	"log"
	"net"
	"net/http"
)

type Config struct {
	User  string
	Pass  string
	Salt  string
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
		auth:  auth(c.User, c.Pass, c.Salt, c.Debug),
		addr:  c.Addr,
		debug: c.Debug,
	}
}

func (a *App) Serve(l net.Listener) error {
	// manages the session cache on separate routine
	go a.auth.manageSessions()

	mux := http.NewServeMux()
	mux.HandleFunc("/", a.serve)
	mux.HandleFunc("/login", a.auth.login)
	mux.HandleFunc("/upload", a.auth.wrap(a.upload))
	log.Println("http://" + l.Addr().String())
	return http.Serve(l, mux)
}
