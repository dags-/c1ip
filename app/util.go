package app

import (
	"encoding/base64"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

func nextId() string {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(time.Now().Nanosecond()))
	token := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(data)
	token = strings.Replace(token, "-", "", -1)
	return token
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

func doClose(name string, c io.Closer) {
	if c != nil {
		logErr(name, c.Close())
	}
}

func logErr(name string, e error) bool {
	if e != nil {
		log.Println(name+":", e)
		return true
	}
	return false
}
