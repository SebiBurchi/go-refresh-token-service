package token

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"refresh-token-service/redis"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

var (
	JWTKey = []byte("secret")
)

type TokenType string

const (
	ACCESS_TOKEN_LIFETIME     = 5
	REFRESH_TOKEN_LIFETIME    = 3
	REDIS_KEY_EXPIRATION_TIME = 24 * time.Hour

	ACCESS  TokenType = "ACCESS_TOKEN"
	REFRESH TokenType = "REFRESH_TOKEN"
)

type TokenPair struct {
	Jwt          *string
	RefreshToken *string
}

func (tokenPair TokenPair) MarshalBinary() ([]byte, error) {
	return json.Marshal(tokenPair)
}

func Login(user, password string, rp redis.RedisProxy) (*TokenPair, error) {
	if user != viper.GetString("credentials.user") || password != viper.GetString("credentials.password") {
		return nil, fmt.Errorf("invalid username or password")
	}

	resp, err := rp.GetObject(context.TODO(), user)
	if err != nil {
		return nil, fmt.Errorf("failed to get token pair for %s: %v", user, err)
	}
	if resp == nil {
		newTokenPair, err := persistNewTokenPair(user, rp)
		if err != nil {
			return nil, err
		}
		log.Printf("new token pair was generated for user %s", user)
		return newTokenPair, nil
	}

	var tokenPair TokenPair
	err = json.Unmarshal(resp, &tokenPair)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token pair for %s: %v", user, err)
	}

	tkn, err := jwt.Parse(*tokenPair.Jwt, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if err != nil || !tkn.Valid {
		log.Printf("refreshing invalid jwt token for user %s", user)
		refreshTokenPair, err := refresh(user, *tokenPair.RefreshToken, rp)
		if err != nil {
			return nil, err
		}
		log.Printf("token pair was refreshed for user %s", user)
		return refreshTokenPair, nil
	}

	return &tokenPair, nil
}

func CreateToken(tokenType TokenType) (string, error) {
	now := time.Now()

	var tokenDuration time.Duration
	if tokenType == ACCESS {
		tokenDuration = time.Duration(ACCESS_TOKEN_LIFETIME) * time.Minute
	}
	if tokenType == REFRESH {
		tokenDuration = time.Duration(REFRESH_TOKEN_LIFETIME) * time.Hour
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":        now.Add(tokenDuration).Unix(),
		"scopes":     []string{"api:read", "api:write"},
		"account_id": "ffd27806-38ce-477b-9e04-326865637f7a",
		"user_id":    "2658a2aa-b92b-4bdc-a605-a345a3304849",
	})

	t, err := token.SignedString(JWTKey)
	if err != nil {
		return "", err
	}

	return t, nil
}

func refresh(user, refreshToken string, rp redis.RedisProxy) (*TokenPair, error) {
	refreshTkn, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if err != nil || !refreshTkn.Valid {
		return nil, fmt.Errorf("invalid refresh token for user %s", user)
	}

	err = rp.DeleteObject(context.TODO(), user)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old refresh token for %s: %v", user, err)
	}

	newAccessToken, err := CreateToken(ACCESS)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token based on refresh token for %s: %v", user, err)
	}
	refreshTokenPair := TokenPair{
		Jwt:          &newAccessToken,
		RefreshToken: &refreshToken,
	}

	err = rp.SetObject(context.TODO(), user, refreshTokenPair, REDIS_KEY_EXPIRATION_TIME)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token pair for %s: %v", user, err)
	}

	return &refreshTokenPair, nil

}

func persistNewTokenPair(user string, rp redis.RedisProxy) (*TokenPair, error) {
	accessToken, err := CreateToken(ACCESS)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token for %s: %v", user, err)
	}
	refreshToken, err := CreateToken(REFRESH)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token for %s: %v", user, err)
	}
	newTokenPair := TokenPair{
		Jwt:          &accessToken,
		RefreshToken: &refreshToken,
	}
	err = rp.SetObject(context.TODO(), user, newTokenPair, REDIS_KEY_EXPIRATION_TIME)
	if err != nil {
		return nil, fmt.Errorf("failed to save new token pair for %s: %v", user, err)
	}

	return &newTokenPair, nil
}
