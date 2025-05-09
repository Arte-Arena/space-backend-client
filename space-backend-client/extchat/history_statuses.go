// history_statuses.go
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

// StatusRegistro mapeia cada item em value.statuses
type StatusRegistro struct {
	ID        string `bson:"id" json:"id"`
	Status    string `bson:"status" json:"status"`
	Timestamp string `bson:"timestamp" json:"timestamp"`
}

// RawEventStatus reflete entry → changes → value.statuses[]
type RawEventStatus struct {
	Entry []struct {
		Changes []struct {
			Field string           `bson:"field"`
			Value []StatusRegistro `bson:"statuses"`
		} `bson:"changes"`
	} `bson:"entry"`
}

// StatusMessage é o payload que vai no JSON de saída
type StatusMessage struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// SimpleStatusEvent é cada objeto retornado pela API
type SimpleStatusEvent struct {
	RawEvent   StatusMessage `json:"raw_event"`
	ReceivedAt time.Time     `json:"received_at"`
}

// HandlerStatusHistory retorna apenas os eventos de status
func HandlerStatusHistory(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("[StatusHistory] Erro ao conectar no MongoDB: %v", err)
		http.Error(w, "Erro de conexão ao banco", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("[StatusHistory] Erro ao desconectar MongoDB: %v", err)
		}
	}()

	// Buscar todos os eventos ordenados por received_at
	col := client.Database(database.GetDB()).Collection("whatsapp_events")
	cursor, err := col.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "received_at", Value: 1}}))
	if err != nil {
		log.Printf("[StatusHistory] Erro ao buscar histórico: %v", err)
		http.Error(w, "Erro ao buscar histórico", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []SimpleStatusEvent

	// Iterar pelos documentos
	for cursor.Next(ctx) {
		var doc struct {
			RawEvent   RawEventStatus `bson:"raw_event"`
			ReceivedAt time.Time      `bson:"received_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("[StatusHistory] Erro ao decodificar documento: %v", err)
			continue
		}

		// Filtrar apenas os changes com field == "statuses"
		for _, entry := range doc.RawEvent.Entry {
			for _, change := range entry.Changes {
				if change.Field != "statuses" {
					continue
				}
				for _, st := range change.Value {
					// Converter timestamp epoch string para time.Time
					secs, err := strconv.ParseInt(st.Timestamp, 10, 64)
					if err != nil {
						log.Printf("[StatusHistory] Timestamp inválido '%s': %v", st.Timestamp, err)
						continue
					}
					ts := time.Unix(secs, 0).UTC()

					results = append(results, SimpleStatusEvent{
						RawEvent: StatusMessage{
							ID:        st.ID,
							Status:    st.Status,
							Timestamp: ts,
						},
						ReceivedAt: doc.ReceivedAt.UTC(),
					})
				}
			}
		}
	}
	if err := cursor.Err(); err != nil {
		log.Printf("[StatusHistory] Erro durante iteração do cursor: %v", err)
	}

	// Serializar e enviar resposta
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("[StatusHistory] Erro ao serializar resposta: %v", err)
	}
}
