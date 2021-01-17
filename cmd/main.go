package main

import (
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/materkov/onlinestat"
)

func main() {
	onlinestat.RedisClient = redis.NewClient(&redis.Options{})

	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatalf("Empty env ACCESS_TOKEN")
	}

	go onlinestat.ServeHTTP()
	onlinestat.FetchForever(accessToken)
}
