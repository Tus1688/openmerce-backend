package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var JwtKey []byte

type JWTClaimAccessToken struct {
	Uid string // user id
	jwt.RegisteredClaims
}

type JWTClaimEmailVerfication struct {
	Email  string // user email
	Status bool   // email verification status
	jwt.RegisteredClaims
}

func GenerateJWTAccessToken(uid string, id string) (string, error) {
	claims := &JWTClaimAccessToken{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID:       id,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GenerateJWTEmailVerification(email string, status bool) (string, error) {
	claims := &JWTClaimEmailVerfication{
		Email:  email,
		Status: status,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ExtractClaimAccessToken(signedToken string) (*JWTClaimAccessToken, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaimAccessToken{},
		func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaimAccessToken)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func ExtractClaimEmailVerification(signedToken string) (*JWTClaimEmailVerfication, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaimEmailVerfication{},
		func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaimEmailVerfication)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
