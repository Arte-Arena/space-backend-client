package clients

import (
	"api/database"
	"api/middlewares"
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
	userId := r.Context().Value(middlewares.UserIDKey)
	if userId == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Usuário não autorizado",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGODB_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB),
		})
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.MONGODB_DB_ADMIN).Collection("clients")

	userIdStr, ok := userId.(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
		})
		return
	}

	objectId, err := utils.ParseObjectIDFromHex(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
		})
		return
	}

	filter := bson.D{{Key: "_id", Value: objectId}}
	client := schemas.ClientFromDB{}
	err = collection.FindOne(ctx, filter).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Cliente não encontrado",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	clientResponse := schemas.ClientResponse{
		ID:        client.ID.Hex(),
		Contact:   client.Contact,
		CreatedAt: client.CreatedAt,
		UpdatedAt: client.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas.ApiResponse{
		Data: clientResponse,
	})
}

func create(w http.ResponseWriter, r *http.Request) {
	clientFromRequest := schemas.ClientCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&clientFromRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CLIENTS_INVALID_REQUEST_DATA),
		})
		return
	}

	if clientFromRequest.Name == "" || clientFromRequest.Email == "" || clientFromRequest.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Nome, email e senha são obrigatórios",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(clientFromRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_CREATE_PASSWORD_HASH),
		})
		return
	}

	contactToCreate := schemas.Contact{
		Name:      clientFromRequest.Name,
		Email:     clientFromRequest.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	clientToCreate := schemas.ClientCreateModel{
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
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB),
		})
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.MONGODB_DB_ADMIN).Collection("clients")

	filter := bson.D{{Key: "contact.email", Value: clientFromRequest.Email}}
	existingClient := schemas.ClientFromDB{}
	err = collection.FindOne(ctx, filter).Decode(&existingClient)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Email já cadastrado",
		})
		return
	}

	_, err = collection.InsertOne(ctx, clientToCreate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_INSERT_CLIENT_TO_MONGODB),
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
		getById(w, r)
	case http.MethodPost:
		create(w, r)
	case http.MethodPatch:
		update(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
	}
}
