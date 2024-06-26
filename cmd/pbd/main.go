package main

import (
	"context"
	"flag"
	"io"
	"log"
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
	ServerListenAddr   = flag.String("l", "127.0.0.1:1488", "Listen address")
	ServerRoot         = flag.String("d", "docs", "Root dir")
	MaxDocSize         = flag.Int64("s", 10000000, "Maximum Doc size")
	MaxDocCount        = flag.Int("x", 2000, "Maximum Doc number")
	DefaultLifetime    = flag.Duration("t", 7*24*time.Hour, "Document lifetime")
	DefaultNameLength  = flag.Int("n", 9, "Document name length")
	DefaultNameCharset = flag.String("c", "abcdefghijklmnopqrstuvwxyz0123456789", "Document name charset")
	DocCount           = 0
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
	for _, doc := range docs {
		srv.scheduleDelete(doc.Name(), *DefaultLifetime)
		DocCount++
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
	// >= because DocCount can exceed maximum when loading pre-existing docs on startup
	if DocCount >= *MaxDocCount {
		rw.WriteHeader(http.StatusInsufficientStorage)
		return
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			rw.WriteHeader(http.StatusRequestTimeout)
			return
		default:
		}
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

	f, err := os.OpenFile(srv.docPath(name), os.O_WRONLY|os.O_CREATE, 0400)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
	srv.mu.Unlock()
	n, err := io.CopyN(f, req.Body, *MaxDocSize)
	if n == 0 {
		os.Remove(srv.docPath(name))
		return
	}
	f.Close()
	srv.scheduleDelete(name, lifetime)
	DocCount++
	rw.Write([]byte(name))
}

func (srv *Server) scheduleDelete(name string, d time.Duration) {
	go func() {
		timer := time.NewTimer(d)
		select {
		case <-timer.C:
			os.Remove(srv.docPath(name))
			DocCount--
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
