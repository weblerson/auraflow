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
}

func GenerateRealPixMP(valor float64, referencia string, cpfOrigem string) string {
	token := os.Getenv("MP_ACCESS_TOKEN")
	if token == "" {
		return "Erro: MP_ACCESS_TOKEN não configurado."
	}

	// O CPF de teste validado
	cpfMercadoPago := "91549965204"

	// 🔥 MÁGICA: E-mail dinâmico usando o relógio do sistema.
	// Isso impede o Mercado Pago de dar erro de "transação repetida" ou spam.
	emailDinamico := fmt.Sprintf("cliente.%d@teste.com", time.Now().Unix())

	reqBody := MPPaymentRequest{
		// Se quiser testar PAGANDO de verdade, troque a variável 'valor' por 0.01 aqui embaixo
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

	jsonBody, _ := json.Marshal(reqBody)

	url := "https://api.mercadopago.com/v1/payments"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Chave única para o Mercado Pago não bloquear
	req.Header.Set("X-Idempotency-Key", fmt.Sprintf("pix-%d", time.Now().UnixNano()))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "Erro de conexão com Mercado Pago"
	}
	defer resp.Body.Close()

	var mpResp MPPaymentResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &mpResp)

	// Se deu tudo certo, retorna o Pix Copia e Cola!
	if mpResp.PointOfInteraction.TransactionData.QRCode != "" {
		return mpResp.PointOfInteraction.TransactionData.QRCode
	}

	erroMP := mpResp.Message
	if erroMP == "" {
		erroMP = "Erro desconhecido."
	}

	// Se falhar e você ainda estiver usando o token TEST, o bot vai te dar o aviso!
	avisoExtra := ""
	if strings.HasPrefix(token, "TEST-") {
		avisoExtra = "\n\n⚠️ O Mercado Pago rejeitou o Sandbox. Vá no painel do MP, pegue a Credencial de PRODUÇÃO (APP_USR-...) e coloque no seu .env!"
	}

	return fmt.Sprintf("⚠️ Erro Mercado Pago: %s%s", erroMP, avisoExtra)
}
