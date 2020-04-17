package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"c1ip/app"
)

func main() {
	port := flag.Int("port", 8088, "Server port")
	addr := flag.String("addr", "", "Server address")
	user := flag.String("user", "", "User")
	pass := flag.String("pass", "", "Pass")
	flag.Parse()

	_ = os.Mkdir("temp", os.ModePerm)
	_ = os.Mkdir("video", os.ModePerm)
	l, e := net.Listen("tcp", fmt.Sprintf("localhost:%v", *port))

	if e != nil {
		panic(e)
	}

	a := app.New(&app.Config{
		User:  *user,
		Pass:  *pass,
		Addr:  *addr,
		Debug: false,
	})

	e = a.Serve(l)
	if e != nil {
		panic(e)
	}
}
