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

func expand(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting homedir")
	}

	if len(path) == 0 {
		return path, nil
	}

	if path[0] != '~' && path[1] != '/' {
		return path, nil
	}

	return fmt.Sprintf("%s%s", homeDir, path[1:]), nil
}
