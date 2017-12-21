package main

import "github.com/VictorNine/bitwarden-go/internal/database/sqlite"

var mySigningKey = []byte("secret")
var jwtExpire = 3600

var db = &sqlite.DB{}

const serverAddr = ":8000"
