// Package validate checks user input before jobs are created.
package validate

import (
	"errors"
	"fmt"
	"net/url"
	"os"
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

func IsValidFormat(format string) error {
	if format == "" {
		return errors.New("format is required")
	}
	for _, char := range format {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
			return fmt.Errorf("invalid format %q: use letters and numbers only", format)
		}
	}
	return nil
}

func IsValidInputPath(path string) error {
	if path == "" {
		return errors.New("input path is required")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input path does not exist: %s", path)
		}
		return fmt.Errorf("failed to inspect input path %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("input path is a directory: %s", path)
	}

	return nil
}

func IsValidOutputPath(path string) error {
	if path == "" {
		path = "."
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", path)
		}
		return fmt.Errorf("failed to inspect output directory %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("output path is not a directory: %s", path)
	}

	return nil
}
