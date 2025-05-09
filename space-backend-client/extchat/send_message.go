// send_message.go
package extchat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SendMessageRequest struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

type send360Response struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

func HandlerSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// 1) decodifica corpo
	var reqBody SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}
	if reqBody.To == "" || reqBody.Body == "" {
		http.Error(w, "Campos 'to' e 'body' são obrigatórios", http.StatusBadRequest)
		return
	}

	// 2) monta payload para 360Dialog
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                reqBody.To,
		"type":              "text",
		"text": map[string]string{
			"body": reqBody.Body,
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Erro ao serializar payload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3) chama 360Dialog
	apiKey := os.Getenv(utils.D360_API_KEY)
	if apiKey == "" {
		log.Println("[HandlerSendMessage] API key não configurada")
		http.Error(w, "Erro de configuração no servidor", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req360, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://waba-v2.360dialog.io/messages",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		http.Error(w, "Erro ao criar requisição externa: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("Accept", "application/json")
	req360.Header.Set("D360-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req360)
	if err != nil {
		log.Printf("[HandlerSendMessage] falha na requisição 360Dialog: %v", err)
		http.Error(w, "Falha ao enviar mensagem", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 4) lê resposta para extrair message ID
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[HandlerSendMessage] erro ao ler resposta: %v", err)
	}
	var resp360 send360Response
	_ = json.Unmarshal(respBytes, &resp360)

	// 5) encaminha resposta ao frontend
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(respBytes); err != nil {
		log.Printf("[HandlerSendMessage] erro ao enviar resposta: %v", err)
	}

	// 6) persiste mensagem enviada no MongoDB (async)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), database.MONGODB_TIMEOUT)
		defer cancel()

		mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
		clientOpts := options.Client().ApplyURI(mongoURI)
		dbClient, err := mongo.Connect(clientOpts)
		if err != nil {
			log.Printf("[SendMessage-Async] erro conectar MongoDB: %v", err)
			return
		}
		defer func() {
			_ = dbClient.Disconnect(ctx)
		}()

		// constrói raw_event similar ao histórico
		now := time.Now().UTC()
		raw := bson.M{
			"entry": []interface{}{
				bson.M{
					"changes": []interface{}{
						bson.M{
							"field": "messages",
							"value": bson.M{
								"messages": []interface{}{
									bson.M{
										"from":      reqBody.To,
										"timestamp": fmt.Sprint(now.Unix()),
										"text":      bson.M{"body": reqBody.Body},
									},
								},
							},
						},
					},
				},
			},
		}

		col := dbClient.Database(database.GetDB()).Collection("whatsapp_events")
		doc := bson.D{
			{Key: "raw_event", Value: raw},
			{Key: "received_at", Value: now},
		}
		if _, err := col.InsertOne(ctx, doc); err != nil {
			log.Printf("[SendMessage-Async] erro ao inserir no MongoDB: %v", err)
		}
	}()
}
