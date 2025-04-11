package clients

import (
	"api/database"
	"api/middlewares"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func getById(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIDKey)
	if userId == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Usuário não autorizado",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGODB_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB),
		})
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection("clients")

	userIdStr, ok := userId.(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
		})
		return
	}

	objectId, err := utils.ParseObjectIDFromHex(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
		})
		return
	}

	filter := bson.D{{Key: "_id", Value: objectId}}
	client := schemas.ClientFromDB{}
	err = collection.FindOne(ctx, filter).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Cliente não encontrado",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	clientResponse := schemas.ClientResponse{
		ID:        client.ID.Hex(),
		Contact:   client.Contact,
		CreatedAt: client.CreatedAt,
		UpdatedAt: client.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas.ApiResponse{
		Data: clientResponse,
	})
}

func update(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middlewares.UserIDKey)
	if userId == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Usuário não autorizado",
		})
		return
	}

	clientFromRequest := schemas.ClientUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&clientFromRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CLIENTS_INVALID_REQUEST_DATA),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGODB_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.ENV_MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB),
		})
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection("clients")

	userIdStr, ok := userId.(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
		})
		return
	}

	objectId, err := utils.ParseObjectIDFromHex(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_USER_ID_FORMAT),
		})
		return
	}

	filter := bson.D{{Key: "_id", Value: objectId}}
	client := schemas.ClientFromDB{}
	err = collection.FindOne(ctx, filter).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Cliente não encontrado",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_TRY_FIND_MONGODB),
		})
		return
	}

	if clientFromRequest.Email != "" && clientFromRequest.Email != client.Contact.Email {
		emailFilter := bson.D{{Key: "contact.email", Value: clientFromRequest.Email}}
		existingClient := schemas.ClientFromDB{}
		err = collection.FindOne(ctx, emailFilter).Decode(&existingClient)
		if err == nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: "Email já cadastrado",
			})
			return
		}
	}

	updateFields := bson.D{}

	if clientFromRequest.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(clientFromRequest.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(schemas.ApiResponse{
				Message: utils.SendInternalError(utils.ERROR_TO_CREATE_PASSWORD_HASH),
			})
			return
		}
		updateFields = append(updateFields, bson.E{Key: "password_hash", Value: string(hashedPassword)})
	}

	if clientFromRequest.Name != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.name", Value: clientFromRequest.Name})
	}
	if clientFromRequest.Email != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.email", Value: clientFromRequest.Email})
	}
	if clientFromRequest.PersonType != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.person_type", Value: clientFromRequest.PersonType})
	}
	if clientFromRequest.IdentityCard != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.identity_card", Value: clientFromRequest.IdentityCard})
	}
	if clientFromRequest.CPF != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.cpf", Value: clientFromRequest.CPF})
	}
	if clientFromRequest.CellPhone != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.cell_phone", Value: clientFromRequest.CellPhone})
	}
	if clientFromRequest.ZipCode != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.zip_code", Value: clientFromRequest.ZipCode})
	}
	if clientFromRequest.Address != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.address", Value: clientFromRequest.Address})
	}
	if clientFromRequest.Number != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.number", Value: clientFromRequest.Number})
	}
	if clientFromRequest.Complement != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.complement", Value: clientFromRequest.Complement})
	}
	if clientFromRequest.Neighborhood != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.neighborhood", Value: clientFromRequest.Neighborhood})
	}
	if clientFromRequest.City != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.city", Value: clientFromRequest.City})
	}
	if clientFromRequest.State != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.state", Value: clientFromRequest.State})
	}
	if clientFromRequest.CompanyName != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.company_name", Value: clientFromRequest.CompanyName})
	}
	if clientFromRequest.CNPJ != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.cnpj", Value: clientFromRequest.CNPJ})
	}
	if clientFromRequest.StateRegistration != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.state_registration", Value: clientFromRequest.StateRegistration})
	}
	if clientFromRequest.BillingZipCode != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.billing_zip_code", Value: clientFromRequest.BillingZipCode})
	}
	if clientFromRequest.BillingAddress != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.billing_address", Value: clientFromRequest.BillingAddress})
	}
	if clientFromRequest.BillingNumber != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.billing_number", Value: clientFromRequest.BillingNumber})
	}
	if clientFromRequest.BillingComplement != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.billing_complement", Value: clientFromRequest.BillingComplement})
	}
	if clientFromRequest.BillingNeighborhood != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.billing_neighborhood", Value: clientFromRequest.BillingNeighborhood})
	}
	if clientFromRequest.BillingCity != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.billing_city", Value: clientFromRequest.BillingCity})
	}
	if clientFromRequest.BillingState != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.billing_state", Value: clientFromRequest.BillingState})
	}
	updateFields = append(updateFields, bson.E{Key: "contact.different_billing_address", Value: clientFromRequest.DifferentBillingAddress})
	if clientFromRequest.Status != "" {
		updateFields = append(updateFields, bson.E{Key: "contact.status", Value: clientFromRequest.Status})
	}

	updateFields = append(updateFields, bson.E{Key: "contact.updated_at", Value: time.Now()})

	if len(updateFields) <= 2 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Nenhum campo para atualizar",
		})
		return
	}

	update := bson.D{{Key: "$set", Value: updateFields}}

	updatedClient := schemas.ClientFromDB{}
	err = collection.FindOne(ctx, filter).Decode(&updatedClient)
	if err == nil {
		if updatedClient.Contact.TinyID != "" {
			tinyRequest := utils.UpdateContactFromClient(updatedClient.Contact, updatedClient.Contact.TinyID)
			tinyID, err := utils.UpdateTinyContact(tinyRequest)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(schemas.ApiResponse{
					Message: utils.SendInternalError(utils.ERROR_TINY_API_INTEGRATION),
				})
				return
			} else if tinyID != "" && tinyID != updatedClient.Contact.TinyID {
				updateTinyIDFields := bson.D{{Key: "contact.tiny_id", Value: tinyID}}
				updateTinyID := bson.D{{Key: "$set", Value: updateTinyIDFields}}
				_, updateErr := collection.UpdateOne(ctx, filter, updateTinyID)
				if updateErr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(schemas.ApiResponse{
						Message: utils.SendInternalError(utils.ERROR_TO_UPDATE_CLIENT_TO_MONGODB),
					})
					return
				}
			}
		} else {
			err = utils.RegisterClientInTinyWithID(&updatedClient.Contact)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(schemas.ApiResponse{
					Message: utils.SendInternalError(utils.ERROR_TINY_API_INTEGRATION),
				})
				return
			} else if updatedClient.Contact.TinyID != "" {
				updateTinyIDFields := bson.D{{Key: "contact.tiny_id", Value: updatedClient.Contact.TinyID}}
				updateTinyID := bson.D{{Key: "$set", Value: updateTinyIDFields}}
				_, updateErr := collection.UpdateOne(ctx, filter, updateTinyID)
				if updateErr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(schemas.ApiResponse{
						Message: utils.SendInternalError(utils.ERROR_TO_UPDATE_CLIENT_TO_MONGODB),
					})
					return
				}
			}
		}
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.ERROR_TO_UPDATE_CLIENT_TO_MONGODB),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getById(w, r)
	case http.MethodPatch:
		update(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.HTTP_METHOD_NO_ALLOWED),
		})
	}
}
