package config

import "github.com/VictorNine/bitwarden-go/internal/database/sqlite"
import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

type Conf struct {
	SigningKey string `yaml:"signingKey"`
	JwtExpire  int    `yaml:"jwtExpire"`
	ServerAddr string `yaml:"serverAddr"`
	ServerPort string `yaml:"serverPort"`
}

func hasColon(s string) bool {
	return strings.LastIndex(s, ":") > -1
}

func Read(configFile string) *Conf {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println(err)
	}

	var config Conf
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Println(err)
	}

	if !hasColon(config.ServerPort) {
		config.ServerPort = ":" + config.ServerPort
	}

	return &config
}

var DB = &sqlite.DB{}
