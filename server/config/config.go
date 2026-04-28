package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Routes struct {
		Root             string `yaml:"root"`
		Health           string `yaml:"health"`
		Ws               string `yaml:"ws"`
		UploadFile       string `yaml:"uploadFile"`
		GetUploadedFiles string `yaml:"getUploadedFiles"`
	} `yaml:"routes"`

	CORS struct {
		AllowedOrigins []string `yaml:"allowed_origins"`
	} `yaml:"cors"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

//Usage
/*
cfg, _ := config.Load("config.yaml")

fmt.Println(cfg.Server.Port)
fmt.Println(cfg.Routes.Login)
fmt.Println(cfg.CORS.AllowedOrigins)
*/
