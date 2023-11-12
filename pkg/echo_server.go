package server

import (
	"com.github/salpreh/devserver/pkg/servers"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	requestHeaderPrefix string = "X-Req-"
	requestMethodHeader        = requestHeaderPrefix + "Method"
	requestPathHeader          = requestHeaderPrefix + "Path"
)

const defaultStatusCode int = http.StatusOK

func CreateEchoServer(port int) {
	http.HandleFunc("/", echoHandler)

	log.Printf("Starting server on port: %d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Panicf("Unable to start server on port %d: %v", port, err)
	}
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: [%s] %s", r.Method, r.URL.Path)

	resStatusCode := servercommons.GetResponseCode(r, defaultStatusCode)

	for key, values := range r.Header {
		headerKey := requestHeaderPrefix + key
		for _, value := range values {
			w.Header().Set(headerKey, value)
		}
	}
	for key, value := range getAdditionalHeaders(r) {
		w.Header().Set(key, value)
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

func getAdditionalHeaders(r *http.Request) map[string]string {
	return map[string]string{
		requestMethodHeader: r.Method,
		requestPathHeader:   r.URL.Path,
	}
}
