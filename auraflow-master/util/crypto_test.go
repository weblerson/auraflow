package util

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("erro ao gerar chave: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{name: "texto simples", plaintext: "hello world"},
		{name: "CPF", plaintext: "52998224725"},
		{name: "texto vazio", plaintext: ""},
		{name: "texto com acentos", plaintext: "Olá, mundo!"},
		{name: "texto longo", plaintext: "Um texto bem longo para testar a criptografia com diferentes tamanhos de entrada"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.plaintext, key)
			if err != nil {
				t.Errorf("Encrypt() error = %v", err)
				return
			}

			if encrypted == tt.plaintext && tt.plaintext != "" {
				t.Error("Encrypt() não criptografou o texto")
				return
			}

			decrypted, err := Decrypt(encrypted, key)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestDecryptComChaveErrada(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	rand.Read(key1)
	rand.Read(key2)

	encrypted, err := Encrypt("secret", key1)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	_, err = Decrypt(encrypted, key2)
	if err == nil {
		t.Error("Decrypt() deveria retornar erro com chave errada")
	}
}

func TestDecryptDadosInvalidos(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	_, err := Decrypt("dados-invalidos", key)
	if err == nil {
		t.Error("Decrypt() deveria retornar erro com dados inválidos")
	}
}

func TestEncryptGeraDadosDiferentes(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	enc1, _ := Encrypt("mesmo texto", key)
	enc2, _ := Encrypt("mesmo texto", key)

	if enc1 == enc2 {
		t.Error("Encrypt() deveria gerar dados diferentes (nonce aleatório)")
	}
}

func TestChaveHex(t *testing.T) {
	keyHex := "e1e4c9cf03ab40d1b1adac5c7cb11e6a82bc0fec6fe7439c99e69aec703407a3"
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		t.Fatalf("hex.DecodeString() error = %v", err)
	}

	if len(key) != 32 {
		t.Errorf("chave deveria ter 32 bytes, tem %d", len(key))
	}

	encrypted, err := Encrypt("test", key)
	if err != nil {
		t.Errorf("Encrypt() error = %v", err)
		return
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Errorf("Decrypt() error = %v", err)
		return
	}

	if decrypted != "test" {
		t.Errorf("Decrypt() = %v, want %v", decrypted, "test")
	}
}

func BenchmarkEncrypt(b *testing.B) {
	key := make([]byte, 32)
	rand.Read(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encrypt("benchmark test text", key)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := make([]byte, 32)
	rand.Read(key)

	encrypted, _ := Encrypt("benchmark test text", key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decrypt(encrypted, key)
	}
}
