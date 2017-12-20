package main

import (
	bw "github.com/VictorNine/bitwarden-go/internal/common"
)

var mySigningKey = []byte("secret")
var jwtExpire = 3600

var db bw.Database = &bw.DB{}

const serverAddr = ":8000"
