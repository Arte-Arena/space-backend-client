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
	BudgetIDs    []int         `bson:"budget_ids,omitempty"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}

type ClientCreateRequest struct {
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"-"`
}

type ClientUpdateRequest struct {
	Name                    string `json:"name,omitempty"`
	Email                   string `json:"email,omitempty"`
	Password                string `json:"password,omitempty"`
	PersonType              string `json:"person_type,omitempty"`
	IdentityCard            string `json:"identity_card,omitempty"`
	CPF                     string `json:"cpf,omitempty"`
	CellPhone               string `json:"cell_phone,omitempty"`
	ZipCode                 string `json:"zip_code,omitempty"`
	Address                 string `json:"address,omitempty"`
	Number                  string `json:"number,omitempty"`
	Complement              string `json:"complement,omitempty"`
	Neighborhood            string `json:"neighborhood,omitempty"`
	City                    string `json:"city,omitempty"`
	State                   string `json:"state,omitempty"`
	CompanyName             string `json:"company_name,omitempty"`
	CNPJ                    string `json:"cnpj,omitempty"`
	StateRegistration       string `json:"state_registration,omitempty"`
	BillingZipCode          string `json:"billing_zip_code,omitempty"`
	BillingAddress          string `json:"billing_address,omitempty"`
	BillingNumber           string `json:"billing_number,omitempty"`
	BillingComplement       string `json:"billing_complement,omitempty"`
	BillingNeighborhood     string `json:"billing_neighborhood,omitempty"`
	BillingCity             string `json:"billing_city,omitempty"`
	BillingState            string `json:"billing_state,omitempty"`
	DifferentBillingAddress bool   `json:"different_billing_address,omitempty"`
	Status                  string `json:"status,omitempty"`
}

type ClientCreateModel struct {
	Contact      Contact   `bson:"contact"`
	PasswordHash string    `bson:"password_hash"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

type ClientResponse struct {
	ID         string       `json:"id"`
	Contact    Contact      `json:"contact"`
	BudgetIDs  []int        `json:"budget_ids,omitempty"`
	HasUniform map[int]bool `json:"has_uniform,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

type ClientAddBudgetRequest struct {
	Email    string `json:"email"`
	BudgetID int    `json:"budget_id"`
}

type ClientsByBudgetIDsRequest struct {
	BudgetIDs []int `json:"budget_ids"`
}
