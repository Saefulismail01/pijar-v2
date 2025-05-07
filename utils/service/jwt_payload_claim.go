package service

import (
	"fmt"
	"pijar/model"
	modelutil "pijar/utils/model_util"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtService interface {
	CreateToken(user model.Users) (string, error)
	VerifyToken(token string) (modelutil.JwtPayloadClaim, error)
}

type jwtService struct {
	signingKey     []byte
	applicationName string
	tokenDuration   time.Duration
}

func NewJwtService(key, appName string, duration time.Duration) JwtService {
	return &jwtService{
		signingKey:     []byte(key),
		applicationName: appName,
		tokenDuration:   duration,
	}
}

func (j *jwtService) CreateToken(user model.Users) (string, error) {
	claims := modelutil.JwtPayloadClaim{
		UserId: strconv.Itoa(user.ID),
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.applicationName,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.signingKey)
}

func (j *jwtService) VerifyToken(tokenStr string) (modelutil.JwtPayloadClaim, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &modelutil.JwtPayloadClaim{}, func(token *jwt.Token) (interface{}, error) {
		return j.signingKey, nil
	})
	if err != nil {
		return modelutil.JwtPayloadClaim{}, err
	}

	if claims, ok := token.Claims.(*modelutil.JwtPayloadClaim); ok && token.Valid {
		return *claims, nil
	}
	return modelutil.JwtPayloadClaim{}, fmt.Errorf("invalid token")
}
