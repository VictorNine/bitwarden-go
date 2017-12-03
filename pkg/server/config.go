package server

var mySigningKey = []byte("secret")
var jwtExpire = 3600

var db database = &DB{}

const serverAddr = ":8000"
