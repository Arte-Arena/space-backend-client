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

func getUniformById(w http.ResponseWriter, r *http.Request) {
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
	filter := bson.D{{Key: "_id", Value: objectID}}

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

	userIdStr, ok := userId.(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
		})
		return
	}

	if uniform.ClientID != userIdStr {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Você não tem permissão para visualizar este uniforme",
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

	uniformsCollection := client.Database(database.MONGODB_DB_ADMIN).Collection("uniforms")

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
		getUniformById(w, r)
	case http.MethodPatch:
		updatePlayers(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
	}
}
