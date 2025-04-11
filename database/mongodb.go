package database

import (
	"api/utils"
	"os"
	"time"
)

func GetDB() string {
	environment := os.Getenv(utils.ENV)

	if environment == utils.ENV_RELEASE {
		return "production"
	}

	if environment == utils.ENV_DEVELOPMENT {
		return "development"
	}

	panic("[MongoDB] Invalid DB name")
}

const (
	MONGODB_TIMEOUT = 20 * time.Second
)
