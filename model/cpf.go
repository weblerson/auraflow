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
