package middlewares

import (
	"api/schemas"
	"api/utils"
	"encoding/json"
	"net/http"
	"os"
)

func ExtChatMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		extChatKey := r.Header.Get("x-api-key")
		if extChatKey == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Chave de segurança não fornecida",
			})
			return
		}

		envExtChatKey := os.Getenv(utils.X_API_KEY_EXTCHAT)
		if extChatKey != envExtChatKey {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Chave de segurança inválida",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}
