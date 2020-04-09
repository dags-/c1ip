package main

import (
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	port = flag.Int("port", 8088, "Server port")
	addr = flag.String("url", "", "Website url")
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

func main() {
	flag.Parse()
	_ = os.Mkdir("temp", os.ModePerm)
	_ = os.Mkdir("video", os.ModePerm)
	l, e := net.Listen("tcp", fmt.Sprintf("localhost:%v", *port))
	if e != nil {
		panic(e)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", serve)
	mux.HandleFunc("/upload", upload)
	log.Println("http://" + l.Addr().String())
	e = http.Serve(l, mux)
	if e != nil {
		panic(e)
	}
}

func serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		files, e := listFiles()
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
		e = home.Execute(w, files)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	} else if strings.LastIndexAny(r.URL.Path, "./") > 0 {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			http.ServeFile(w, r, filepath.Join("video", r.URL.Path))
		} else {
			http.ServeFile(w, r, filepath.Join("html", r.URL.Path))
		}
	} else {
		e := page.Execute(w, *addr+r.URL.Path)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, filepath.Join("html", r.URL.Path))
		return
	}

	if r.Method == http.MethodPost {
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

		go convert(src, dest)

		http.Redirect(w, r, "/"+name, 302)
		return
	}
}

func convert(src, dest string) {
	c := exec.Command(
		"ffmpeg",
		"-i", src,
		"-c:v", "libx264",
		"-b:v", "4M", "-maxrate", "4M", "-bufsize", "1M",
		"-vf", "scale=-1:720:flags=lanczos",
		"-preset", "fast",
		dest,
	)
	logErr(c.Run())
	logErr(os.Remove(src))
}

func listFiles() ([]os.FileInfo, error) {
	files, e := ioutil.ReadDir("video")
	if e != nil {
		return nil, e
	}
	sort.Slice(files, func(i, j int) bool {
		return files[j].ModTime().Unix() < files[i].ModTime().Unix()
	})
	return files, nil
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
	binary.LittleEndian.PutUint32(data, uint32(time.Now().Nanosecond()))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(data)
}
