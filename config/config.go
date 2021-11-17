package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Project      Project      `yaml:"project"`
	Repositories []Repository `yaml:"repositories"`
}

type Project struct {
	Organization string `yaml:"organization"`
	Number       int    `yaml:"number"`
}

type Repository struct {
	Name   string  `yaml:"name"`
	Fields []Field `yaml:"fields"`
}

type Field struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
	Type  string `yaml:"type"`
}

func LoadConfig(file string) (*Config, error) {
	c, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(c, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
