package clients

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
	"golang.org/x/crypto/bcrypt"
)

func getAll(w http.ResponseWriter, r *http.Request) {

}

func getById(w http.ResponseWriter, r *http.Request) {

}

func create(w http.ResponseWriter, r *http.Request) {
	clientFromRequest := schemas.ClientsRequestToCreateOne{}
	if err := json.NewDecoder(r.Body).Decode(&clientFromRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Dados inválidos",
		})
		return
	}

	if clientFromRequest.Name == "" || clientFromRequest.Email == "" || clientFromRequest.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Nome, email e senha são obrigatórios",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(clientFromRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro ao processar a senha",
		})
		return
	}

	contactToCreate := schemas.ContactToCreateOne{
		Name:      clientFromRequest.Name,
		Email:     clientFromRequest.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	clientToCreate := schemas.ClientsToCreateOne{
		Contact:      contactToCreate,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGODB_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro ao conectar ao banco de dados",
		})
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.MONGODB_DB_ADMIN).Collection("clients")

	filter := bson.D{{Key: "contact.email", Value: clientFromRequest.Email}}
	existingClient := schemas.ClientsFromMongoDBFindOne{}
	err = collection.FindOne(ctx, filter).Decode(&existingClient)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Email já cadastrado",
		})
		return
	}

	_, err = collection.InsertOne(ctx, clientToCreate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro ao criar cliente",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func update(w http.ResponseWriter, r *http.Request) {

}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Query().Get("id") != "" {
			getById(w, r)
		} else {
			getAll(w, r)
		}
	case http.MethodPost:
		create(w, r)
	case http.MethodPatch:
		update(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Método não permitido",
		})
	}
}
