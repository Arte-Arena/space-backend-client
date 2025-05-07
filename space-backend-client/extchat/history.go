package extchat

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// HandlerHistory retorna o histórico de chats e mensagens
func HandlerHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), database.MONGODB_TIMEOUT)
	defer cancel()

	// Configura e conecta ao cliente MongoDB
	mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Printf("[History] Erro ao conectar no MongoDB: %v", err)
		http.Error(w, "Erro de conexão ao banco", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("[History] Erro ao desconectar MongoDB: %v", err)
		}
	}()

	col := client.Database(database.GetDB()).Collection("whatsapp_events")
	cursor, err := col.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "received_at", Value: 1}}))
	if err != nil {
		log.Printf("[History] Erro ao buscar histórico: %v", err)
		http.Error(w, "Erro ao buscar histórico", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	type Event struct {
		RawEvent   map[string]interface{} `json:"raw_event"`
		ReceivedAt time.Time              `json:"received_at"`
	}

	var history []Event
	for cursor.Next(ctx) {
		var doc struct {
			RawEvent   map[string]interface{} `bson:"raw_event"`
			ReceivedAt time.Time              `bson:"received_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("[History] Erro ao decodificar documento: %v", err)
			continue
		}
		history = append(history, Event{RawEvent: doc.RawEvent, ReceivedAt: doc.ReceivedAt})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(history); err != nil {
		log.Printf("[History] Erro ao serializar resposta: %v", err)
	}
}
