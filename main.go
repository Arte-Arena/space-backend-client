package main

import (
	"api/auth"
	"api/clients"
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

	mux.HandleFunc("/v1/auth/signin", auth.Signin)
	mux.HandleFunc("/v1/auth/authorize", auth.Authorize)
	mux.HandleFunc("/v1/auth/signout", auth.Signout)

	mux.HandleFunc("/v1/clients", middlewares.AuthMiddleware(clients.Handler))
	mux.HandleFunc("/v1/uniforms", middlewares.AuthMiddleware(health.Handler))

	handler := middlewares.Logging(middlewares.SecurityHeaders(middlewares.Cors(mux)))
	return handler
}

func main() {
	utils.LoadEnvVariables()

	config := utils.Config{
		Port:         os.Getenv(utils.ENV_PORT),
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
