package main

import (
	bsdc "code-samples/blob-store-data-converter"
	"code-samples/blob-store-data-converter/blobstore"
	"flag"
	"go.temporal.io/sdk/converter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

var portFlag int

func init() {
	flag.IntVar(&portFlag, "port", 8081, "Port to listen on")
}

func main() {
	flag.Parse()

	// This example codec server does not support varying config per namespace,
	// decoding for the Temporal Web UI or oauth.
	// For a more complete example of a codec server please see the codec-server sample at:
	// ../../codec-server.
	handler := converter.NewPayloadCodecHTTPHandler(
		bsdc.NewBaseCodec(blobstore.NewClient()),
	)

	srv := &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(portFlag),
		Handler: newCORSHTTPHandler(handler),
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

// newCORSHTTPHandler wraps a HTTP handler with CORS support
func newCORSHTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8233")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Namespace")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
