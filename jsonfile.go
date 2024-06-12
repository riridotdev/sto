package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (jf jsonFile) read(data interface{}) error {
	f, err := os.OpenFile(string(jf), os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening file %q: %v", string(jf), err)
	}

	dec := json.NewDecoder(f)
	if err := dec.Decode(data); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading file %q: %v", string(jf), err)
	}

	return nil
}

func (jf jsonFile) write(data interface{}) error {
	f, err := os.OpenFile(string(jf), os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("opening file %q: %v", string(jf), err)
	}

	buf := bytes.NewBuffer([]byte{})

	enc := json.NewEncoder(buf)
	enc.SetIndent("", "")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encoding json: %v", err)
	}

	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("writing to %q: %v", string(jf), err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing file %q: %v", string(jf), err)
	}

	return nil
}
