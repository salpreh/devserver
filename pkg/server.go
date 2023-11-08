package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

const (
	responseCodeHeader string = "X-Response-Code"
)

const defaultStatusCode int = 200
const requestHeaderPrefix string = "X-Req-"

func CreateEchoServer(port int) {
	log.Printf("Starting server on port: %d", port)

	http.HandleFunc("/", echoHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Panicf("Unable to start server on port %d: %v", port, err)
	}
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: [%s] %s", r.Method, r.URL.Path)

	resStatusCode := defaultStatusCode
	clientStatusCode := r.Header.Get(responseCodeHeader)
	r.Header.Del(responseCodeHeader)
	if clientStatusCode != "" {
		resStatusCode, _ = strconv.Atoi(clientStatusCode)
	}

	for key, values := range r.Header {
		headerKey := requestHeaderPrefix + key
		for _, value := range values {
			w.Header().Set(headerKey, value)
		}
	}
	w.WriteHeader(resStatusCode)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v\n", err)
		return
	}

	_, err = w.Write(body)
	if err != nil {
		log.Printf("Error writing response body: %v\n", err)
	}
}
