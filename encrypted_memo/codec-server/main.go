package main

import (
	"encrypted_memo/codec"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"go.temporal.io/sdk/converter"
)

var portFlag int

func init() {
	flag.IntVar(&portFlag, "port", 8081, "Port to listen on")
}

func main() {
	flag.Parse()

	handler := converter.NewPayloadCodecHTTPHandler(&codec.Codec{}, converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}))

	handler = corsWrap(handler)

	srv := &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(portFlag),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-sigCh:
		_ = srv.Close()
	case err := <-errCh:
		log.Fatal(err)
	}
}

func corsWrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Namespace")
		w.Header().Set("Access-Control-Max-Age", "600")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
