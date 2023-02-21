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

	Porkbun struct {
		BaseUrl string `yaml:"baseurl"`
		TTL     int    `yaml:"ttl"`
	} `yaml:"porkbun"`

	Api struct {
		Port             string `yaml:"port"`
		ApiKeyHidden     bool   `yaml:"hideApiKeyInLogs"`
		StatusLogEnabled bool   `yaml:"enableStatusLog"`
	} `yaml:"api"`

	Test struct {
		Gandi struct {
			ApiKey    string `yaml:"apiKey"`
			Domain    string `yaml:"domain"`
			Subdomain string `yaml:"subdomain"`
			IP        string `yaml:"ip"`
		} `yaml:"gandi"`
		Porkbun struct {
			ApiKey       string `yaml:"apiKey"`
			ApiSecretKey string `yaml:"apiSecretKey"`
			Domain       string `yaml:"domain"`
			Subdomain    string `yaml:"subdomain"`
			IP           string `yaml:"ip"`
		} `yaml:"porkbun"`
	} `yaml:"test"`
}

var AppConfig *Config

// InitConfig parses a config.yml file placed in the root execution path containing credentials and settings for the application.
func InitConfig() error {

	if _, err := os.Stat("./config.yml"); errors.Is(err, os.ErrNotExist) {
		return errors.New("config.yml not found")
	}

	AppConfig = &Config{}

	// open config file
	p := filepath.FromSlash("./config.yml")
	file, err := os.Open(p)
	if err != nil {
		return err
	}
	defer file.Close()

	// init new YAML decode
	d := yaml.NewDecoder(file)

	// start YAML decoding from file
	if err := d.Decode(&AppConfig); err != nil {
		return err
	}

	if (&Config{}) == AppConfig {
		return errors.New("config is empty")
	}

	return err
}
