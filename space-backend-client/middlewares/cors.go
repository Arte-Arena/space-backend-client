package middlewares

import (
	"api/utils"
	"net/http"
	"os"
	"slices"
)

func Cors(next http.Handler) http.Handler {
	allowedOrigins := []string{
		"http://localhost:8000",
		"http://localhost:3000",
	}

	if os.Getenv(utils.ENV) == utils.ENV_RELEASE {
		allowedOrigins = []string{
			"https://api.spacearena.net",
			"https://my.spacearena.net",
			"https://spacearena.net",
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
