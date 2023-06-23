// Copyright (c) 2023. Tus1688
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var JwtKeyCustomer []byte
var JwtKeyStaff []byte

type JWTClaimAccessTokenCustomer struct {
	Uid string // user id
	jwt.RegisteredClaims
}

type JWTClaimEmailVerification struct {
	Email  string // user email
	Status bool   // email verification status
	jwt.RegisteredClaims
}

type JWTClaimAccessTokenStaff struct {
	Id       uint
	Username string
	FinUser  bool
	InvUser  bool
	SysAdmin bool
	jwt.RegisteredClaims
}

func GenerateJWTAccessTokenCustomer(uid string, id string) (string, error) {
	claims := &JWTClaimAccessTokenCustomer{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID:       id,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKeyCustomer)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GenerateJWTEmailVerification(email string, status bool) (string, error) {
	claims := &JWTClaimEmailVerification{
		Email:  email,
		Status: status,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKeyCustomer)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GenerateJWTAccessTokenStaff(
	id uint, username string, finUser bool, invUser bool, sysAdmin bool, jti string,
) (string, error) {
	claims := &JWTClaimAccessTokenStaff{
		Id:       id,
		Username: username,
		FinUser:  finUser,
		InvUser:  invUser,
		SysAdmin: sysAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID:       jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKeyStaff)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ExtractClaimAccessTokenCustomer(signedToken string) (*JWTClaimAccessTokenCustomer, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaimAccessTokenCustomer{},
		func(token *jwt.Token) (interface{}, error) {
			return JwtKeyCustomer, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaimAccessTokenCustomer)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func ExtractClaimEmailVerification(signedToken string) (*JWTClaimEmailVerification, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaimEmailVerification{},
		func(token *jwt.Token) (interface{}, error) {
			return JwtKeyCustomer, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaimEmailVerification)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func ExtractClaimAccessTokenStaff(signedToken string) (*JWTClaimAccessTokenStaff, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaimAccessTokenStaff{},
		func(token *jwt.Token) (interface{}, error) {
			return JwtKeyStaff, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaimAccessTokenStaff)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
