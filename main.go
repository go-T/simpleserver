package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
	"syscall"
)

var (
	flagDebug      = flag.Bool("debug", false, "debug mode")
	flagBind       = flag.String("bind", ":8080", "http listen address")
	flagResponse   = flag.String("response", "OK\n", "response body")
	flagStatusCode = flag.Int("code", 200, "response status code")
	flagNoFile     = flag.Int("nofile", 0, "ulimit -n: file descriptors")
)

func main() {
	flag.Parse()

	var (
		code     = *flagStatusCode
		debug    = *flagDebug
		response = []byte(*flagResponse)
		spliter  = []byte("\n>>>>> [request] <<<<<\n")
	)

	if *flagNoFile > 0 {
		fmt.Println("set file limit", *flagNoFile)
		nofile := uint64(*flagNoFile)
		err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{nofile, nofile})
		if err != nil {
			panic(err)
		}
	}

	log := make(chan []byte, 1024)
	go func() {
		for data := range log {
			os.Stdout.Write(data)
		}
	}()

	fmt.Println("listen", *flagBind)
	err := http.ListenAndServe(*flagBind, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if debug {
			var buf bytes.Buffer
			buf.Grow(int(r.ContentLength) + 4096)
			buf.Write(spliter)
			fmt.Fprintf(&buf, "%s %s%s\n", r.Method, r.Host, r.RequestURI)
			for key, _ := range r.Header {
				fmt.Fprintf(&buf, "%s: %s\n", key, r.Header.Get(key))
			}
			fmt.Fprintf(&buf, "\n")
			buf.ReadFrom(r.Body)
			r.Body.Close()
			log <- buf.Bytes()
		}
		w.WriteHeader(code)
		w.Write(response)
	}))
	if err != nil {
		panic(err)
	}
}
