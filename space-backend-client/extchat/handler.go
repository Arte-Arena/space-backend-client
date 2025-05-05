package extchat

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// WebhookEvent representa o wrapper do Cloud API do 360Dialog
// Ref.: https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/payload-examples
type WebhookEvent struct {
	Object string `json:"object"`
	Entry  []struct {
		ID      string `json:"id"`
		Changes []struct {
			Field string `json:"field"`
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					WaID    string `json:"wa_id"`
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
				} `json:"contacts"`
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

// HandlerWhatsapp processa eventos do 360Dialog Cloud API e persiste mensagens no MongoDB
func HandlerWhatsapp(w http.ResponseWriter, r *http.Request) {
	// Apenas POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED)})
		return
	}

	// Lê payload bruto para debugar
	payloadBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{Message: "Erro ao ler payload: " + err.Error()})
		return
	}
	// Log do payload recebido
	log.Printf("[Webhook] payload: %s", string(payloadBytes))

	// Decodifica JSON
	var evt WebhookEvent
	if err := json.Unmarshal(payloadBytes, &evt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{Message: "Requisição inválida: " + err.Error()})
		return
	}

	// Contexto para operações no MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGODB_TIMEOUT)
	defer cancel()

	// Configura cliente MongoDB
	mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(schemas.ApiResponse{Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB)})
		return
	}
	// Garante desconexão
	defer client.Disconnect(ctx)

	col := client.Database(database.GetDB()).Collection("whatsapp_events")

	// Processa cada entry/change/mensagem
	for _, entry := range evt.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}
			for _, msg := range change.Value.Messages {
				doc := bson.D{
					{Key: "entry_id", Value: entry.ID},
					{Key: "phone_number_id", Value: change.Value.Metadata.PhoneNumberID},
					{Key: "from", Value: msg.From},
					{Key: "message_id", Value: msg.ID},
					{Key: "timestamp", Value: msg.Timestamp},
					{Key: "type", Value: msg.Type},
					{Key: "body", Value: msg.Text.Body},
					{Key: "received_at", Value: time.Now()},
				}
				if _, err := col.InsertOne(ctx, doc); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(schemas.ApiResponse{Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB)})
					return
				}
			}
		}
	}

	// Retorna sucesso
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schemas.ApiResponse{Message: "Eventos processados com sucesso"})
}
