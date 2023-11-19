package collectionutils

import (
	"encoding/json"
	"fmt"
	"os"
)

func ExportToFileAsJson(filePath string, data any) error {
	encodedData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return fmt.Errorf("error serializing data to json: %w", err)
	}

	if err = os.WriteFile(filePath, encodedData, 0644); err != nil {
		return fmt.Errorf("error writing to file %s: %w", filePath, err)
	}

	return nil
}
