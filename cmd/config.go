package main

import (
	bw "github.com/VictorNine/bitwarden-go/internal/common"
	"github.com/VictorNine/bitwarden-go/internal/database/sqlite"
)

var mySigningKey = []byte("secret")
var jwtExpire = 3600

var db bw.Database = &sqlite.DB{}

const serverAddr = ":8000"
