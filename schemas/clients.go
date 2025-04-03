package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ClientLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ClientRefreshTokenUpdate struct {
	RefreshToken string `bson:"refresh_token"`
}

type ClientLogoutRequest struct {
}

type ClientFromDB struct {
	ID           bson.ObjectID `bson:"_id"`
	Contact      Contact       `bson:"contact,omitempty"`
	PasswordHash string        `bson:"password_hash"`
	RefreshToken string        `bson:"refresh_token,omitempty"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}

type ClientCreateRequest struct {
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"-"`
}

type ClientCreateModel struct {
	Contact      Contact   `bson:"contact"`
	PasswordHash string    `bson:"password_hash"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

type ClientResponse struct {
	ID        string    `json:"id"`
	Contact   Contact   `json:"contact"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
