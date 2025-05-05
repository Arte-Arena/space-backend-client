package orders

import (
	"api/database"
	"api/middlewares"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func getAllOrders(w http.ResponseWriter, r *http.Request) {
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

	objectId, err := utils.ParseObjectIDFromHex(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
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

	collection := mongoClient.Database(database.GetDB()).Collection("clients")
	filter := bson.D{{Key: "_id", Value: objectId}}

	clientData := schemas.ClientFromDB{}
	err = collection.FindOne(ctx, filter).Decode(&clientData)
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

	if len(clientData.BudgetIDs) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Cliente não possui orçamentos",
			Data: schemas.OrderResponse{
				Resultados:   []schemas.OrderResult{},
				TotalPedidos: 0,
			},
		})
		return
	}

	budgetIDStrings := make([]string, len(clientData.BudgetIDs))
	for i, id := range clientData.BudgetIDs {
		budgetIDStrings[i] = strconv.Itoa(id)
	}
	budgetsIds := strings.Join(budgetIDStrings, ",")

	spaceErpUri := os.Getenv(utils.SPACE_ERP_URI)
	laravelURL := spaceErpUri + "/api/pedidos/consultar-multiplos?orcamento_ids=" + budgetsIds

	req, err := http.NewRequest("GET", laravelURL, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_LARAVEL_API_REQUEST_CREATION),
		})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_LARAVEL_API_COMMUNICATION),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_LARAVEL_API_RESPONSE_READING),
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_LARAVEL_API_RESPONSE_STATUS),
		})
		return
	}

	var orderResponse schemas.OrderResponse
	if err := json.Unmarshal(body, &orderResponse); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_LARAVEL_API_RESPONSE_PARSING),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schemas.ApiResponse{
		Data: orderResponse,
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAllOrders(w, r)
	case http.MethodPatch:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
	}
}
