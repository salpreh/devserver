package servercommons

import (
	"bytes"
	collectionutils "com.github/salpreh/devserver/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const defaultResponseCode int = http.StatusOK
const defaultResponseMethod string = "get"
const (
	responsesConfigKey string = "responses"
	headersConfigKey   string = "headers"
)

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
		returnStatusCode := GetResponseCode(r, defaultStatusCode)
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

type ImportedMockConfig struct {
	Headers map[string]string
	Paths   map[HttpPath]map[string]json.RawMessage
}

func (c *ImportedMockConfig) LoadConfig() *MockConfig {
	paths := make(map[HttpPath]Path)
	for path, data := range c.Paths {
		rawHeaders := data[headersConfigKey]
		var headers map[string]string
		if rawHeaders != nil {
			if err := json.Unmarshal(rawHeaders, &headers); err != nil {
				log.Panicf("Unable to process headers for path %s in config: %v", path, err)
			}
		}

		rawResponses, exists := data[responsesConfigKey]
		if exists { // If responses keyword exists we build path with common responses
			var responses Responses
			if err := json.Unmarshal(rawResponses, &responses); err != nil {
				log.Panicf("Unable to process common responses for path %s in config: %v", path, err)
			}
			paths[path] = Path{
				headers,
				responses,
				nil,
			}
		} else { // Otherwise build path with per method responses
			delete(data, headersConfigKey)
			methods := make(map[string]Responses)
			for method, rawResponses := range data {
				var responses map[string]Responses
				if err := json.Unmarshal(rawResponses, &responses); err != nil {
					log.Panicf("Unable to process responses for path %s, method %s: %v", path, method, err)
				}
				methods[method] = responses[responsesConfigKey]
			}
			paths[path] = Path{
				headers,
				nil,
				methods,
			}
		}
	}

	return &MockConfig{
		c.Headers,
		paths,
	}
}

type MockConfig struct {
	Headers map[string]string
	Paths   map[HttpPath]Path
}

type Path struct {
	Headers   map[string]string
	Responses Responses
	Methods   map[string]Responses
}

func (p *Path) GetResponseByCode(statusCode int) (json.RawMessage, bool) {
	res, exists := p.Responses[statusCode]
type Responses map[int]json.RawMessage
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
