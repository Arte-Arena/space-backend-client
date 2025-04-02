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
		accessCookie, err := r.Cookie("access_token")
		if err == nil {
			claims, err := utils.ValidateAccessKey(accessCookie.Value)
			if err == nil {
				ctx := context.WithValue(r.Context(), UserIDKey, claims.UserId)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
		}

		refreshCookie, err := r.Cookie("refresh_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: "Tokens de autenticação não encontrados",
			})
			return
		}

		refreshClaims, err := utils.ValidateRefreshKey(refreshCookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: "Refresh token inválido ou expirado",
			})
			return
		}

		newAccessToken, err := utils.GenerateAccessKey(refreshClaims.UserId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: "Erro ao gerar novo token de acesso",
			})
			return
		}

		newRefreshToken, err := utils.GenerateRefreshKey(refreshClaims.UserId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: "Erro ao gerar novo refresh token",
			})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    newAccessToken,
			Path:     "/",
			MaxAge:   15 * 60,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			Path:     "/",
			MaxAge:   7 * 24 * 60 * 60,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		ctx := context.WithValue(r.Context(), UserIDKey, refreshClaims.UserId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
}
