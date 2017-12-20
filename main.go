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

	authHandler := newAuth(db)
	apiHandler := newAPI(db)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts/register", authHandler.handleRegister)
	mux.HandleFunc("/identity/connect/token", authHandler.handleLogin)

	mux.Handle("/api/accounts/keys", jwtMiddleware(http.HandlerFunc(apiHandler.HandleKeysUpdate)))
	mux.Handle("/api/accounts/profile", jwtMiddleware(http.HandlerFunc(apiHandler.HandleProfile)))
	mux.Handle("/api/collections", jwtMiddleware(http.HandlerFunc(apiHandler.HandleCollections)))
	mux.Handle("/api/folders", jwtMiddleware(http.HandlerFunc(apiHandler.HandleFolder)))
	mux.Handle("/apifolders", jwtMiddleware(http.HandlerFunc(apiHandler.HandleFolder))) // The android app want's the address like this, will be fixed in the next version. Issue #174
	mux.Handle("/api/sync", jwtMiddleware(http.HandlerFunc(apiHandler.HandleSync)))

	mux.Handle("/api/ciphers/import", jwtMiddleware(http.HandlerFunc(apiHandler.HandleImport)))
	mux.Handle("/api/ciphers", jwtMiddleware(http.HandlerFunc(apiHandler.HandleCipher)))
	mux.Handle("/api/ciphers/", jwtMiddleware(http.HandlerFunc(apiHandler.HandleCipherUpdate)))

	log.Println("Starting server on " + serverAddr)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
	}).Handler(mux)
	http.ListenAndServe(serverAddr, handler)
}
