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

	existingUniformFilter := bson.D{
		{Key: "client_id", Value: existingClient.ID.Hex()},
		{Key: "budget_id", Value: uniformRequest.BudgetID},
	}

	existingUniform := schemas.UniformFromDB{}
	err = uniformsCollection.FindOne(ctx, existingUniformFilter).Decode(&existingUniform)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Já existe um uniforme cadastrado para este cliente com este orçamento",
		})
		return
	} else if err != mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	uniformToCreate := schemas.UniformToDB{
		ClientID:  existingClient.ID.Hex(),
		BudgetID:  uniformRequest.BudgetID,
		Sketches:  uniformRequest.Sketches,
		Editable:  true,
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

func updatePlayersData(w http.ResponseWriter, r *http.Request) {
	uniformRequest := schemas.PlayersUpdateRequest{}

	if err := json.NewDecoder(r.Body).Decode(&uniformRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Dados inválidos",
		})
		return
	}

	budgetIDStr := r.URL.Query().Get("budget_id")
	if budgetIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "BudgetID é obrigatório",
		})
		return
	}

	budgetID, err := utils.ParseIntOrDefault(budgetIDStr, 0)
	if err != nil || budgetID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "BudgetID inválido",
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
	filter := bson.D{{Key: "budget_id", Value: budgetID}}

	uniform := schemas.UniformFromDB{}
	err = uniformsCollection.FindOne(ctx, filter).Decode(&uniform)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Uniforme não encontrado",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	updateDoc := bson.D{}

	editableParam := r.URL.Query().Get("editable")
	if editableParam == "true" {
		updateDoc = append(updateDoc, bson.E{Key: "editable", Value: true})
	}

	if len(uniformRequest.Updates) > 0 {
		for _, update := range uniformRequest.Updates {
			for i, sketch := range uniform.Sketches {
				if sketch.ID == update.SketchID {
					uniform.Sketches[i].Players = update.Players
				}
			}
		}
		updateDoc = append(updateDoc, bson.E{Key: "sketches", Value: uniform.Sketches})
	}

	if len(updateDoc) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Nenhuma atualização fornecida",
		})
		return
	}

	updateDoc = append(updateDoc, bson.E{Key: "updated_at", Value: time.Now()})
	update := bson.D{{Key: "$set", Value: updateDoc}}

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

func getUniformsByBudgetId(w http.ResponseWriter, r *http.Request) {
	budgetIDStr := r.URL.Query().Get("budget_id")
	if budgetIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "BudgetID é obrigatório",
		})
		return
	}

	budgetID, err := utils.ParseIntOrDefault(budgetIDStr, 0)
	if err != nil || budgetID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "BudgetID inválido",
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
	filter := bson.D{{Key: "budget_id", Value: budgetID}}

	cursor, err := uniformsCollection.Find(ctx, filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}
	defer cursor.Close(ctx)

	var uniforms []schemas.UniformFromDB
	if err = cursor.All(ctx, &uniforms); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	if len(uniforms) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Nenhum uniforme encontrado para este orçamento",
		})
		return
	}

	var uniformResponses []schemas.UniformResponse
	for _, uniform := range uniforms {
		uniformResponses = append(uniformResponses, schemas.UniformResponse{
			ID:        uniform.ID.Hex(),
			ClientID:  uniform.ClientID,
			BudgetID:  uniform.BudgetID,
			Sketches:  uniform.Sketches,
			Editable:  uniform.Editable,
			CreatedAt: uniform.CreatedAt,
			UpdatedAt: uniform.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schemas.ApiResponse{
		Data: uniformResponses,
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addUniformWithBudgetId(w, r)
	case http.MethodPatch:
		updatePlayersData(w, r)
	case http.MethodGet:
		getUniformsByBudgetId(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
	}
}
