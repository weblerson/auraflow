package util

import (
	"context"
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Rdb           *redis.Client
	EncryptionKey []byte
)

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     GetEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		log.Fatal("ENCRYPTION_KEY environment variable is required (32 bytes hex-encoded)")
	}

	var err error
	EncryptionKey, err = hex.DecodeString(keyHex)
	if err != nil || len(EncryptionKey) != 32 {
		log.Fatal("ENCRYPTION_KEY must be 64 hex characters (32 bytes for AES-256)")
	}
}
