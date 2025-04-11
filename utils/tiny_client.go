package utils

import (
	"api/schemas"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	TINY_API_URL   = "https://api.tiny.com.br/api2"
	TINY_TIMEOUT   = 10 * time.Second
	TINY_TOKEN_ENV = "TINY_API_TOKEN"
)

type TinyContact struct {
	Sequencia           int    `json:"sequencia"`
	Id                  int    `json:"id,omitempty"`
	Codigo              string `json:"codigo,omitempty"`
	Nome                string `json:"nome"`
	Situacao            string `json:"situacao"`
	Fantasia            string `json:"fantasia,omitempty"`
	TipoPessoa          string `json:"tipo_pessoa,omitempty"`
	Cpf_Cnpj            string `json:"cpf_cnpj,omitempty"`
	Ie                  string `json:"ie,omitempty"`
	Rg                  string `json:"rg,omitempty"`
	Im                  string `json:"im,omitempty"`
	Endereco            string `json:"endereco,omitempty"`
	Numero              string `json:"numero,omitempty"`
	Complemento         string `json:"complemento,omitempty"`
	Bairro              string `json:"bairro,omitempty"`
	Cep                 string `json:"cep,omitempty"`
	Cidade              string `json:"cidade,omitempty"`
	Uf                  string `json:"uf,omitempty"`
	Pais                string `json:"pais,omitempty"`
	Contatos            string `json:"contatos,omitempty"`
	Fone                string `json:"fone,omitempty"`
	Fax                 string `json:"fax,omitempty"`
	Celular             string `json:"celular,omitempty"`
	Email               string `json:"email,omitempty"`
	EmailNfe            string `json:"email_nfe,omitempty"`
	Site                string `json:"site,omitempty"`
	Contribuinte        string `json:"contribuinte,omitempty"`
	EnderecoCobranca    string `json:"endereco_cobranca,omitempty"`
	NumeroCobranca      string `json:"numero_cobranca,omitempty"`
	ComplementoCobranca string `json:"complemento_cobranca,omitempty"`
	BairroCobranca      string `json:"bairro_cobranca,omitempty"`
	CepCobranca         string `json:"cep_cobranca,omitempty"`
	CidadeCobranca      string `json:"cidade_cobranca,omitempty"`
	UfCobranca          string `json:"uf_cobranca,omitempty"`
	Obs                 string `json:"obs,omitempty"`
}

type TinyContactWrapper struct {
	Contato TinyContact `json:"contato"`
}

type TinyClientRequest struct {
	Contatos []TinyContactWrapper `json:"contatos"`
}

type TinyAPIResponse struct {
	Retorno struct {
		StatusProcessamento string `json:"status_processamento"`
		Status              string `json:"status"`
		CodigoErro          int    `json:"codigo_erro,omitempty"`
		Erros               []struct {
			Erro string `json:"erro"`
		} `json:"erros,omitempty"`
		Registros []struct {
			Registro struct {
				Sequencia  string `json:"sequencia"`
				Status     string `json:"status"`
				ID         int    `json:"id,omitempty"`
				CodigoErro int    `json:"codigo_erro,omitempty"`
				Erros      []struct {
					Erro string `json:"erro"`
				} `json:"erros,omitempty"`
			} `json:"registro"`
		} `json:"registros,omitempty"`
	} `json:"retorno"`
}

func RegisterClientInTiny(clientData schemas.ClientCreateRequest) error {
	tinyContact := createTinyClientRequest(clientData)
	_, err := RegisterTinyContact(tinyContact)
	return err
}

func RegisterClientInTinyWithID(contact *schemas.Contact) error {
	tinyRequest := CreateContactFromClient(*contact)
	tinyID, err := RegisterTinyContact(tinyRequest)
	if err != nil {
		return err
	}

	if tinyID != "" {
		contact.TinyID = tinyID
	}

	return nil
}

func createTinyClientRequest(client schemas.ClientCreateRequest) TinyClientRequest {
	tinyContact := TinyContact{
		Sequencia: int(time.Now().UnixNano() % 1000000000),
		Nome:      client.Name,
		Situacao:  "A",
	}

	if client.Email != "" {
		tinyContact.Email = client.Email
	}

	wrapper := TinyContactWrapper{
		Contato: tinyContact,
	}

	return TinyClientRequest{
		Contatos: []TinyContactWrapper{wrapper},
	}
}

