package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// config.yml to struct (ConfigData consists of all itemson the top level of the yml)
type Directories struct {
	Databasedir string `yaml:"databasedir"`
	Musicdir    string `yaml:"musicdir"`
	Newmusicdir string `yaml:"newmusicdir"`
}

type Loglevel struct {
	Databaselevel string `yaml:"databaselevel"`
	Filelevel     string `yaml:"filelevel"`
}

// root struct (top level YAML stuff)
type ConfigData struct {
	Directories Directories `yaml:"directories"`
	Loglevel    Loglevel    `yaml:"loglevel"`
}

// pretty simple config loader
func LoadConfig(configPath string) (ConfigData, error) {
	var configData ConfigData

	configRaw, err := os.ReadFile(configPath)
	if err != nil {
		return configData, err
	}

	// Unmarshal yaml to configData
	if err := yaml.Unmarshal(configRaw, &configData); err != nil {
		return configData, err
	}
	return configData, nil
}
