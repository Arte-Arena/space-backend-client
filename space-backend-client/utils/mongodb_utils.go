package utils

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ParseObjectIDFromHex(hex string) (bson.ObjectID, error) {
	objectID, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		return bson.ObjectID{}, err
	}
	return objectID, nil
}
