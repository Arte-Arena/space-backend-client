package middlewares

import (
	"api/database"
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

type ContextKey string

const (
	UserIDKey ContextKey = "userId"

	ACCESS_TOKEN_COOKIE_EXPIRATION  = 15 * time.Minute
	REFRESH_TOKEN_COOKIE_EXPIRATION = 7 * 24 * time.Hour
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessCookie, err := r.Cookie("access_token")
		if err == nil {
			claims, err := utils.ValidateAccessKey(accessCookie.Value)
			if err == nil {
				ctx := context.WithValue(r.Context(), UserIDKey, claims.UserId)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
		}

		refreshCookie, err := r.Cookie("refresh_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: utils.SendInternalError(utils.MISSING_REFRESH_TOKEN_IN_COOKIES),
			})
			return
		}

		refreshClaims, err := utils.ValidateRefreshKey(refreshCookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: utils.SendInternalError(utils.MIDDLEWARE_REFRESH_TOKEN_INVALID_OR_EXPIRED),
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
				Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB),
			})
			return
		}
		defer client.Disconnect(ctx)

		collection := client.Database(database.MONGODB_DB_ADMIN).Collection("clients")

		userId, err := utils.ParseObjectIDFromHex(refreshClaims.UserId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
			})
			return
		}

		filter := bson.D{{Key: "_id", Value: userId}}

		var result struct {
			RefreshToken string `bson:"refresh_token"`
		}

		err = collection.FindOne(ctx, filter).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(utils.ApiResponse{
					Message: utils.SendInternalError(utils.MIDDLEWARE_REFRESH_TOKEN_INVALID_OR_EXPIRED),
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
			})
			return
		}

		if refreshCookie.Value != result.RefreshToken {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: utils.SendInternalError(utils.REFRESH_TOKEN_NOT_MATCHING_DATABASE),
			})
			return
		}

		newAccessToken, err := utils.GenerateAccessKey(refreshClaims.UserId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: utils.SendInternalError(utils.ERROR_WHEN_GENERATE_ACCESS_TOKEN),
			})
			return
		}

		newRefreshToken, err := utils.GenerateRefreshKey(refreshClaims.UserId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: utils.SendInternalError(utils.ERROR_WHEN_GENERATE_REFRESH_TOKEN),
			})
			return
		}

		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "refresh_token", Value: newRefreshToken},
			{Key: "updated_at", Value: time.Now()},
		}}}

		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(utils.ApiResponse{
				Message: utils.SendInternalError(utils.ERROR_TO_UPDATE_REFRESH_TOKEN),
			})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    newAccessToken,
			Path:     "/",
			MaxAge:   int(ACCESS_TOKEN_COOKIE_EXPIRATION.Seconds()),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			Path:     "/",
			MaxAge:   int(REFRESH_TOKEN_COOKIE_EXPIRATION.Seconds()),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		ctx = context.WithValue(r.Context(), UserIDKey, refreshClaims.UserId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
}
