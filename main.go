package main

import (
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	_ = os.Mkdir("video", os.ModePerm)
	port := flag.Int("port", 0, "Server port")
	flag.Parse()
	l, e := net.Listen("tcp", fmt.Sprintf("localhost:%v", *port))
	if e != nil {
		panic(e)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", serve)
	mux.Handle("/manage/", http.StripPrefix("/manage/", http.FileServer(http.Dir("html"))))
	mux.HandleFunc("/manage/upload", upload)
	log.Println("Serving at:", "http://"+l.Addr().String()+"/manage/")
	e = http.Serve(l, mux)
	if e != nil {
		panic(e)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	e := r.ParseMultipartForm(200 << 20)
	if e != nil {
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	in, _, e := r.FormFile("upload")
	if e != nil {
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	defer doClose(in)
	token := nextId()
	path := filepath.Join("video", token)
	out, e := os.Create(path)
	if e != nil {
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	defer doClose(out)
	_, _ = io.Copy(out, in)
	http.Redirect(w, r, "/"+token, 301)
}

func serve(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("video", r.URL.Path)
	http.ServeFile(w, r, path)
}

func doClose(c io.Closer) {
	if c != nil {
		logErr(c.Close())
	}
}

func logErr(e error) {
	if e != nil {
		log.Println(e)
	}
}

func nextId() string {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(time.Now().UnixNano()))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(data)
}
