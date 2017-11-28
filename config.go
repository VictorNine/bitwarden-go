package main

var mySigningKey = []byte("secret")
var jwtExpire = 36000

var db database = &DB{}

const serverAddr = ":8000"
