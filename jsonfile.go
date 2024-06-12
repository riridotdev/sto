package main

import (
	"fmt"
	"os"
)

type jsonFile string

func newJsonFile(path string) (jsonFile, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return jsonFile(path), fmt.Errorf("opening file %q: %v", path, err)
	}
	if err := f.Close(); err != nil {
		return jsonFile(path), fmt.Errorf("closing file %q: %v", path, err)
	}

	return jsonFile(path), nil
}
