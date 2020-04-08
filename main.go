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
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	_ = os.Mkdir("temp", os.ModePerm)
	_ = os.Mkdir("video", os.ModePerm)
	port := flag.Int("port", 0, "Server port")
	flag.Parse()
	l, e := net.Listen("tcp", fmt.Sprintf("localhost:%v", *port))
	if e != nil {
		panic(e)
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("video")))
	mux.Handle("/upload/", http.StripPrefix("/upload/", http.FileServer(http.Dir("html"))))
	mux.HandleFunc("/upload/file", upload)
	log.Println("http://" + l.Addr().String() + "/upload")
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

	name := nextId()
	src := filepath.Join("temp", name)
	dest := filepath.Join("video", name+".mp4")
	out, e := os.Create(src)
	if e != nil {
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	_, _ = io.Copy(out, in)
	doClose(out)

	logErr(convert(src, dest))
	logErr(os.Remove(src))

	http.Redirect(w, r, "/"+name+".mp4", 302)
}

func convert(src, dest string) error {
	c := exec.Command(
		"ffmpeg",
		"-i", src,
		"-c:v", "libx264",
		"-b:v", "4M", "-maxrate", "4M", "-bufsize", "1M",
		"-vf", "scale=-1:720:flags=lanczos",
		"-preset", "fast",
		dest,
	)
	er, _ := c.StderrPipe()
	go func() {
		defer doClose(er)
		_, _ = io.Copy(os.Stdout, er)
	}()
	return c.Run()
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
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(time.Now().Unix()))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(data)
}
