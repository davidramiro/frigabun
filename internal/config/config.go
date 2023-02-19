package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config contains all the settings parsed from the config file.
type Config struct {
	Gandi struct {
		BaseUrl string `yaml:"baseurl"`
		TTL     int    `yaml:"ttl"`
	} `yaml:"gandi"`

	Api struct {
		Port         string `yaml:"port"`
		ApiKeyHidden bool   `yaml:"hideApiKeyInLogs"`
	} `yaml:"api"`
}

var config *Config

// GetConfig parses a config.yml file placed in the root execution path containing credentials and settings for the application.
// It returns an object containing the parsed settings.
func GetConfig() (*Config, error) {

	config = &Config{}

	// open config file
	p := filepath.FromSlash("./config.yml")
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// init new YAML decode
	d := yaml.NewDecoder(file)

	// start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	if (&Config{}) == config {
		return nil, errors.New("config is empty")
	}

	return config, err
}

func CheckPresent() (bool, error) {
	if _, err := os.Stat("./config.yml"); errors.Is(err, os.ErrNotExist) {
		return false, errors.New("config.yml not found")
	}
	return true, nil
}
