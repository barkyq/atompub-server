package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"log"
	"net/http"
	"sync"
)

var listen_address_flag = flag.String("listen", "127.0.0.1:8357", "listen address")
var gitdir_flag = flag.String("gitdir", ".atompub", "git directory")

func main() {
	flag.Parse()
	h := &Handler{
		B:     NewBackend(NewBillyStorer(*gitdir_flag)),
		gzw:   gzip.NewWriter(nil),
		mutex: new(sync.Mutex),
		buf:   bytes.NewBuffer(nil),
		bw:    bufio.NewWriter(nil),
	}
	log.Printf("Starting AtomPub server on %s", *listen_address_flag)
	log.Fatal(http.ListenAndServe(*listen_address_flag, h))
}
