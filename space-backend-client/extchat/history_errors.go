// history_errors.go
package extchat

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ErrorRegistro mapeia cada item em value.errors
type ErrorRegistro struct {
	Code      string `bson:"code" json:"code"`
	Message   string `bson:"message" json:"message"`
	Timestamp string `bson:"timestamp" json:"timestamp"`
}

// RawEventError reflete entry → changes → value.errors[]
type RawEventError struct {
	Entry []struct {
		Changes []struct {
			Field string          `bson:"field"`
			Value []ErrorRegistro `bson:"errors"`
		} `bson:"changes"`
	} `bson:"entry"`
}

// ErrorMessage é o payload para o JSON de saída
type ErrorMessage struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// SimpleErrorEvent é cada objeto retornado pela API
type SimpleErrorEvent struct {
	RawEvent   ErrorMessage `json:"raw_event"`
	ReceivedAt time.Time    `json:"received_at"`
}

// HandlerErrorHistory retorna apenas os eventos de erro
func HandlerErrorHistory(w http.ResponseWriter, r *http.Request) {
	// Só GET permitido
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Contexto com timeout
	ctx, cancel := context.WithTimeout(r.Context(), database.MONGODB_TIMEOUT)
	defer cancel()

	// Conectar ao MongoDB
	mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Printf("[ErrorHistory] Erro ao conectar no MongoDB: %v", err)
		http.Error(w, "Erro de conexão ao banco", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("[ErrorHistory] Erro ao desconectar MongoDB: %v", err)
		}
	}()

	// Buscar todos os eventos ordenados por received_at
	col := client.Database(database.GetDB()).Collection("whatsapp_events")
	cursor, err := col.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "received_at", Value: 1}}))
	if err != nil {
		log.Printf("[ErrorHistory] Erro ao buscar histórico: %v", err)
		http.Error(w, "Erro ao buscar histórico", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []SimpleErrorEvent

	// Iterar pelos documentos
	for cursor.Next(ctx) {
		var doc struct {
			RawEvent   RawEventError `bson:"raw_event"`
			ReceivedAt time.Time     `bson:"received_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("[ErrorHistory] Erro ao decodificar documento: %v", err)
			continue
		}

		// Filtrar apenas os changes com field == "errors"
		for _, entry := range doc.RawEvent.Entry {
			for _, change := range entry.Changes {
				if change.Field != "errors" {
					continue
				}
				for _, er := range change.Value {
					// Converter timestamp epoch string para time.Time
					secs, err := strconv.ParseInt(er.Timestamp, 10, 64)
					if err != nil {
						log.Printf("[ErrorHistory] Timestamp inválido '%s': %v", er.Timestamp, err)
						continue
					}
					ts := time.Unix(secs, 0).UTC()

					results = append(results, SimpleErrorEvent{
						RawEvent: ErrorMessage{
							Code:      er.Code,
							Message:   er.Message,
							Timestamp: ts,
						},
						ReceivedAt: doc.ReceivedAt.UTC(),
					})
				}
			}
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("[ErrorHistory] Erro durante iteração do cursor: %v", err)
	}

	// Serializar e enviar resposta
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("[ErrorHistory] Erro ao serializar resposta: %v", err)
	}
}
