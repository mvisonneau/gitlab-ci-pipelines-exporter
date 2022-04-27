package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Format represents the format of the config file.
type Format uint8

const (
	// FormatYAML represents a Config written in yaml format.
	FormatYAML Format = iota
)

// ParseFile reads the content of a file and attempt to unmarshal it
// into a Config.
func ParseFile(filename string) (c Config, err error) {
	var (
		t         Format
		fileBytes []byte
	)

	// Figure out what type of config file we provided
	t, err = GetTypeFromFileExtension(filename)
	if err != nil {
		return
	}

	// Read the content of the config file
	fileBytes, err = ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return
	}

	// Parse the content and return Config
	return Parse(t, fileBytes)
}

// Parse unmarshal provided bytes with given ConfigType into a Config object.
func Parse(f Format, bytes []byte) (cfg Config, err error) {
	switch f {
	case FormatYAML:
		err = yaml.Unmarshal(bytes, &cfg)
	default:
		err = fmt.Errorf("unsupported config type '%+v'", f)
	}

	// hack: automatically update the cfg.GitLab.HealthURL for self-hosted GitLab
	if cfg.Gitlab.URL != "https://gitlab.com" &&
		cfg.Gitlab.HealthURL == "https://gitlab.com/explore" {
		cfg.Gitlab.HealthURL = fmt.Sprintf("%s/-/health", cfg.Gitlab.URL)
	}

	return
}

// GetTypeFromFileExtension returns the ConfigType based upon the extension of
// the file.
func GetTypeFromFileExtension(filename string) (f Format, err error) {
	switch ext := filepath.Ext(filename); ext {
	case ".yml", ".yaml":
		f = FormatYAML
	default:
		err = fmt.Errorf("unsupported config type '%s', expected .y(a)ml", ext)
	}

	return
}
