package cmdutil

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func WriteDetailsToFile(data []byte, outPath string) error {
	fmt.Println(">>> data", string(data))
	fmt.Println(">>> outPath", outPath)
	err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(outPath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func UnmarshallJsonFromReader(file io.Reader, object interface{}) error {
	jsonFile, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonFile, &object)
	if err != nil {
		return err
	}
	return nil
}
