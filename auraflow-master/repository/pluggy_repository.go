package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"auraflow/model"
	"auraflow/util"
)

type PluggyRepository struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	ItemID       string
	HTTPClient   *http.Client
}

func NewPluggyRepository(clientID, clientSecret string) *PluggyRepository {
	itemID := os.Getenv("PLUGGY_ITEM_ID")
	if itemID == "" {
		itemID = "2b9b995c-6f16-411a-bfa4-3d18f2caafb8"
	}

	return &PluggyRepository{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		BaseURL:      "https://api.pluggy.ai",
		ItemID:       itemID,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
	}
}

func (r *PluggyRepository) getAuthToken() (string, error) {
	payload := map[string]string{
		"clientId":     r.ClientID,
		"clientSecret": r.ClientSecret,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", r.BaseURL+"/auth", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição de autenticação: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao conectar com Pluggy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro ao autenticar na Pluggy: status %d", resp.StatusCode)
	}

	var result struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta de autenticação: %w", err)
	}

	if result.APIKey == "" {
		return "", fmt.Errorf("API key vazia na resposta da Pluggy")
	}

	return result.APIKey, nil
}

func (r *PluggyRepository) ConsultarBoletos(cpf string) ([]model.Boleto, error) {
	if r.ItemID == "" {
		return nil, fmt.Errorf("PLUGGY_ITEM_ID não configurado")
	}

	token, err := r.getAuthToken()
	if err != nil {
		return nil, err
	}

	var boletos []model.Boleto

	boletosCartao, err := r.consultarCartoesCredito(token, cpf)
	if err != nil {
		log.Printf("Aviso ao consultar cartões de crédito: %v", err)
	}
	boletos = append(boletos, boletosCartao...)

	boletosEmprestimos, err := r.consultarEmprestimosAPI(token, cpf)
	if err != nil {
		log.Printf("Aviso ao consultar empréstimos: %v", err)
	}
	boletos = append(boletos, boletosEmprestimos...)

	if len(boletos) == 0 {
		boletos = r.fallbackDeSeguranca(cpf)
	}

	return boletos, nil
}

