package main

import (
	"api/health"
	"api/middlewares"
	"api/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func setupRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", health.Handler)
	handler := middlewares.Logging(middlewares.SecurityHeaders(middlewares.Cors(mux)))
	return handler
}

func main() {
	config := utils.Config{
		Port:         "8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      setupRouter(),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	go func() {
		log.Printf("Server started on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("Shutdown initiated...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

	log.Println("Server successfully terminated")
}
