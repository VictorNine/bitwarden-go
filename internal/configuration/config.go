package config

import "github.com/VictorNine/bitwarden-go/internal/database/sqlite"

type config struct {
	SigningKey []byte
	JwtExpire  int
	ServerAddr string
	ServerPort string
}

var CFG = config{[]byte("secret"), 3600, "", ":8000"}

var DB = &sqlite.DB{}
