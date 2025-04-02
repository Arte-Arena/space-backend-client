package auth

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func Signin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Método não permitido",
		})
		return
	}

	req := schemas.ClientsRequestToSignin{}
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
	filter := bson.D{{Key: "contact.email", Value: req.Email}}

	result := schemas.ClientsFromMongoDBFindOne{}

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

	err = bcrypt.CompareHashAndPassword([]byte(result.PasswordHash), []byte(req.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Credenciais inválidas",
		})
		return
	}

	accessToken, err := utils.GenerateAccessKey(result.ID.Hex())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro ao gerar token de acesso",
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshKey(result.ID.Hex())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro ao gerar token de atualização",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   15 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func Signout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Método não permitido",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

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

	claims, err := utils.ValidateAccessKey(tokenString)
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

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Método não permitido",
		})
		return
	}

	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Refresh token não encontrado",
		})
		return
	}

	claims, err := utils.ValidateRefreshKey(refreshCookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Refresh token inválido ou expirado",
		})
		return
	}

	newAccessToken, err := utils.GenerateAccessKey(claims.UserId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(utils.ApiResponse{
			Message: "Erro ao gerar novo token de acesso",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Path:     "/",
		MaxAge:   15 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(utils.ApiResponse{
		Message: "Token atualizado com sucesso",
	})
}
