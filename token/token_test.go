package token

import (
	"context"
	"encoding/json"
	"refresh-token-service/redis"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var testUserName string
var testPassword string
var rp redis.RedisProxy
var expiredAccessToken string

func setUp(t *testing.T) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../")
	err := viper.ReadInConfig()
	if err != nil {
		t.Error(err)
	}

	testUserName = viper.GetString("credentials.user")
	testPassword = viper.GetString("credentials.password")
	rp = redis.NewRedisClient(viper.GetString("redis.host"), viper.GetString("redis.password"), viper.GetUint32("redis.port"))

	expiredAccessToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjoiZmZkMjc4MDYtMzhjZS00NzdiLTllMDQtMzI2ODY1NjM3ZjdhIiwiZXhwIjoxNjgzNjQxNTkwLCJzY29wZXMiOlsiYXBpOnJlYWQiLCJhcGk6d3JpdGUiXSwidXNlcl9pZCI6IjI2NThhMmFhLWI5MmItNGJkYy1hNjA1LWEzNDVhMzMwNDg0OSJ9.GACjLZOstCbj0Jyng34YoFpYVbLWAzFgU5rwBRoULQA"
}

func tearDown(t *testing.T) {
	resp, err := rp.GetObject(context.TODO(), testUserName)
	if err == nil && resp != nil {
		err := rp.DeleteObject(context.TODO(), testUserName)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGeneratingTokenPair(t *testing.T) {
	setUp(t)
	tearDown(t)

	t.Run("generating new token pair", func(t *testing.T) {
		resp, err := rp.GetObject(context.TODO(), testUserName)
		assert.Nil(t, err)
		assert.Nil(t, resp)

		tokenPairRes, err := Login(testUserName, testPassword, rp)
		assert.NoError(t, err)
		assert.NotNil(t, tokenPairRes)

		resp, err = rp.GetObject(context.TODO(), testUserName)
		assert.Nil(t, err)
		assert.NotNil(t, resp)

		var tokenPairExpected TokenPair
		err = json.Unmarshal(resp, &tokenPairExpected)
		assert.NoError(t, err)

		assert.Equal(t, *tokenPairExpected.Jwt, *tokenPairRes.Jwt)
		assert.Equal(t, *tokenPairExpected.RefreshToken, *tokenPairRes.RefreshToken)

		tearDown(t)
	})

	t.Run("refreshing token pair", func(t *testing.T) {
		refreshToken, err := CreateToken(REFRESH)
		assert.NoError(t, err)
		testTokenPair := TokenPair{
			Jwt:          &expiredAccessToken,
			RefreshToken: &refreshToken,
		}
		err = rp.SetObject(context.TODO(), testUserName, testTokenPair, REDIS_KEY_EXPIRATION_TIME)
		assert.NoError(t, err)

		resp, err := rp.GetObject(context.TODO(), testUserName)
		assert.Nil(t, err)
		assert.NotNil(t, resp)

		tokenPairRes, err := Login(testUserName, testPassword, rp)
		assert.NoError(t, err)
		assert.NotNil(t, tokenPairRes)

		assert.Equal(t, refreshToken, *tokenPairRes.RefreshToken)
		assert.NotEqual(t, *testTokenPair.Jwt, *tokenPairRes.Jwt)

		tearDown(t)
	})
}
