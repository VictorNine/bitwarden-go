package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/VictorNine/bitwarden-go/internal/api"
	"github.com/VictorNine/bitwarden-go/internal/auth"
	"github.com/VictorNine/bitwarden-go/internal/config"
	"github.com/rs/cors"
)

func main() {
	initDB := flag.Bool("init", false, "Initialize the database")
	configFile := flag.String("conf","../../conf.yaml","Location of the config file. Default: ../../conf.yaml")
	flag.Parse()

	cfg := config.Read(*configFile)

	err := config.DB.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer config.DB.Close()

	// Create a new database
	if *initDB {
		err := config.DB.Init()
		if err != nil {
			log.Fatal(err)
		}
	}

	authHandler := auth.New(config.DB, cfg.SigningKey, cfg.JwtExpire)
	apiHandler := api.New(config.DB)

	mux := http.NewServeMux()

	if cfg.DisableRegistration == false {
		mux.HandleFunc("/api/accounts/register", authHandler.HandleRegister)
	}
	mux.HandleFunc("/identity/connect/token", authHandler.HandleLogin)

	mux.Handle("/api/accounts/keys", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleKeysUpdate)))
	mux.Handle("/api/accounts/profile", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleProfile)))
	mux.Handle("/api/collections", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleCollections)))
	mux.Handle("/api/folders", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleFolder)))
	mux.Handle("/api/folders/", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleFolderUpdate)))
	mux.Handle("/apifolders", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleFolder))) // The android app want's the address like this, will be fixed in the next version. Issue #174
	mux.Handle("/api/sync", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleSync)))

	mux.Handle("/api/ciphers/import", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleImport)))
	mux.Handle("/api/ciphers", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleCipher)))
	mux.Handle("/api/ciphers/", authHandler.JwtMiddleware(http.HandlerFunc(apiHandler.HandleCipherUpdate)))

	log.Println("Starting server on " + cfg.ServerAddr + cfg.ServerPort)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
	}).Handler(mux)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr+cfg.ServerPort, handler))
}
