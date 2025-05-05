package uniforms

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
)

func getUniforms(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIDKey)
	if userId == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Usuário não autorizado",
		})
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
		})
		return
	}

	budgetIDParam := r.URL.Query().Get("id")

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

	uniformsCollection := client.Database(database.GetDB()).Collection("uniforms")

	if budgetIDParam != "" {
		budgetID, err := utils.ParseIntOrDefault(budgetIDParam, 0)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "ID do orçamento inválido",
			})
			return
		}

		filter := bson.D{
			{Key: "budget_id", Value: budgetID},
			{Key: "client_id", Value: userIdStr},
		}

		var uniform schemas.UniformFromDB
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

		uniformResponse := schemas.UniformResponse{
			ID:        uniform.ID.Hex(),
			ClientID:  uniform.ClientID,
			BudgetID:  uniform.BudgetID,
			Sketches:  uniform.Sketches,
			Editable:  uniform.Editable,
			CreatedAt: uniform.CreatedAt,
			UpdatedAt: uniform.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Data: uniformResponse,
		})
		return
	}

	filter := bson.D{{Key: "client_id", Value: userIdStr}}
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := uniformsCollection.Find(ctx, filter, findOptions)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Data: []schemas.UniformResponse{},
		})
		return
	}

	uniformResponses := make([]schemas.UniformResponse, len(uniforms))
	for i, uniform := range uniforms {
		uniformResponses[i] = schemas.UniformResponse{
			ID:        uniform.ID.Hex(),
			ClientID:  uniform.ClientID,
			BudgetID:  uniform.BudgetID,
			Sketches:  uniform.Sketches,
			Editable:  uniform.Editable,
			CreatedAt: uniform.CreatedAt,
			UpdatedAt: uniform.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schemas.ApiResponse{
		Data: uniformResponses,
	})
}

func updatePlayers(w http.ResponseWriter, r *http.Request) {
	uniformID := r.URL.Query().Get("id")
	if uniformID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "ID do uniforme é obrigatório",
		})
		return
	}

	objectID, err := utils.ParseObjectIDFromHex(uniformID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "ID do uniforme inválido",
		})
		return
	}

	userId := r.Context().Value(middlewares.UserIDKey)
	if userId == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Usuário não autorizado",
		})
		return
	}

	var updateRequest schemas.PlayersUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Dados inválidos",
		})
		return
	}

	if len(updateRequest.Updates) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "É necessário informar pelo menos uma atualização de sketch",
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

	uniformsCollection := client.Database(database.GetDB()).Collection("uniforms")

	filter := bson.D{{Key: "_id", Value: objectID}}
	var existingUniform schemas.UniformFromDB
	err = uniformsCollection.FindOne(ctx, filter).Decode(&existingUniform)
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

	userIdStr, ok := userId.(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
		})
		return
	}

	if existingUniform.ClientID != userIdStr {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Você não tem permissão para editar este uniforme",
		})
		return
	}

	if !existingUniform.Editable {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Este uniforme não está disponível para edição",
		})
		return
	}

	sketchMap := make(map[string]int)
	for i, sketch := range existingUniform.Sketches {
		sketchMap[sketch.ID] = i
	}

	updatedSketches := make([]schemas.Sketch, len(existingUniform.Sketches))
	copy(updatedSketches, existingUniform.Sketches)

	for _, update := range updateRequest.Updates {
		sketchIndex, exists := sketchMap[update.SketchID]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Sketch ID não encontrado: " + update.SketchID,
			})
			return
		}

		if len(update.Players) > updatedSketches[sketchIndex].PlayerCount {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "O número de jogadores excede o player_count definido para o sketch " + update.SketchID,
			})
			return
		}

		for i := range update.Players {
			update.Players[i].Ready = false
		}

		updatedSketches[sketchIndex].Players = update.Players
	}

	mongoUpdate := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "sketches", Value: updatedSketches},
			{Key: "updated_at", Value: time.Now()},
			{Key: "editable", Value: false},
		}},
	}

	result, err := uniformsCollection.UpdateOne(ctx, filter, mongoUpdate)
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

	var updatedUniform schemas.UniformFromDB
	err = uniformsCollection.FindOne(ctx, filter).Decode(&updatedUniform)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	uniformResponse := schemas.UniformResponse{
		ID:        updatedUniform.ID.Hex(),
		ClientID:  updatedUniform.ClientID,
		BudgetID:  updatedUniform.BudgetID,
		Sketches:  updatedUniform.Sketches,
		Editable:  updatedUniform.Editable,
		CreatedAt: updatedUniform.CreatedAt,
		UpdatedAt: updatedUniform.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schemas.ApiResponse{
		Data: uniformResponse,
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUniforms(w, r)
	case http.MethodPatch:
		updatePlayers(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
	}
}
