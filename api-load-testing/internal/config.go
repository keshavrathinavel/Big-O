package internal

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerIps [7]string `yaml:"server_ips"`
}

func ReadConfig() [7]string {
	fileData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error while reading config file: %v", err)
	}

	data := Config{}
	err = yaml.Unmarshal(fileData, &data)
	if err != nil {
		log.Fatalf("Error while parsing YAML: %v", err)
	}
	return data.ServerIps
}
