package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const defaultResponseCode int = http.StatusOK
const defaultResponseMethod string = "get"

const (
	responsesConfigKey string = "responses"
	headersConfigKey   string = "headers"
)

func NewImportedMockConfig(config *MockConfig) (*ImportedMockConfig, error) {
	paths := make(map[HttpPath]map[string]json.RawMessage)
	for path, pathData := range config.Paths {
		pathConfig := make(map[string]json.RawMessage)
		if pathData.Headers != nil {
			headersEncoded, err := json.Marshal(pathData.Headers)
			if err != nil {
				return nil, fmt.Errorf("unable to encode config headers for path %s: %w", path, err)
			}
			pathConfig[headersConfigKey] = headersEncoded
		}

		if pathData.Responses != nil && len(pathData.Responses) > 0 { // If common responses we encode that to config
			responsesEncoded, err := json.Marshal(pathData.Responses)
			if err != nil {
				return nil, fmt.Errorf("unable to encode common responses for path %s: %w", path, err)
			}
			pathConfig[headersConfigKey] = responsesEncoded
		} else if pathData.Methods != nil && len(pathData.Methods) > 0 {
			for method, response := range pathData.Methods {
				js := response.ToJsonSerializable()
				responseEncoded, err := json.Marshal(js)
				if err != nil {
					return nil, fmt.Errorf("unable to encode method responses for path [%s]%s: %w", method, path, err)
				}
				pathConfig[method] = responseEncoded
			}
		}

		paths[path] = pathConfig
	}

	return &ImportedMockConfig{
		Headers: config.Headers,
		Paths:   paths,
	}, nil
}

type HttpPath string

type ImportedMockConfig struct {
	Headers map[string]string                       `json:"headers"`
	Paths   map[HttpPath]map[string]json.RawMessage `json:"paths"`
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

func (rs *Responses) ToJsonSerializable() map[string]json.RawMessage {
	serializable := make(map[string]json.RawMessage)
	for code, res := range *rs {
		if len(res) == 0 {
			res = []byte("{}")
		}
		serializable[strconv.Itoa(code)] = res
	}

	return serializable
}

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
