package server

import "encoding/json"

type HttpPath string

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
