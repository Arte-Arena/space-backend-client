package auth

import "go.mongodb.org/mongo-driver/v2/bson"

type Request struct {
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
}

type Response struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type FromMongoDBFind struct {
	ID           bson.ObjectID `bson:"_id"`
	Email        string        `bson:"email"`
	PasswordHash string        `bson:"passwordHash"`
}
