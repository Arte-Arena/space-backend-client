package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PackageType string

const (
	PackageTypeStart    PackageType = "Start"
	PackageTypePrata    PackageType = "Prata"
	PackageTypeOuro     PackageType = "Ouro"
	PackageTypeDiamante PackageType = "Diamante"
	PackageTypePro      PackageType = "Pro"
	PackageTypePremium  PackageType = "Premium"
)

type Player struct {
	Gender       string `json:"gender" bson:"gender"`
	Name         string `json:"name" bson:"name"`
	ShirtSize    string `json:"shirt_size" bson:"shirt_size"`
	Number       string `json:"number" bson:"number"`
	ShortsSize   string `json:"shorts_size" bson:"shorts_size"`
	Ready        bool   `json:"ready" bson:"ready"`
	Observations string `json:"observations,omitempty" bson:"observations,omitempty"`
}

type Sketch struct {
	ID          string      `json:"id" bson:"id"`
	PackageType PackageType `json:"package_type" bson:"package_type"`
	Players     []Player    `json:"players" bson:"players"`
}

type UniformFromDB struct {
	ID        bson.ObjectID `bson:"_id"`
	ClientID  string        `bson:"client_id"`
	BudgetID  int           `bson:"budget_id"`
	Sketches  []Sketch      `bson:"sketches"`
	Editable  bool          `bson:"editable"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

type UniformCreateRequest struct {
	ClientID string   `json:"client_id"`
	BudgetID int      `json:"budget_id"`
	Sketches []Sketch `json:"sketches"`
	Editable bool     `json:"editable"`
}

type UniformUpdateRequest struct {
	ClientID string   `json:"client_id,omitempty"`
	BudgetID int      `json:"budget_id,omitempty"`
	Sketches []Sketch `json:"sketches,omitempty"`
	Editable bool     `json:"editable,omitempty"`
}

type UniformResponse struct {
	ID        string    `json:"id"`
	ClientID  string    `json:"client_id"`
	BudgetID  int       `json:"budget_id"`
	Sketches  []Sketch  `json:"sketches"`
	Editable  bool      `json:"editable"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
