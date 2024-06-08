package main

import (
	"fmt"
	"os"
	"strings"
)

func compress(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting homedir")
	}

	if !strings.HasPrefix(path, homeDir) {
		return path, nil
	}

	return strings.Replace(path, homeDir, "~", 1), nil
}
