package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	Lifetime    = flag.Duration("t", 7*24*time.Hour, "Document lifetime")
	NameLength  = flag.Int("n", 16, "Document name length")
	NameCharset = flag.String("c", "abcdefghijklmnopqrstuvwxyz0123456789", "Document name charset")
)

func main() {
	flag.Parse()

	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if len(flag.Args()) != 1 {
		log.Fatal("usage: pb [-tnc] url")
	}
	req, err := http.NewRequest("POST", flag.Args()[0], bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Doc-Lifetime", Lifetime.String())
	req.Header.Add("Doc-Name-Charset", *NameCharset)
	req.Header.Add("Doc-Name-Length", strconv.Itoa(*NameLength))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, resp.Body)
	resp.Body.Close()
}
