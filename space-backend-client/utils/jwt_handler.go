package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ACCESS_TOKEN_EXPIRATION    = 15 * time.Minute
	REFRESH_TOKEN_EXPIRATION   = 168 * time.Hour
	SPACE_ERP_TOKEN_EXPIRATION = 5 * time.Minute
)

type Claims struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}

func GenerateAccessKey(userId string) (string, error) {
	accessTokenClaims := Claims{
		userId,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ACCESS_TOKEN_EXPIRATION)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    os.Getenv(TOKEN_ISSUER),
			Audience:  []string{os.Getenv(TOKEN_AUDIENCE)},
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv(ACCESS_TOKEN_SECRET)))
	if err != nil {
		return "", err
	}

	return accessTokenString, nil
}

func GenerateRefreshKey(userId string) (string, error) {
	refreshTokenClaims := Claims{
		userId,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(REFRESH_TOKEN_EXPIRATION)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    os.Getenv(TOKEN_ISSUER),
			Audience:  []string{os.Getenv(TOKEN_AUDIENCE)},
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv(REFRESH_TOKEN_SECRET)))
	if err != nil {
		return "", err
	}

	return refreshTokenString, nil
}

func ValidateAccessKey(tokenString string) (*Claims, error) {
	claims := &Claims{}

	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}),
		jwt.WithIssuer(os.Getenv(TOKEN_ISSUER)),
		jwt.WithAudience(os.Getenv(TOKEN_AUDIENCE)),
		jwt.WithExpirationRequired(),
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv(ACCESS_TOKEN_SECRET)), nil
	}, parserOptions...)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}

func ValidateRefreshKey(tokenString string) (*Claims, error) {
	claims := &Claims{}

	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}),
		jwt.WithIssuer(os.Getenv(TOKEN_ISSUER)),
		jwt.WithAudience(os.Getenv(TOKEN_AUDIENCE)),
		jwt.WithExpirationRequired(),
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv(REFRESH_TOKEN_SECRET)), nil
	}, parserOptions...)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}
