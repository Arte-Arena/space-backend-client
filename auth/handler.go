package auth

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Método não permitido",
		})
		return
	}

	req := Request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Email e senha são obrigatórios",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Email e senha são obrigatórios",
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
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro ao conectar ao banco de dados",
		})
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.MONGODB_DB_ADMIN).Collection("clients")
	filter := bson.D{{Key: "email", Value: req.Email}}

	result := FromMongoDBFind{}

	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: "Credenciais inválidas",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro interno do servidor",
		})
		return
	}

	if result.Password != req.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Credenciais inválidas",
		})
		return
	}

	accessToken, err := GenerateAccessKey(result.ID.Hex())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro ao gerar token de acesso",
		})
		return
	}

	refreshToken, err := GenerateRefreshKey(result.ID.Hex())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro ao gerar token de atualização",
		})
		return
	}

	response := Response{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Método não permitido",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.ApiResponse{
		Message: "Logout realizado com sucesso",
	})
}

func Authorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Método não permitido",
		})
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Cabeçalho de autorização é obrigatório",
		})
		return
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Formato de autorização inválido",
		})
		return
	}
	tokenString := authHeader[7:]

	claims, err := ValidateAccessKey(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Token inválido ou expirado",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.ApiResponse{
		Message: "Token válido",
		Data: map[string]string{
			"userId": claims.UserId,
		},
	})
}
