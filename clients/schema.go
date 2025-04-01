package clients

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ClientSchema struct {
	ID                      bson.ObjectID   `bson:"_id"`                                 // Obrigatório
	PersonType              string          `bson:"person_type"`                         // Obrigatório - 'J' for Legal entity or 'F' for Individual
	Name                    string          `bson:"name"`                                // Obrigatório
	IdentityCard            string          `bson:"identity_card,omitempty"`             // RG
	CPF                     string          `bson:"cpf,omitempty"`                       // Obrigatório para pessoa física
	Email                   string          `bson:"email"`                               // Obrigatório
	CellPhone               string          `bson:"cell_phone"`                          // Obrigatório
	ZipCode                 string          `bson:"zip_code,omitempty"`                  // CEP
	Address                 string          `bson:"address,omitempty"`                   // Endereço
	Number                  string          `bson:"number,omitempty"`                    // Número
	Complement              string          `bson:"complement,omitempty"`                // Complemento
	Neighborhood            string          `bson:"neighborhood,omitempty"`              // Bairro
	City                    string          `bson:"city,omitempty"`                      // Cidade
	State                   string          `bson:"state,omitempty"`                     // UF/Estado
	CompanyName             string          `bson:"company_name,omitempty"`              // Obrigatório para pessoa jurídica - Razão Social
	CNPJ                    string          `bson:"cnpj,omitempty"`                      // Obrigatório para pessoa jurídica
	StateRegistration       string          `bson:"state_registration,omitempty"`        // Inscrição Estadual
	BillingZipCode          string          `bson:"billing_zip_code,omitempty"`          // CEP de cobrança
	BillingAddress          string          `bson:"billing_address,omitempty"`           // Endereço de cobrança
	BillingNumber           string          `bson:"billing_number,omitempty"`            // Número de cobrança
	BillingComplement       string          `bson:"billing_complement,omitempty"`        // Complemento de cobrança
	BillingNeighborhood     string          `bson:"billing_neighborhood,omitempty"`      // Bairro de cobrança
	BillingCity             string          `bson:"billing_city,omitempty"`              // Cidade de cobrança
	BillingState            string          `bson:"billing_state,omitempty"`             // UF/Estado de cobrança
	DifferentBillingAddress bool            `bson:"different_billing_address,omitempty"` // Endereço de cobrança diferente
	QuoteIDs                []bson.ObjectID `bson:"quote_ids,omitempty"`                 // Lista de IDs de orçamentos associados ao cliente
	Status                  string          `bson:"status,omitempty"`                    // Situação: 'A' for Active, 'I' for Inactive, 'S' for Suspended
	CreatedAt               time.Time       `bson:"created_at"`
	UpdatedAt               time.Time       `bson:"updated_at"`
}

const (
	PersonTypeIndividual  = "F"
	PersonTypeLegalEntity = "J"
	StatusActive          = "A"
	StatusInactive        = "I"
	StatusSuspended       = "S"
)
