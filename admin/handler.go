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
	uniformRequest := schemas.AllowUniformToEditRequest{}

	if err := json.NewDecoder(r.Body).Decode(&uniformRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Dados inválidos",
		})
		return
	}

	if uniformRequest.BudgetID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "BudgetID é obrigatório",
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

	uniformsCollection := client.Database(database.MONGODB_DB_ADMIN).Collection("uniforms")
	filter := bson.D{{Key: "budget_id", Value: uniformRequest.BudgetID}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "editable", Value: true}}}}

	result, err := uniformsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	if result.MatchedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Uniforme não encontrado",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addUniformWithBudgetId(w, r)
	case http.MethodPatch:
		allowUniformToEdit(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
	}
}
