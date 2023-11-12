package server

import (
	"bytes"
	servercommons "com.github/salpreh/devserver/pkg/servers"
	collectionutils "com.github/salpreh/devserver/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const defaultResponseCode int = http.StatusOK

func CreateMockServer(port int, mockConfigFile string) {
	config := parseConfigFile(mockConfigFile)
	handlers := generateHandlers(config)

	for path, handler := range handlers {
		log.Printf("Registering handler for path %s", path)
		http.HandleFunc(string(path), handler)
	}

	log.Printf("Starting server on port %d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Panicf("Unable to start server: %v", err)
	}
}

func parseConfigFile(configFile string) *MockConfig {
	fileContent, e := os.ReadFile(configFile)
	if e != nil {
		log.Panicf("Unable to parse config file %v", e)
	}
	log.Printf("Config file readed")

	var config MockConfig
	e = json.Unmarshal(fileContent, &config)
	if e != nil {
		log.Panicf("Unable to parse config file %v", e)
	}

	return &config
}

func generateHandlers(config *MockConfig) map[HttpPath]func(http.ResponseWriter, *http.Request) {
	handlers := make(map[HttpPath]func(http.ResponseWriter, *http.Request))
	for path, pathConfig := range config.Paths {
		handlers[path] = generateHandler(path, pathConfig, config.Headers)
	}

	return handlers
}

func generateHandler(path HttpPath, pathConfig Path, commonHeaders map[string]string) func(http.ResponseWriter, *http.Request) {
	headers := collectionutils.MergeMaps(commonHeaders, pathConfig.Headers)
	handler := func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Add(k, v)
		}
		returnStatusCode := servercommons.GetResponseCode(r, defaultStatusCode)
		w.WriteHeader(returnStatusCode)

		response, hasValue := pathConfig.GetResponseByCode(returnStatusCode)
		if !hasValue {
			var code int
			code, response = pathConfig.GetDefaultResponse()
			log.Printf("Using %d response as default for hadler %s", code, path)
		}

		var prettyResponse bytes.Buffer
		json.Indent(&prettyResponse, response, "", "\t")
		_, err := w.Write(prettyResponse.Bytes())
		if err != nil {
			log.Panicf("Unable to generate handler for path %s: %v", path, err)
		}
	}

	return handler
}

type HttpPath string

type MockConfig struct {
	Headers map[string]string
	Paths   map[HttpPath]Path
}

type Path struct {
	Headers   map[string]string
	Responses map[int]json.RawMessage
}

func (p *Path) GetResponseByCode(statusCode int) (json.RawMessage, bool) {
	res, exists := p.Responses[statusCode]
	return res, exists
}

func (p *Path) GetDefaultResponse() (int, json.RawMessage) {
	if p.Responses == nil || len(p.Responses) == 0 {
		return defaultStatusCode, nil
	}

	code := defaultStatusCode
	res, exists := p.Responses[defaultResponseCode]
	if !exists {
		for c, r := range p.Responses {
			code, res = c, r
		}
	}

	return code, res
}
