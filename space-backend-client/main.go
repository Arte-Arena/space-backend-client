package main

import (
	"api/admin"
	"api/auth"
	"api/clients"
	"api/extchat"
	"api/middlewares"
	"api/schemas"
	"api/uniforms"
	"api/utils"
	"api/ws"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func setupRouter(hub *ws.Hub) http.Handler {
	// 1) Roteador exclusivo para WebSocket, sem middlewares que embrulham ResponseWriter
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/v1/ws/whatsapp", hub.Handler())

	// 2) Roteador para as demais rotas REST, com middlewares aplicados
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/v1/auth/signin", auth.Signin)
	apiMux.HandleFunc("/v1/auth/signup", auth.Signup)
	apiMux.HandleFunc("/v1/auth/authorize", auth.Authorize)
	apiMux.HandleFunc("/v1/auth/signout", auth.Signout)

	apiMux.HandleFunc("/v1/admin/uniforms", middlewares.AdminMiddleware(admin.HandlerUniforms))
	apiMux.HandleFunc("/v1/admin/clients", middlewares.AdminMiddleware(admin.HandlerClients))

	apiMux.HandleFunc("/v1/clients", middlewares.AuthMiddleware(clients.Handler))
	apiMux.HandleFunc("/v1/uniforms", middlewares.AuthMiddleware(uniforms.Handler))

	// apiMux.HandleFunc("/v1/webhook/whatsapp", middlewares.ExtChatMiddleware(extchat.HandlerWhatsapp))
	apiMux.HandleFunc("/v1/webhook/whatsapp", extchat.HandlerWhatsapp)
	apiMux.HandleFunc("/v1/history/whatsapp", extchat.HandlerHistory)

	// 3) Dispatcher que escolhe qual mux usar com base no path
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/ws/") {
			// WebSocket: bypass de middlewares
			wsMux.ServeHTTP(w, r)
		} else {
			// REST: passa pelos wrappers de logging, CORS e security headers
			handler := middlewares.Logging(
				middlewares.SecurityHeaders(
					middlewares.Cors(apiMux),
				),
			)
			handler.ServeHTTP(w, r)
		}
	})
}

func main() {
	utils.LoadEnvVariables()

	// Inicializa e dispara o Hub de WebSocket
	hub := ws.NewHub()
	go hub.Run()
	extchat.Hub = hub

	// Carrega configuração
	config := schemas.Config{
		Port:         os.Getenv(utils.ENV_PORT),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Cria e inicia o servidor HTTP
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      setupRouter(hub),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	// Inicia em goroutine para que possamos escutar sinais de shutdown
	go func() {
		log.Printf("Server started on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Aguarda sinal de interrupção para shutdown gracioso
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
