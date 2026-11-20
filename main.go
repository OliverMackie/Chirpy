package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	database "github.com/OliverMackie/chirpy/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	SecretKey := os.Getenv("SECRET_KEY")
	API_key := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error opening database: %s", err)
	}
	dbQueries := database.New(db)

	config := &APIConfig{
		fileserverHits: atomic.Int32{},
		database:       dbQueries,
		plat:           platform,
		SecretKey:      SecretKey,
		APIkey:         API_key,
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", config.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("GET /admin/metrics", config.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", config.handlerResetMetrics)
	mux.HandleFunc("GET /api/chirps", config.handlerGetChirps)
	mux.HandleFunc("POST /api/users", config.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", config.handlerValidate)
	mux.HandleFunc("GET /api/chirps/{chirpID}", config.handlerGetChirp)
	mux.HandleFunc("POST /api/login", config.handlerLogin)
	mux.HandleFunc("POST /api/refresh", config.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", config.handlerRevoke)
	mux.HandleFunc("PUT /api/users", config.handlerUpdateUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", config.handlerDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", config.handlerUpgradeUser)
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
