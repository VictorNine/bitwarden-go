package main

var mySigningKey = []byte("secret")
var jwtExpire = 36000

var db DB

const serverAddr = ":8000"
