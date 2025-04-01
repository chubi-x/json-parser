package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type Json struct {
	file *os.File
	path string
}

var validFiles = []Json{}
var invalidFiles = []Json{}
var _ = filepath.Walk("tests", func(path string, info os.FileInfo, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}

	if ext := filepath.Ext(path); ext != "" && ext == ".json" {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		if strings.Contains(path, "fail") || strings.Contains(path, "invalid") {
			invalidFiles = append(invalidFiles, Json{f, path})
		} else if strings.Contains(path, "pass") || strings.Index(filepath.Base(path), "valid") == 0 {
			validFiles = append(validFiles, Json{f, path})
		}
	}
	return nil
})

// tests/rfctests/pass1.json will fail because of the naive handling of escaped characters
func TestInvalid(t *testing.T) {

	for _, json := range invalidFiles {

		buf := bytes.NewBuffer(make([]byte, 0))
		io.Copy(buf, json.file)
		tokens := slices.Concat(Lex(buf)...)
		_, parseErr := Parse(&tokens)
		if parseErr == nil {

			t.Errorf("Expected invalid but got valid: %s", json.path)
		}
	}
}
func TestValid(t *testing.T) {

	for _, json := range validFiles {

		buf := bytes.NewBuffer(make([]byte, 0))
		io.Copy(buf, json.file)
		tokens := slices.Concat(Lex(buf)...)
		_, parseErr := Parse(&tokens)
		if parseErr != nil {

			t.Errorf("Expected valid but got invalid for %s: %s", json.path, parseErr)
		}
	}
}
