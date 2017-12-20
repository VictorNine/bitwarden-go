package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	initDB := flag.Bool("init", false, "Initialize the database")
	flag.Parse()

	err := db.open()
	if err != nil {
		log.Fatal(err)
	}

	defer db.close()

	// Create a new database
	if *initDB {
		err := db.init()
		if err != nil {
			log.Fatal(err)
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts/register", handleRegister)
	mux.HandleFunc("/identity/connect/token", handleLogin)

	mux.Handle("/api/accounts/keys", jwtMiddleware(http.HandlerFunc(handleKeysUpdate)))
	mux.Handle("/api/accounts/profile", jwtMiddleware(http.HandlerFunc(handleProfile)))
	mux.Handle("/api/collections", jwtMiddleware(http.HandlerFunc(handleCollections)))
	mux.Handle("/api/folders", jwtMiddleware(http.HandlerFunc(handleFolder)))
	mux.Handle("/apifolders", jwtMiddleware(http.HandlerFunc(handleFolder))) // The android app want's the address like this, will be fixed in the next version. Issue #174
	mux.Handle("/api/sync", jwtMiddleware(http.HandlerFunc(handleSync)))

	mux.Handle("/api/ciphers/import", jwtMiddleware(http.HandlerFunc(handleImport)))
	mux.Handle("/api/ciphers", jwtMiddleware(http.HandlerFunc(handleCipher)))
	mux.Handle("/api/ciphers/", jwtMiddleware(http.HandlerFunc(handleCipherUpdate)))

	log.Println("Starting server on " + serverAddr)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
	}).Handler(mux)
	http.ListenAndServe(serverAddr, handler)
}
