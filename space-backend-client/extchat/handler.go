package extchat

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"api/ws"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var Hub *ws.Hub

func FormatRawPayload(data []byte, receivedAt time.Time) ([]SimpleEvent, error) {
	// 1) Unmarshal data em RawEvent
	var raw RawEvent
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	// 2) Mesmo loop de HandlerHistory2, usando receivedAt
	var out []SimpleEvent
	for _, entry := range raw.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}
			for _, msg := range change.Value.Messages {
				secs, err := strconv.ParseInt(msg.Timestamp, 10, 64)
				if err != nil {
					continue
				}
				out = append(out, SimpleEvent{
					RawEvent: RawEventMessage{
						From:      msg.From,
						Message:   msg.Text.Body,
						Timestamp: time.Unix(secs, 0).UTC(),
					},
					ReceivedAt: receivedAt.UTC(),
				})
			}
		}
	}
	return out, nil
}

// HandlerWhatsapp processa o webhook e faz braodcast via WS
func HandlerWhatsapp(w http.ResponseWriter, r *http.Request) {
	// Somente POST permitido
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED)})
		return
	}

	// Lê payload completo do corpo da requisição
	payloadBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("[Webhook] Respondendo status %d Erro ao ler payload: %v", http.StatusBadRequest, err)
		json.NewEncoder(w).Encode(schemas.ApiResponse{Message: "Erro ao ler payload: " + err.Error()})
		return
	}

	// Responde imediatamente 200 OK ao provedor para evitar retries
	w.WriteHeader(http.StatusOK)
	log.Printf("[Webhook] Respondendo status %d OK", http.StatusOK)

	// Processamento assíncrono para gravar no MongoDB sem bloquear a resposta
	go func(data []byte) {

		// if Hub != nil {
		// 	Hub.Broadcast(data)
		// }

		// 1) Formata o payload
		now := time.Now()
		events, err := FormatRawPayload(data, now)
		if err != nil {
			log.Printf("[WS-Format] erro ao formatar payload: %v", err)
			return
		}
		// 2) Serializa e broadcast
		outBytes, _ := json.Marshal(events)

		// <<< aqui o log para conferir o que será enviado
		log.Printf("[WebSocket] enviando payload transformado: %s", string(outBytes))

		// 3) transmite via WS
		Hub.Broadcast(outBytes)

		// Parse genérico do JSON para capturar todo o payload
		var rawEvent map[string]interface{}
		if err := json.Unmarshal(data, &rawEvent); err != nil {
			log.Printf("[Webhook-Async] Erro ao decodificar JSON bruto: %v", err)
			return
		}

		// Contexto para operação no MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), database.MONGODB_TIMEOUT)
		defer cancel()

		// Configura e conecta ao cliente MongoDB
		mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
		clientOpts := options.Client().ApplyURI(mongoURI)
		client, err := mongo.Connect(clientOpts)
		if err != nil {
			log.Printf("[Webhook-Async] Erro ao conectar no MongoDB: %v", err)
			return
		}
		defer func() {
			if err := client.Disconnect(ctx); err != nil {
				log.Printf("[Webhook-Async] Erro ao desconectar MongoDB: %v", err)
			}
		}()

		// Seleciona coleção e insere documento com raw_event
		col := client.Database(database.GetDB()).Collection("whatsapp_events")
		doc := bson.D{
			{Key: "raw_event", Value: rawEvent},
			{Key: "received_at", Value: time.Now()},
		}
		if _, err := col.InsertOne(ctx, doc); err != nil {
			log.Printf("[Webhook-Async] Erro ao inserir documento no MongoDB: %v", err)
		}
	}(payloadBytes)
}
