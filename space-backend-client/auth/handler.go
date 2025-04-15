package auth

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const (
	ACCESS_TOKEN_COOKIE_EXPIRATION  = 15 * time.Minute
	REFRESH_TOKEN_COOKIE_EXPIRATION = 7 * 24 * time.Hour
)

func Signin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
		return
	}

	req := schemas.ClientLoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Email e senha são obrigatórios",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
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
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB),
		})
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection("clients")
	filter := bson.D{{Key: "contact.email", Value: req.Email}}

	result := schemas.ClientFromDB{}

	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Credenciais inválidas",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.PasswordHash), []byte(req.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Credenciais inválidas",
		})
		return
	}

	accessToken, err := utils.GenerateAccessKey(result.ID.Hex())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_WHEN_GENERATE_ACCESS_TOKEN),
		})
		return
	}

	refreshToken, err := utils.GenerateRefreshKey(result.ID.Hex())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_WHEN_GENERATE_REFRESH_TOKEN),
		})
		return
	}

	filter = bson.D{{Key: "_id", Value: result.ID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "refresh_token", Value: refreshToken},
		{Key: "updated_at", Value: time.Now()},
	}}}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_UPDATE_REFRESH_TOKEN),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(ACCESS_TOKEN_COOKIE_EXPIRATION),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   int(REFRESH_TOKEN_COOKIE_EXPIRATION),
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
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
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
	json.NewEncoder(w).Encode(schemas.ApiResponse{
		Message: "Logout realizado com sucesso",
	})
}

func Authorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
		return
	}

	accessToken, err := r.Cookie("access_token")
	if err != nil || accessToken.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.MISSING_AUTHORIZATION_HEADER),
		})
		return
	}

	_, errValidate := utils.ValidateAccessKey(accessToken.Value)
	if errValidate != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ACCESS_TOKEN_INVALID_OR_EXPIRED),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Signup(w http.ResponseWriter, r *http.Request) {
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

	collection := mongoClient.Database(database.GetDB()).Collection("clients")

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

	err = utils.RegisterClientInTinyWithID(&contactToCreate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TINY_API_INTEGRATION),
		})
		return
	}

	if contactToCreate.TinyID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TINY_ID_NOT_GENERATED),
		})
		return
	}

	clientToCreate.Contact.TinyID = contactToCreate.TinyID

	_, err = collection.InsertOne(ctx, clientToCreate)
	if err != nil {
		log.Printf("Erro ao criar cliente: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_INSERT_CLIENT_TO_MONGODB),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
