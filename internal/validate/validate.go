// Package validate contains validation functions
package validate

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

func IsValidURL(str string) error {
	if str == "" {
		return errors.New("url is required")
	}
	url, err := url.Parse(str)
	if err != nil || url.Scheme == "" || url.Host == "" {
		return fmt.Errorf("invalid url: %s", str)
	}
	return nil
}

func IsValidInputPath(path string) error {
	if path == "" {
		return fmt.Errorf("input path is required")
	}

	// os.Stat returns a FileInfo struct (we ignore it) and an error,
	// os.IsNotExist() takes in the error and returns true if the directory exists
	// and false if it doesnt
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	return nil
}

func IsValidOutputPath(path string) error {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	return nil
}