func CreateContactFromClient(contact schemas.Contact) TinyClientRequest {
	tinyContact := TinyContact{
		Sequencia: int(time.Now().UnixNano() % 1000000000),
		Nome:      contact.Name,
		Situacao:  "A",
	}

	if contact.Email != "" {
		tinyContact.Email = contact.Email
	}

	if contact.PersonType == "PJ" {
		tinyContact.TipoPessoa = "J"
		if contact.CNPJ != "" {
			tinyContact.Cpf_Cnpj = contact.CNPJ
		}
		if contact.StateRegistration != "" {
			tinyContact.Ie = contact.StateRegistration
		}
	} else if contact.PersonType != "" {
		tinyContact.TipoPessoa = "F"
		if contact.CPF != "" {
			tinyContact.Cpf_Cnpj = contact.CPF
		}
		if contact.IdentityCard != "" {
			tinyContact.Rg = contact.IdentityCard
		}
	}

	if contact.CellPhone != "" {
		tinyContact.Celular = contact.CellPhone
	}

	if contact.ZipCode != "" {
		tinyContact.Cep = contact.ZipCode
	}
	if contact.Address != "" {
		tinyContact.Endereco = contact.Address
	}
	if contact.Number != "" {
		tinyContact.Numero = contact.Number
	}
	if contact.Complement != "" {
		tinyContact.Complemento = contact.Complement
	}
	if contact.Neighborhood != "" {
		tinyContact.Bairro = contact.Neighborhood
	}
	if contact.City != "" {
		tinyContact.Cidade = contact.City
	}
	if contact.State != "" {
		tinyContact.Uf = contact.State
	}

	if contact.DifferentBillingAddress {
		if contact.BillingZipCode != "" {
			tinyContact.CepCobranca = contact.BillingZipCode
		}
		if contact.BillingAddress != "" {
			tinyContact.EnderecoCobranca = contact.BillingAddress
		}
		if contact.BillingNumber != "" {
			tinyContact.NumeroCobranca = contact.BillingNumber
		}
		if contact.BillingComplement != "" {
			tinyContact.ComplementoCobranca = contact.BillingComplement
		}
		if contact.BillingNeighborhood != "" {
			tinyContact.BairroCobranca = contact.BillingNeighborhood
		}
		if contact.BillingCity != "" {
			tinyContact.CidadeCobranca = contact.BillingCity
		}
		if contact.BillingState != "" {
			tinyContact.UfCobranca = contact.BillingState
		}
	}

	wrapper := TinyContactWrapper{
		Contato: tinyContact,
	}

	return TinyClientRequest{
		Contatos: []TinyContactWrapper{wrapper},
	}
}

func RegisterTinyContact(tinyClientRequest TinyClientRequest) (string, error) {
	apiToken := os.Getenv(TINY_TOKEN_ENV)
	if apiToken == "" {
		return "", fmt.Errorf("token da API Tiny não encontrado nas variáveis de ambiente")
	}

	contatoJSON, err := json.Marshal(tinyClientRequest)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar contato para JSON: %v", err)
	}

	data := url.Values{}
	data.Set("token", apiToken)
	data.Set("formato", "json")
	data.Set("contato", string(contatoJSON))

	client := http.Client{
		Timeout: TINY_TIMEOUT,
	}

	resp, err := client.PostForm(
		fmt.Sprintf("%s/contato.incluir.php", TINY_API_URL),
		data,
	)
	if err != nil {
		return "", fmt.Errorf("erro ao chamar API Tiny: %v", err)
	}
	defer resp.Body.Close()

	var apiResponse TinyAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta da API: %v", err)
	}

	if apiResponse.Retorno.Status != "OK" {
		if len(apiResponse.Retorno.Registros) > 0 && apiResponse.Retorno.Registros[0].Registro.Status != "OK" {
			var errMsgs []string
			for _, errItem := range apiResponse.Retorno.Registros[0].Registro.Erros {
				errMsgs = append(errMsgs, errItem.Erro)
			}
			return "", fmt.Errorf("erros na API Tiny: %v", errMsgs)
		}

		if len(apiResponse.Retorno.Erros) > 0 {
			var errMsgs []string
			for _, errItem := range apiResponse.Retorno.Erros {
				errMsgs = append(errMsgs, errItem.Erro)
			}
			return "", fmt.Errorf("erros na API Tiny: %v", errMsgs)
		}

		return "", fmt.Errorf("erro desconhecido na API Tiny")
	}

	if len(apiResponse.Retorno.Registros) > 0 && apiResponse.Retorno.Registros[0].Registro.ID > 0 {
		return strconv.Itoa(apiResponse.Retorno.Registros[0].Registro.ID), nil
	}

	return "", nil
}

