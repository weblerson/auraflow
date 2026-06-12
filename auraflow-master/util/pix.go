package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type MPPaymentRequest struct {
	TransactionAmount float64 `json:"transaction_amount"`
	Description       string  `json:"description"`
	PaymentMethodID   string  `json:"payment_method_id"`
	Payer             Payer   `json:"payer"`
}

type Payer struct {
	Email          string         `json:"email"`
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	Identification Identification `json:"identification"`
}

type Identification struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type MPPaymentResponse struct {
	PointOfInteraction struct {
		TransactionData struct {
			QRCode string `json:"qr_code"`
		} `json:"transaction_data"`
	} `json:"point_of_interaction"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// GenerateRealPixMP gera um PIX via Mercado Pago e retorna o código Copia e Cola
func GenerateRealPixMP(valor float64, referencia string, cpfOrigem string) (string, error) {
	token := os.Getenv("MP_ACCESS_TOKEN")
	if token == "" {
		return "", fmt.Errorf("MP_ACCESS_TOKEN não configurado")
	}

	if valor <= 0 {
		return "", fmt.Errorf("valor do PIX deve ser maior que zero")
	}

	cpfMercadoPago := "91549965204"

	emailDinamico := fmt.Sprintf("cliente.%d@teste.com", time.Now().Unix())

	reqBody := MPPaymentRequest{
		TransactionAmount: valor,
		Description:       "AuraFlow - Fatura",
		PaymentMethodID:   "pix",
		Payer: Payer{
			Email:     emailDinamico,
			FirstName: "Usuario",
			LastName:  "Teste",
			Identification: Identification{
				Type:   "CPF",
				Number: cpfMercadoPago,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	url := "https://api.mercadopago.com/v1/payments"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", fmt.Sprintf("pix-%d", time.Now().UnixNano()))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro de conexão com Mercado Pago: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta do Mercado Pago: %w", err)
	}

	var mpResp MPPaymentResponse
	if err := json.Unmarshal(body, &mpResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta do Mercado Pago: %w", err)
	}

	if qrCode := mpResp.PointOfInteraction.TransactionData.QRCode; qrCode != "" {
		return qrCode, nil
	}

	erroMP := mpResp.Message
	if erroMP == "" {
		erroMP = "Erro desconhecido"
	}

	if mpResp.Status != "" {
		erroMP = fmt.Sprintf("%s (status: %s)", erroMP, mpResp.Status)
	}

	if strings.HasPrefix(token, "TEST-") {
		erroMP += "\n\n⚠️ O Mercado Pago rejeitou o Sandbox. Vá no painel do MP, pegue a Credencial de PRODUÇÃO (APP_USR-...) e coloque no seu .env!"
	}

	return "", fmt.Errorf("erro Mercado Pago: %s", erroMP)
}
