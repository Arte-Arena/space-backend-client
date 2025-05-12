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

// MessageText já existia
type MessageText struct {
	Body string `bson:"body" json:"body"`
}

// Message armazena o raw_message do Mongo, com Timestamp ainda como string
type Message struct {
	From      string      `bson:"from" json:"from"`
	Timestamp string      `bson:"timestamp" json:"timestamp"`
	Text      MessageText `bson:"text" json:"text"`
}

// Value e Change já existiam
type Value struct {
	Messages []Message `bson:"messages"`
}

type Change struct {
	Field string `bson:"field"`
	Value Value  `bson:"value"`
}

type Entry struct {
	Changes []Change `bson:"changes"`
}

type RawEvent struct {
	Entry []Entry `bson:"entry"`
}

// RawEventMessage é o que vai para o JSON de saída em raw_event
type RawEventMessage struct {
	From      string    `json:"from"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// SimpleEvent é o formato final de cada objeto no array de resposta
type SimpleEvent struct {
	RawEvent   RawEventMessage `json:"raw_event"`
	ReceivedAt time.Time       `json:"received_at"`
}

// HandlerHistory2 retorna apenas as mensagens no formato desejado
func HandlerHistory2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Contexto com timeout
	ctx, cancel := context.WithTimeout(r.Context(), database.MONGODB_TIMEOUT)
	defer cancel()

	// Conexão ao MongoDB
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

	// Faz a consulta e ordena por received_at
	col := client.Database(database.GetDB()).Collection("whatsapp_events")
	cursor, err := col.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "received_at", Value: 1}}))
	if err != nil {
		log.Printf("[History] Erro ao buscar histórico: %v", err)
		http.Error(w, "Erro ao buscar histórico", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var result []SimpleEvent

	// Itera pelos documentos
	for cursor.Next(ctx) {
		var doc struct {
			RawEvent   RawEvent  `bson:"raw_event"`
			ReceivedAt time.Time `bson:"received_at"`
		}
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("[History] Erro ao decodificar documento: %v", err)
			continue
		}

		// Para cada entry → change where field == "messages"
		for _, entry := range doc.RawEvent.Entry {
			for _, change := range entry.Changes {
				if change.Field != "messages" {
					continue
				}
				for _, msg := range change.Value.Messages {
					// converte timestamp string (epoch) para time.Time
					secs, err := strconv.ParseInt(msg.Timestamp, 10, 64)
					if err != nil {
						log.Printf("[History] Timestamp inválido '%s': %v", msg.Timestamp, err)
						continue
					}
					ts := time.Unix(secs, 0).UTC()

					// adiciona ao resultado
					result = append(result, SimpleEvent{
						RawEvent: RawEventMessage{
							From:      msg.From,
							Message:   msg.Text.Body,
							Timestamp: ts,
						},
						ReceivedAt: doc.ReceivedAt.UTC(),
					})
				}
			}
		}
	}
	if err := cursor.Err(); err != nil {
		log.Printf("[History] Erro durante iteração do cursor: %v", err)
	}

	// Serializa JSON de saída
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("[History] Erro ao serializar resposta: %v", err)
	}
}
