package middlewares

import (
	"api/schemas"
	"api/utils"
	"encoding/json"
	"net/http"
	"os"
)

func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminKey := r.Header.Get("X-Admin-Key")
		if adminKey == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Chave de administrador não fornecida",
			})
			return
		}

		envAdminKey := os.Getenv(utils.ADMIN_KEY)
		if adminKey != envAdminKey {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Chave de administrador inválida",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}
