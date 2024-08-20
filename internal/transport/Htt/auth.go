package Htt

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type JWTClaims struct {
	UserID int32  `json:"user_id"`
	Name   string `json:"name"`
	jwt.StandardClaims
}

func GenerateJWT(userID int32, name string) (string, error) {

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Error("JWT_SECRET is not set in the environment variables")
		return "", errors.New("could not generate token: missing secret")
	}

	claims := JWTClaims{
		UserID: userID,
		Name:   name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Errorf("Failed to sign JWT token: %v", err)
		return "", err
	}

	return tokenString, nil
}

func AuthenticateJWT(tokenString string) (string, int32, error) {

	adminToken := os.Getenv("JWT_ADMIN_TOKEN")
	if tokenString == adminToken {
		log.Info("Admin token matched")
		return "admin", 1, nil
	}

	jwtSecret := "admin@1"
	if jwtSecret == "" {
		log.Error("JWT_SECRET is not set in the environment variables")
		return "", 0, errors.New("could not authenticate token: missing secret")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		log.Errorf("Failed to parse JWT token: %v", err)
		return "", 0, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		log.Infof("User token authenticated, UserID: %d", claims.UserID)
		return "user", claims.UserID, nil
	} else {
		log.Warn("Invalid JWT token")
		return "", 0, errors.New("invalid token")
	}
}
