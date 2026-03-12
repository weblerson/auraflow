package model

import (
	"context"
	"fmt"
	"time"

	"auraflow/util"

	"github.com/redis/go-redis/v9"
)

var cpfTTL = 24 * time.Hour

func StoreCPF(rdb *redis.Client, encryptionKey []byte, chatID int64, cpf string) error {
	encrypted, err := util.Encrypt(cpf, encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt CPF: %w", err)
	}

	ctx := context.Background()
	key := fmt.Sprintf("cpf:%d", chatID)
	return rdb.Set(ctx, key, encrypted, cpfTTL).Err()
}

func SetWaitingForCPF(rdb *redis.Client, chatID int64, waiting bool) error {
	ctx := context.Background()
	key := fmt.Sprintf("waiting_cpf:%d", chatID)

	if !waiting {
		return rdb.Del(ctx, key).Err()
	}
	return rdb.Set(ctx, key, "1", 5*time.Minute).Err()
}

func IsWaitingForCPF(rdb *redis.Client, chatID int64) bool {
	ctx := context.Background()
	key := fmt.Sprintf("waiting_cpf:%d", chatID)

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	return val == "1"
}

func GetCPF(rdb *redis.Client, encryptionKey []byte, chatID int64) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("cpf:%d", chatID)

	encrypted, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get CPF: %w", err)
	}

	return util.Decrypt(encrypted, encryptionKey)
}
