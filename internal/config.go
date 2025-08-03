package internal

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	ListenPort int `yaml:"port"`
}

type AccessPointConfig struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	KeyFile  string `yaml:"keyfile"`
}

type Config struct {
	Global       GlobalConfig        `yaml:"global"`
	AccessPoints []AccessPointConfig `yaml:"accesspoints"`
}

func NewConfig(path string) (*Config, error) {
	config := &Config{
		Global: GlobalConfig{
			ListenPort: 9130,
		},
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	data := yaml.NewDecoder(file)
	data.KnownFields(true)
	if err := data.Decode(&config); err != nil {
		return nil, err
	}

	if config == nil {
		return nil, errors.New("config is empty")
	}

	if len(config.AccessPoints) == 0 {
		return nil, errors.New("no access points defined")
	}

	/* Check configuration */
	for i, accessPoint := range config.AccessPoints {
		if accessPoint.Name == "" {
			return nil, fmt.Errorf("accesspoint #%d is missing `name`", i+1)
		}
		if accessPoint.Address == "" {
			return nil, fmt.Errorf("accesspoint #%d is missing `address`", i+1)
		}
		if accessPoint.Username == "" {
			return nil, fmt.Errorf("accesspoint #%d is missing `username`", i+1)
		}
		if accessPoint.Password == "" && accessPoint.KeyFile == "" {
			return nil, fmt.Errorf("accesspoint #%d requires either `password` or `keyfile`", i+1)
		}
	}

	return config, nil
}
