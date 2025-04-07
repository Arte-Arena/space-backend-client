package admin

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

func addUniformWithBudgetId(w http.ResponseWriter, r *http.Request) {
	adminKey := r.Header.Get("X-Admin-Key")
	if adminKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Chave de administrador não fornecida",
		})
		return
	}

	envAdminKey := os.Getenv(utils.ADMIN_KEY)
	if adminKey != envAdminKey {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Chave de administrador inválida",
		})
		return
	}

	uniformRequest := schemas.AdminUniformCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&uniformRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Dados inválidos",
		})
		return
	}

	if uniformRequest.ClientEmail == "" || uniformRequest.BudgetID == 0 || len(uniformRequest.Sketches) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Email do cliente, BudgetID e Sketches são obrigatórios",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGODB_TIMEOUT)
	defer cancel()

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

	clientsCollection := client.Database(database.MONGODB_DB_ADMIN).Collection("clients")
	filter := bson.D{{Key: "contact.email", Value: uniformRequest.ClientEmail}}
	existingClient := schemas.ClientFromDB{}
	err = clientsCollection.FindOne(ctx, filter).Decode(&existingClient)
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

	uniformsCollection := client.Database(database.MONGODB_DB_ADMIN).Collection("uniforms")
	uniformToCreate := schemas.UniformFromDB{
		ClientID:  existingClient.ID.Hex(),
		BudgetID:  uniformRequest.BudgetID,
		Sketches:  uniformRequest.Sketches,
		Editable:  uniformRequest.Editable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = uniformsCollection.InsertOne(ctx, uniformToCreate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func allowUniformToEdit(w http.ResponseWriter, r *http.Request) {

}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addUniformWithBudgetId(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
	}
}
