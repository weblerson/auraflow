package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb           *redis.Client
	encryptionKey []byte
	cpfTTL        = 24 * time.Hour
)

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		log.Fatal("ENCRYPTION_KEY environment variable is required (32 bytes hex-encoded)")
	}

	var err error
	encryptionKey, err = hex.DecodeString(keyHex)
	if err != nil || len(encryptionKey) != 32 {
		log.Fatal("ENCRYPTION_KEY must be 64 hex characters (32 bytes for AES-256)")
	}
}

func storeCPF(chatID int64, cpf string) error {
	encrypted, err := encrypt(cpf, encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt CPF: %w", err)
	}

	ctx := context.Background()
	key := fmt.Sprintf("cpf:%d", chatID)
	return rdb.Set(ctx, key, encrypted, cpfTTL).Err()
}

func getCPF(chatID int64) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("cpf:%d", chatID)

	encrypted, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get CPF: %w", err)
	}

	return decrypt(encrypted, encryptionKey)
}

func setWaitingForCPF(chatID int64, waiting bool) error {
	ctx := context.Background()
	key := fmt.Sprintf("waiting_cpf:%d", chatID)

	if !waiting {
		return rdb.Del(ctx, key).Err()
	}
	return rdb.Set(ctx, key, "1", 5*time.Minute).Err()
}

func isWaitingForCPF(chatID int64) bool {
	ctx := context.Background()
	key := fmt.Sprintf("waiting_cpf:%d", chatID)

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	return val == "1"
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
