package server

import (
	"bytes"
	collectionutils "com.github/salpreh/devserver/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const defaultResponseCode int = http.StatusOK
const defaultResponseMethod string = "get"

const noResponseCode int = -1
const noBodyMarker string = "null"

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

func (p *Path) HasPerMethodResponses() bool {
	return p.Methods != nil
}

func (p *Path) GetResponse(httpMethod string, statusCode int) (int, json.RawMessage) {
	code := statusCode
	responses := p.GetAvailableResponses(httpMethod)
	res, exists := responses[statusCode]
	if !exists {
		code, res = responses.getDefaultResponse()
		exists = res != nil
	}

	return code, res
}

func (p *Path) GetDefaultResponse() (int, json.RawMessage) {
	responses := p.GetAnyAvailableResponses()
	if responses == nil || len(responses) == 0 {
		return defaultStatusCode, nil
	}

	return responses.getDefaultResponse()
}

func (p *Path) GetAvailableResponses(httpMethod string) Responses {
	responses := p.Responses
	if p.HasPerMethodResponses() {
		responses = p.Methods[httpMethod]
	}

	return responses
}

func (p *Path) GetAnyAvailableResponses() Responses {
	if !p.HasPerMethodResponses() {
		return p.Responses
	}

	responses := p.Methods[defaultResponseMethod]
	for _, res := range p.Methods {
		responses = res
		break
	}

	return responses
}

type Responses map[int]json.RawMessage

func (rs *Responses) getDefaultResponse() (int, json.RawMessage) {
	code := defaultStatusCode
	res, exists := (*rs)[defaultResponseCode]
	if !exists {
		for c, r := range *rs {
			code, res = c, r
			break
		}
	}

	return code, res
}
