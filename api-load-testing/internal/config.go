package internal

import (
	"log"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerIps [7]string `yaml:"server_ips"`
}

func ReadConfig(fileContents []byte) [7]string {
	data := Config{}

	err := yaml.Unmarshal(fileContents, &data)
	if err != nil {
		log.Fatalf("Error while parsing YAML: %v", err)
	}
	return data.ServerIps
}
