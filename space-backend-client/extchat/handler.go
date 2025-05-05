package extchat

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// WebhookRequest representa o payload enviado pelo 360Dialog
// Adaptar este schema caso sua API retorne campos adicionais
type WebhookRequest struct {
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
}

// HandlerWhatsapp processa eventos do 360Dialog e persiste no MongoDB
func HandlerWhatsapp(w http.ResponseWriter, r *http.Request) {
	// Apenas POST é permitido
	switch r.Method {
	case http.MethodPost:
		// segue
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
		return
	}

	// Decodifica payload
	var payload WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Requisição inválida: " + err.Error(),
		})
		return
	}

	// Contexto com timeout para operações no MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGODB_TIMEOUT)
	defer cancel()

	// Conecta ao MongoDB usando variáveis de ambiente e utilitários
	mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB),
		})
		return
	}
	defer client.Disconnect(ctx)

	// Seleciona a coleção de eventos de WhatsApp
	collection := client.Database(database.GetDB()).Collection("whatsapp_events")

	// Insere cada mensagem no banco
	for _, msg := range payload.Messages {
		doc := bson.D{
			{Key: "from", Value: msg.From},
			{Key: "message_id", Value: msg.ID},
			{Key: "timestamp", Value: msg.Timestamp},
			{Key: "type", Value: msg.Type},
			{Key: "body", Value: msg.Text.Body},
			{Key: "received_at", Value: time.Now()},
		}
		if _, err := collection.InsertOne(ctx, doc); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
			})
			return
		}
	}

	// Sucesso
	w.WriteHeader(http.StatusOK)
}