func UpdateContactFromClient(contact schemas.Contact, tinyID string) TinyClientRequest {
	tinyContact := TinyContact{
		Sequencia: int(time.Now().UnixNano() % 1000000000),
		Nome:      contact.Name,
		Situacao:  "A",
	}

	if tinyID != "" {
		id, err := strconv.Atoi(tinyID)
		if err == nil {
			tinyContact.Id = id
		}
	}

	if contact.Email != "" {
		tinyContact.Email = contact.Email
	}

	if contact.PersonType == "PJ" {
		tinyContact.TipoPessoa = "J"
		if contact.StateRegistration != "" {
			tinyContact.Ie = contact.StateRegistration
		}
	} else if contact.PersonType != "" {
		tinyContact.TipoPessoa = "F"
		if contact.IdentityCard != "" {
			tinyContact.Rg = contact.IdentityCard
		}
	}

	if contact.CellPhone != "" {
		tinyContact.Celular = contact.CellPhone
	}

	if contact.ZipCode != "" {
		tinyContact.Cep = contact.ZipCode
	}
	if contact.Address != "" {
		tinyContact.Endereco = contact.Address
	}
	if contact.Number != "" {
		tinyContact.Numero = contact.Number
	}
	if contact.Complement != "" {
		tinyContact.Complemento = contact.Complement
	}
	if contact.Neighborhood != "" {
		tinyContact.Bairro = contact.Neighborhood
	}
	if contact.City != "" {
		tinyContact.Cidade = contact.City
	}
	if contact.State != "" {
		tinyContact.Uf = contact.State
	}

	if contact.DifferentBillingAddress {
		if contact.BillingZipCode != "" {
			tinyContact.CepCobranca = contact.BillingZipCode
		}
		if contact.BillingAddress != "" {
			tinyContact.EnderecoCobranca = contact.BillingAddress
		}
		if contact.BillingNumber != "" {
			tinyContact.NumeroCobranca = contact.BillingNumber
		}
		if contact.BillingComplement != "" {
			tinyContact.ComplementoCobranca = contact.BillingComplement
		}
		if contact.BillingNeighborhood != "" {
			tinyContact.BairroCobranca = contact.BillingNeighborhood
		}
		if contact.BillingCity != "" {
			tinyContact.CidadeCobranca = contact.BillingCity
		}
		if contact.BillingState != "" {
			tinyContact.UfCobranca = contact.BillingState
		}
	}

	wrapper := TinyContactWrapper{
		Contato: tinyContact,
	}

	return TinyClientRequest{
		Contatos: []TinyContactWrapper{wrapper},
	}
}

func UpdateTinyContact(tinyClientRequest TinyClientRequest) (string, error) {
	apiToken := os.Getenv(TINY_TOKEN_ENV)
	if apiToken == "" {
		return "", fmt.Errorf("token da API Tiny não encontrado nas variáveis de ambiente")
	}

	contatoJSON, err := json.Marshal(tinyClientRequest)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar contato para JSON: %v", err)
	}

	data := url.Values{}
	data.Set("token", apiToken)
	data.Set("formato", "json")
	data.Set("contato", string(contatoJSON))

	client := http.Client{
		Timeout: TINY_TIMEOUT,
	}

	resp, err := client.PostForm(
		fmt.Sprintf("%s/contato.alterar.php", TINY_API_URL),
		data,
	)
	if err != nil {
		return "", fmt.Errorf("erro ao chamar API Tiny: %v", err)
	}
	defer resp.Body.Close()

	var apiResponse TinyAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta da API: %v", err)
	}

	if apiResponse.Retorno.Status != "OK" {
		if len(apiResponse.Retorno.Registros) > 0 && apiResponse.Retorno.Registros[0].Registro.Status != "OK" {
			var errMsgs []string
			for _, errItem := range apiResponse.Retorno.Registros[0].Registro.Erros {
				errMsgs = append(errMsgs, errItem.Erro)
			}
			return "", fmt.Errorf("erros na API Tiny: %v", errMsgs)
		}

		if len(apiResponse.Retorno.Erros) > 0 {
			var errMsgs []string
			for _, errItem := range apiResponse.Retorno.Erros {
				errMsgs = append(errMsgs, errItem.Erro)
			}
			return "", fmt.Errorf("erros na API Tiny: %v", errMsgs)
		}

		return "", fmt.Errorf("erro desconhecido na API Tiny")
	}

	if len(apiResponse.Retorno.Registros) > 0 && apiResponse.Retorno.Registros[0].Registro.ID > 0 {
		return strconv.Itoa(apiResponse.Retorno.Registros[0].Registro.ID), nil
	}

	return "", nil
}
