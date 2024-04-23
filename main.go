package main

import (
	"flag"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	root string
	mu   *sync.Mutex
}

var (
	ServerListenAddr   = flag.String("l", "127.0.0.1:8080", "Listen address")
	ServerRoot         = flag.String("d", "docs", "Root dir")
	DefaultLifetime    = flag.Duration("t", 7*24*time.Hour, "Document lifetime")
	DefaultNameLength  = flag.Int("n", 9, "Document name length")
	DefaultNameCharset = flag.String("c", "abcdefghijklmnopqrstuvwxyz0123456789", "Document name charset")
)

func main() {
	flag.Parse()

	srv := &Server{
		root: *ServerRoot,
		mu:   new(sync.Mutex),
	}

	// The server doesn't keep track of non-deleted docs through restarts, load
	// the existing docs by scheduling there deletion with the default lifetime.
	docs, err := os.ReadDir(srv.root)
	if err != nil {
		log.Fatal(err)
	}
	for _, doc := range(docs) {
		srv.scheduleDelete(doc.Name(), *DefaultLifetime)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{doc}", srv.getDoc)
	mux.HandleFunc("POST /", srv.postDoc)
	log.Fatal(http.ListenAndServe(*ServerListenAddr, mux))
}

func (srv *Server) docPath(name string) string {
	return path.Clean(srv.root + "/" + name)
}

func (srv *Server) postDoc(rw http.ResponseWriter, req *http.Request) {
	var lifetime time.Duration
	if headerVal := req.Header.Get("Doc-Lifetime"); headerVal != "" {
		var err error
		lifetime, err = time.ParseDuration(headerVal)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		lifetime = *DefaultLifetime
	}

	var nameCharset string
	if headerVal := req.Header.Get("Doc-Name-Charset"); headerVal != "" {
		nameCharset = headerVal
	} else {
		nameCharset = *DefaultNameCharset
	}

	var nameLen int
	if headerVal := req.Header.Get("Doc-Name-Length"); headerVal != "" {
		var err error
		nameLen, err = strconv.Atoi(headerVal)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		nameLen = *DefaultNameLength
	}

	var name string
	srv.mu.Lock()
	var k int
	for k = 0; k < int(math.Pow(float64(len(nameCharset)), float64(nameLen))); k++ {
		for i := 0; i < nameLen; i++ {
			name += string(nameCharset[rand.Intn(len(nameCharset))])
		}
		docs, err := os.ReadDir(srv.root)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		// check if the generate file name is unique
		isUnique := true
		for _, doc := range docs {
			if doc.Name() == name {
				isUnique = false
			}
		}
		if isUnique {
			break
		}
	}
	if k == int(math.Pow(float64(len(nameCharset)), float64(nameLen))) {
		// no doc name available
		rw.WriteHeader(http.StatusInsufficientStorage)
		return
	}

	f, err := os.OpenFile(srv.docPath(name), os.O_WRONLY|os.O_CREATE, 0400)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
	_, err = io.Copy(f, req.Body)
	if err != nil {
		os.Remove(srv.docPath(name))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	f.Close()
	srv.mu.Unlock()
	srv.scheduleDelete(name, lifetime)
	rw.Write([]byte(name))
}

func (srv *Server) scheduleDelete(name string, d time.Duration) {
	go func() {
		timer := time.NewTimer(d)
		select {
		case <-timer.C:
			os.Remove(srv.docPath(name))
		}
	}()
}

func (srv *Server) getDoc(rw http.ResponseWriter, req *http.Request) {
	name := req.PathValue("doc")
	f, err := os.Open(srv.docPath(name))
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	_, err = io.Copy(rw, f)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
