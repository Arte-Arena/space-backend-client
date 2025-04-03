package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ClientsRequestToSignin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ClientsUpdateAfterGenerateRefreshToken struct {
	RefreshToken string `bson:"refresh_token"`
}

type ClientsRequestToSignout struct {
}

type ClientsFromMongoDBFindOne struct {
	ID           bson.ObjectID      `bson:"_id"`
	Contact      ContactToCreateOne `bson:"contact,omitempty"`
	PasswordHash string             `bson:"password_hash"`
}

type ClientsRequestToCreateOne struct {
	Name     string `bson:"name"`
	Email    string `bson:"email"`
	Password string `bson:"password"`
}

type ClientsToCreateOne struct {
	Contact      ContactToCreateOne `bson:"contact"`
	PasswordHash string             `bson:"password_hash"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}

type ContactToCreateOne struct {
	PersonType              string    `bson:"person_type,omitempty"`               // Obrigatório - 'J' for Legal entity or 'F' for Individual
	Name                    string    `bson:"name"`                                // Obrigatório
	IdentityCard            string    `bson:"identity_card,omitempty"`             // RG
	CPF                     string    `bson:"cpf,omitempty"`                       // Obrigatório para pessoa física
	Email                   string    `bson:"email"`                               // Obrigatório
	CellPhone               string    `bson:"cell_phone,omitempty"`                // Obrigatório
	ZipCode                 string    `bson:"zip_code,omitempty"`                  // CEP
	Address                 string    `bson:"address,omitempty"`                   // Endereço
	Number                  string    `bson:"number,omitempty"`                    // Número
	Complement              string    `bson:"complement,omitempty"`                // Complemento
	Neighborhood            string    `bson:"neighborhood,omitempty"`              // Bairro
	City                    string    `bson:"city,omitempty"`                      // Cidade
	State                   string    `bson:"state,omitempty"`                     // UF/Estado
	CompanyName             string    `bson:"company_name,omitempty"`              // Obrigatório para pessoa jurídica - Razão Social
	CNPJ                    string    `bson:"cnpj,omitempty"`                      // Obrigatório para pessoa jurídica
	StateRegistration       string    `bson:"state_registration,omitempty"`        // Inscrição Estadual
	BillingZipCode          string    `bson:"billing_zip_code,omitempty"`          // CEP de cobrança
	BillingAddress          string    `bson:"billing_address,omitempty"`           // Endereço de cobrança
	BillingNumber           string    `bson:"billing_number,omitempty"`            // Número de cobrança
	BillingComplement       string    `bson:"billing_complement,omitempty"`        // Complemento de cobrança
	BillingNeighborhood     string    `bson:"billing_neighborhood,omitempty"`      // Bairro de cobrança
	BillingCity             string    `bson:"billing_city,omitempty"`              // Cidade de cobrança
	BillingState            string    `bson:"billing_state,omitempty"`             // UF/Estado de cobrança
	DifferentBillingAddress bool      `bson:"different_billing_address,omitempty"` // Endereço de cobrança diferente
	Status                  string    `bson:"status,omitempty"`                    // Situação: 'A' for Active, 'I' for Inactive, 'S' for Suspended
	CreatedAt               time.Time `bson:"created_at,omitempty"`
	UpdatedAt               time.Time `bson:"updated_at"`
}
