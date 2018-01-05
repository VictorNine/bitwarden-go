package config

import "github.com/VictorNine/bitwarden-go/internal/database/sqlite"
import (
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

type Conf struct {
	SigningKey          string `yaml:"signingKey"`
	JwtExpire           int    `yaml:"jwtExpire"`
	ServerAddr          string `yaml:"serverAddr"`
	ServerPort          string `yaml:"serverPort"`
	DisableRegistration bool   `yaml:"disableRegistration"`
}

func hasColon(s string) bool {
	return strings.LastIndex(s, ":") > -1
}

func Read() *Conf {
	yamlFile, err := ioutil.ReadFile("../../conf.yaml")
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
