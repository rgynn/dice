package main

import (
	"log"
	"net/http"

	"github.com/rgynn/dice/pkg/api"
	"github.com/rgynn/dice/pkg/config"

	"github.com/gorilla/mux"
)

func main() {

	cfg, err := config.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	svc, err := api.NewService(cfg)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/sessions", svc.NewSessionHandler).Methods(http.MethodPost)
	router.HandleFunc("/sessions/{sessionID}/{playerID}", svc.NewRollHandler).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	log.Printf("Listening on: %s\n", cfg.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
