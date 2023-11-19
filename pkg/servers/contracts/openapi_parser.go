package contracts

import (
	server "com.github/salpreh/devserver/pkg/servers"
	"encoding/json"
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"log"
	"os"
	"strconv"
)

func LoadContractMockConfig(contractPath string) *server.MockConfig {
	contract := ReadOpenapiSpec(contractPath)
	paths := processPaths(contract.Model.Paths)

	return &server.MockConfig{
		Headers: make(map[string]string),
		Paths:   paths,
	}
}

func ReadOpenapiSpec(contractPath string) *libopenapi.DocumentModel[v3.Document] {
	contractF, _ := os.ReadFile(contractPath)

	contract, err := libopenapi.NewDocument(contractF)
	if err != nil {
		log.Panicf("Unable to parse contract: %v", err)
	}

	model, errs := contract.BuildV3Model()
	if len(errs) > 0 {
		log.Printf("Unable to load contract. Errors: ")
		for _, e := range errs {
			log.Printf("- %e", e)
		}
		log.Panic("\n")
	}

	return model
}

func processPaths(paths *v3.Paths) map[server.HttpPath]server.Path {
	processedPaths := make(map[server.HttpPath]server.Path)
	for path, pathItem := range paths.PathItems {
		log.Printf("Parsing path %s contract", path)
		methods := make(map[string]server.Responses)
		for method, operation := range pathItem.GetOperations() {
			log.Printf("Parsing %s responses", method)
			responses := processMethodOperation(operation.Responses)
			methods[method] = responses
		}

		processedPaths[server.HttpPath(path)] = server.Path{Methods: methods}
	}

	return processedPaths
}

func processMethodOperation(responses *v3.Responses) server.Responses {
	processedResponses := server.Responses{}
	for codeValue, responseData := range responses.Codes {
		code, _ := strconv.Atoi(codeValue)
		response := getResponseContent(responseData)
		processedResponses[code] = response
	}

	return processedResponses
}

func getResponseContent(response *v3.Response) []byte {
	content, isJson := response.Content["application/json"]
	if !isJson {
		for mediaType, c := range response.Content {
			log.Printf("Using response from content type: %s", mediaType)
			content = c
		}
	}

	responseData := make([]byte, 0)
	if content == nil {
		return responseData
	}

	schema := content.Schema.Schema()
	if schema.Example == nil {
		return responseData
	}

	if isJson {

		var err error
		if responseData, err = json.Marshal(schema.Example); err != nil {
			log.Printf("Unable to parse response data")
		}
	} else if ex, isString := schema.Example.(string); isString {
		log.Printf("Processing raw example")
		return []byte(ex)
	}

	return responseData
}
