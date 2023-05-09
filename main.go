package main

import (
	"encoding/json"
	"fmt"
	"log"
	"refresh-token-service/redis"
	"refresh-token-service/token"

	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	rp := redis.NewRedisClient(viper.GetString("redis.host"), viper.GetString("redis.password"), viper.GetUint32("redis.port"))
	tokenPair, err := token.Login("dummy", "dummy", rp)
	if err != nil {
		panic(err)
	}
	jsonData, err := json.Marshal(&tokenPair)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("login success")
	log.Printf("token pair: %s", string(jsonData))
}