func (r *PluggyRepository) consultarCartoesCredito(token, cpf string) ([]model.Boleto, error) {
	url := fmt.Sprintf("%s/accounts?itemId=%s", r.BaseURL, r.ItemID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de contas: %w", err)
	}
	req.Header.Set("X-API-KEY", token)
	req.Header.Set("Accept", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar contas: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao consultar contas: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta de contas: %w", err)
	}

	var accResponse struct {
		Results []struct {
			ID      string  `json:"id"`
			Type    string  `json:"type"`
			Name    string  `json:"name"`
			Balance float64 `json:"balance"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &accResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar contas: %w", err)
	}

	var boletos []model.Boleto

	for _, acc := range accResponse.Results {
		if acc.Type != "CREDIT" {
			continue
		}

		faturas, err := r.consultarFaturasCartao(token, acc.ID, acc.Name, cpf)
		if err != nil {
			log.Printf("Aviso ao consultar faturas da conta %s: %v", acc.ID, err)
			continue
		}
		boletos = append(boletos, faturas...)
	}

	return boletos, nil
}

func (r *PluggyRepository) consultarFaturasCartao(token, contaID, nomeConta, cpf string) ([]model.Boleto, error) {
	url := fmt.Sprintf("%s/invoices?accountId=%s", r.BaseURL, contaID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de faturas: %w", err)
	}
	req.Header.Set("X-API-KEY", token)

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar faturas: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta de faturas: %w", err)
	}

	var invResponse struct {
		Results []struct {
			ID            string  `json:"id"`
			Amount        float64 `json:"amount"`
			DueDate       string  `json:"dueDate"`
			PaymentStatus string  `json:"paymentStatus"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &invResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar faturas: %w", err)
	}

	var boletos []model.Boleto
	nomeInstituicao := nomeConta + " (Pluggy Bank)"

	for _, inv := range invResponse.Results {
		dueDate, err := time.Parse("2006-01-02", inv.DueDate)
		if err != nil {
			log.Printf("Aviso ao parsear data de vencimento: %v", err)
			continue
		}

		valorFatura := inv.Amount
		if valorFatura < 0 {
			valorFatura = valorFatura * -1
		}
		if valorFatura <= 0 {
			continue
		}

		situacao := "PENDENTE"
		if inv.PaymentStatus == "CLOSED" {
			situacao = "PAGO"
		}

		pixCopiaECola, err := util.GenerateRealPixMP(valorFatura, inv.ID, cpf)
		if err != nil {
			log.Printf("Aviso ao gerar PIX para fatura %s: %v", inv.ID, err)
			pixCopiaECola = "Erro ao gerar PIX"
		}

		boletos = append(boletos, model.Boleto{
			ID:               inv.ID,
			Numero:           "Fatura Cartão #" + inv.ID[:min(6, len(inv.ID))],
			Valor:            valorFatura,
			CPFSacado:        cpf,
			DataVencimento:   dueDate,
			Situacao:         situacao,
			NomeBeneficiario: nomeInstituicao,
			PixCopiaECola:    pixCopiaECola,
		})
	}

	return boletos, nil
}

func (r *PluggyRepository) ConsultarEmprestimos(cpf string) ([]model.Boleto, error) {
	if r.ItemID == "" {
		return nil, fmt.Errorf("PLUGGY_ITEM_ID não configurado")
	}

	token, err := r.getAuthToken()
	if err != nil {
		return nil, err
	}

	boletos, err := r.consultarEmprestimosAPI(token, cpf)
	if err != nil {
		log.Printf("Aviso ao consultar empréstimos: %v", err)
	}

	return boletos, nil
}

func (r *PluggyRepository) ConsultarFaturas(cpf string) ([]model.Boleto, error) {
	if r.ItemID == "" {
		return nil, fmt.Errorf("PLUGGY_ITEM_ID não configurado")
	}

	token, err := r.getAuthToken()
	if err != nil {
		return nil, err
	}

	boletos, err := r.consultarCartoesCredito(token, cpf)
	if err != nil {
		log.Printf("Aviso ao consultar cartões de crédito: %v", err)
	}

	return boletos, nil
}

func (r *PluggyRepository) consultarEmprestimosAPI(token, cpf string) ([]model.Boleto, error) {
	url := fmt.Sprintf("%s/loans?itemId=%s", r.BaseURL, r.ItemID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de empréstimos: %w", err)
	}
	req.Header.Set("X-API-KEY", token)
	req.Header.Set("Accept", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar empréstimos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao consultar empréstimos: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta de empréstimos: %w", err)
	}

	var loansResponse struct {
		Results []struct {
			ID          string `json:"id"`
			ProductName string `json:"productName"`
			Payments    struct {
				ContractOutstandingBalance float64 `json:"contractOutstandingBalance"`
			} `json:"payments"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &loansResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar empréstimos: %w", err)
	}

	var boletos []model.Boleto

	for _, loan := range loansResponse.Results {
		valor := loan.Payments.ContractOutstandingBalance
		if valor <= 0 {
			continue
		}

		nomeInstituicao := loan.ProductName
		if nomeInstituicao == "" {
			nomeInstituicao = "Crédito/Financiamento"
		}

		pixCopiaECola, err := util.GenerateRealPixMP(valor, loan.ID, cpf)
		if err != nil {
			log.Printf("Aviso ao gerar PIX para empréstimo %s: %v", loan.ID, err)
			pixCopiaECola = "Erro ao gerar PIX"
		}

		boletos = append(boletos, model.Boleto{
			ID:               loan.ID,
			Numero:           "Contrato: " + loan.ID[:min(8, len(loan.ID))],
			Valor:            valor,
			CPFSacado:        cpf,
			DataVencimento:   time.Now().AddDate(0, 1, 0),
			Situacao:         "PENDENTE",
			NomeBeneficiario: nomeInstituicao + " (Pluggy Bank)",
			PixCopiaECola:    pixCopiaECola,
		})
	}

	return boletos, nil
}

func (r *PluggyRepository) fallbackDeSeguranca(cpf string) []model.Boleto {
	pixCopiaECola, err := util.GenerateRealPixMP(0.01, "fatura_seg_999", cpf)
	if err != nil {
		pixCopiaECola = "Erro ao gerar PIX"
	}

	return []model.Boleto{
		{
			ID:               "fatura_seguranca",
			Numero:           "Fatura ID: fatura_seg_999",
			Valor:            0.01,
			CPFSacado:        cpf,
			DataVencimento:   time.Now().AddDate(0, 0, 5),
			Situacao:         "PENDENTE",
			NomeBeneficiario: "AuraFlow (Pagamento Teste)",
			PixCopiaECola:    pixCopiaECola,
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
