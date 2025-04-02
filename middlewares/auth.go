package middlewares

import (
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
)

type ContextKey string

const (
	UserIDKey ContextKey = "userId"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: "Token de acesso não encontrado",
			})
			return
		}

		claims, err := utils.ValidateAccessKey(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: "Token inválido ou expirado",
			})
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserId)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
}
