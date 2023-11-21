package server

import (
	"bytes"
	"com.github/salpreh/devserver/pkg/servers/contracts"
	collectionutils "com.github/salpreh/devserver/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const noResponseCode int = -1

func CreateMockServerWithContract(port int, contractFile string, mockConfigFile string) {
	contractConfig := contracts.LoadContractMockConfig(contractFile)
	mockConfig := parseConfigFile(mockConfigFile)
	contractConfig.MergeConfig(mockConfig)

	createMockServer(port, contractConfig)
}

func CreateMockServer(port int, mockConfigFile string) {
	config := parseConfigFile(mockConfigFile)
	createMockServer(port, config)
}

func createMockServer(port int, config *MockConfig) {
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

	var config ImportedMockConfig
	e = json.Unmarshal(fileContent, &config)
	if e != nil {
		log.Panicf("Unable to parse config file %v", e)
	}

	return config.LoadConfig()
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

		returnStatusCode := GetResponseCode(r, noResponseCode)
		code, response := pathConfig.GetResponse(strings.ToLower(r.Method), returnStatusCode)
		if response == nil {
			var code int
			code, response = pathConfig.GetDefaultResponse()
			log.Printf("Using %d response as default for hadler %s", code, path)
		}

		if returnStatusCode == noResponseCode {
			returnStatusCode = code
		}
		w.WriteHeader(returnStatusCode)

		if string(response) == noBodyMarker {
			return
		}

		var minifiedResponse bytes.Buffer
		json.Compact(&minifiedResponse, response)

		_, err := w.Write(minifiedResponse.Bytes())
		if err != nil {
			log.Panicf("Unable to generate handler for path %s: %v", path, err)
		}
	}

	return handler
}
