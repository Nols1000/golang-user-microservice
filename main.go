package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"v1/config"
	"v1/core"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	config.ReadConfig()

	// Parsing wait for graceful shutdown setting
	wait, err := time.ParseDuration(config.App.Server.WaitDurationForGracefulShutdown)

	if err != nil {
		log.Printf("Failed to parse wait duration of '%s' shutdown without waiting for connections to close.", config.App.Server.WaitDurationForGracefulShutdown)
		wait = 0
	}

	cor := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Access-Control-Allow-Origin", "Content-Type", "Session-Key", "Device-ID"},
		Debug:          true,
	})

	router := mux.NewRouter()
	router.HandleFunc("/api/ping", core.OnPing).Methods("GET")
	router.HandleFunc("/api/users", core.OnSignup).Methods("POST")
	router.HandleFunc("/api/users", core.OnSignup).Methods("POST")
	router.HandleFunc("/api/users", core.OnGetUsers).Methods("GET")
	router.HandleFunc("/api/users/{uuid}", core.OnDeleteUser).Methods("DELETE")
	router.HandleFunc("/api/users/{uuid}", core.OnGetUser).Methods("GET")
	router.HandleFunc("/api/users/{uuid}", core.OnUpdateUser).Methods("PUT")

	handler := cor.Handler(router)
	address := fmt.Sprintf(":%s", config.App.Server.Port)

	server := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	log.Printf("Started server listening at %s", address)

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c
	log.Printf("Waiting %s for connections to close.", wait.String())

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	server.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("Server gracefully shutdown.")
	os.Exit(0)
}
