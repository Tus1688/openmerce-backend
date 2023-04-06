package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var JwtKey []byte

type JWTClaim struct {
	Uid string // user id
	jwt.RegisteredClaims
}

func GenerateJWT(uid string, id string) (string, error) {
	claims := &JWTClaim{
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

func ExtractClaim(signedToken string) (*JWTClaim, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
