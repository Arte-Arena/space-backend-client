package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Player struct {
	Gender       string `json:"gender" bson:"gender"`
	Name         string `json:"name" bson:"name"`
	ShirtSize    string `json:"shirtSize" bson:"shirt_size"`
	Number       string `json:"number" bson:"number"`
	ShortsSize   string `json:"shortsSize" bson:"shorts_size"`
	Ready        bool   `json:"ready" bson:"ready"`
	Observations string `json:"observations,omitempty" bson:"observations,omitempty"`
}

type Sketch struct {
	ID          string   `json:"id" bson:"id"`
	PackageType string   `json:"packageType" bson:"package_type"`
	Players     []Player `json:"players" bson:"players"`
}

type UniformFromDB struct {
	ID        bson.ObjectID `bson:"_id"`
	Sketches  []Sketch      `bson:"sketches"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

type UniformCreateRequest struct {
	Sketches []Sketch `json:"sketches"`
}

type UniformUpdateRequest struct {
	Sketches []Sketch `json:"sketches,omitempty"`
}

type UniformResponse struct {
	ID        string    `json:"id"`
	Sketches  []Sketch  `json:"sketches"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
