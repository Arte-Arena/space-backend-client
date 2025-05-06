package middlewares

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Lê e loga payload da requisição sem consumir r.Body permanentemente
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("{04} - [Logging] erro ao ler body: %v", err)
		} else {
			log.Printf("{05} - [Logging] payload recebido: %s", string(bodyBytes))
		}
		// Reposiciona r.Body para próxima leitura pelo handler
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(w, r)

		// Log final com método, URI, remoto e duração
		log.Printf(
			"%s %s %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}
