package schemas

import (
	"time"
)

type Contact struct {
	PersonType              string    `bson:"person_type,omitempty"`
	Name                    string    `bson:"name"`
	IdentityCard            string    `bson:"identity_card,omitempty"`
	CPF                     string    `bson:"cpf,omitempty"`
	Email                   string    `bson:"email"`
	CellPhone               string    `bson:"cell_phone,omitempty"`
	ZipCode                 string    `bson:"zip_code,omitempty"`
	Address                 string    `bson:"address,omitempty"`
	Number                  string    `bson:"number,omitempty"`
	Complement              string    `bson:"complement,omitempty"`
	Neighborhood            string    `bson:"neighborhood,omitempty"`
	City                    string    `bson:"city,omitempty"`
	State                   string    `bson:"state,omitempty"`
	CompanyName             string    `bson:"company_name,omitempty"`
	CNPJ                    string    `bson:"cnpj,omitempty"`
	StateRegistration       string    `bson:"state_registration,omitempty"`
	BillingZipCode          string    `bson:"billing_zip_code,omitempty"`
	BillingAddress          string    `bson:"billing_address,omitempty"`
	BillingNumber           string    `bson:"billing_number,omitempty"`
	BillingComplement       string    `bson:"billing_complement,omitempty"`
	BillingNeighborhood     string    `bson:"billing_neighborhood,omitempty"`
	BillingCity             string    `bson:"billing_city,omitempty"`
	BillingState            string    `bson:"billing_state,omitempty"`
	DifferentBillingAddress bool      `bson:"different_billing_address,omitempty"`
	Status                  string    `bson:"status,omitempty"`
	CreatedAt               time.Time `bson:"created_at,omitempty"`
	UpdatedAt               time.Time `bson:"updated_at"`
}

type Timestamps struct {
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

type ApiResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}
