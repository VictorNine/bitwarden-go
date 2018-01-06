package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/VictorNine/bitwarden-go/internal/api"
	"github.com/VictorNine/bitwarden-go/internal/auth"
	"github.com/VictorNine/bitwarden-go/internal/database/sqlite"
	"github.com/rs/cors"
)

var cfg struct {
	initDB              bool
	signingKey          string
	jwtExpire           int
	hostAddr            string
	hostPort            string
	disableRegistration bool
}

func init() {
	flag.BoolVar(&cfg.initDB, "init", false, "Initalizes the database.")
	flag.StringVar(&cfg.signingKey, "key", "secret", "Sets the signing key")
	flag.IntVar(&cfg.jwtExpire, "tokenTime", 3600, "Sets the ammount of time (in seconds) the generated JSON Web Tokens will last before expiry.")
	flag.StringVar(&cfg.hostAddr, "host", "", "Sets the interface that the application will listen on.")
	flag.StringVar(&cfg.hostPort, "port", "8000", "Sets the port")
	flag.BoolVar(&cfg.disableRegistration, "disableRegistration", false, "Disables user registration.")
}

func main() {
	db := &sqlite.DB{}
	flag.Parse()

	err := db.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Create a new database
	if cfg.initDB {
		err := db.Init()
		if err != nil {
			log.Fatal(err)
		}
	}

	authHandler := auth.New(db, cfg.signingKey, cfg.jwtExpire)
	apiHandler := api.New(db)

	mux := http.NewServeMux()

	if cfg.disableRegistration == false {
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

	log.Println("Starting server on " + cfg.hostAddr + ":" + cfg.hostPort)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
	}).Handler(mux)
	log.Fatal(http.ListenAndServe(cfg.hostAddr+":"+cfg.hostPort, handler))
}
