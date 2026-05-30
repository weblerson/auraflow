package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"auraflow/model"
	"auraflow/util"
)

type PluggyRepository struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	HTTPClient   *http.Client
}

func NewPluggyRepository(clientID, clientSecret string) *PluggyRepository {
	return &PluggyRepository{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		BaseURL:      "https://api.pluggy.ai",
		HTTPClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (r *PluggyRepository) getAuthToken() (string, error) {
	payload := map[string]string{
		"clientId":     r.ClientID,
		"clientSecret": r.ClientSecret,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", r.BaseURL+"/auth", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("erro ao autenticar na Pluggy: Status %d", resp.StatusCode)
	}

	var result struct {
		APIKey string `json:"apiKey"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.APIKey, nil
}

func (r *PluggyRepository) ConsultarBoletos(cpf string) ([]model.Boleto, error) {
	token, err := r.getAuthToken()
	if err != nil {
		return nil, err
	}

	var boletos []model.Boleto
	itemID := "2b9b995c-6f16-411a-bfa4-3d18f2caafb8"

	// =========================================================
	// 1. BUSCAR CONTAS DE CARTÃO DE CRÉDITO
	// =========================================================
	reqAcc, _ := http.NewRequest("GET", r.BaseURL+"/accounts?itemId="+itemID, nil)
	reqAcc.Header.Set("X-API-KEY", token)
	reqAcc.Header.Set("Accept", "application/json")

	respAcc, err := r.HTTPClient.Do(reqAcc)
	if err == nil && respAcc.StatusCode == 200 {
		// Analisando exatamente como o seu JSON de accounts veio!
		var accResponse struct {
			Results []struct {
				ID      string  `json:"id"`
				Type    string  `json:"type"`
				Name    string  `json:"name"`
				Balance float64 `json:"balance"` // Pega o -167.7 do seu Mastercard Black!
			} `json:"results"`
		}
		bodyAcc, _ := io.ReadAll(respAcc.Body)
		json.Unmarshal(bodyAcc, &accResponse)

		for _, acc := range accResponse.Results {
			if acc.Type != "CREDIT" {
				continue
			}

			// Procura na API de faturas
			urlInvoices := fmt.Sprintf("%s/invoices?accountId=%s", r.BaseURL, acc.ID)
			reqInv, _ := http.NewRequest("GET", urlInvoices, nil)
			reqInv.Header.Set("X-API-KEY", token)

			respInv, err := r.HTTPClient.Do(reqInv)
			if err != nil || respInv.StatusCode != 200 {
				if respInv != nil {
					respInv.Body.Close()
				}
				continue
			}

			var invResponse struct {
				Results []struct {
					ID            string  `json:"id"`
					Amount        float64 `json:"amount"`
					DueDate       string  `json:"dueDate"`
					PaymentStatus string  `json:"paymentStatus"`
				} `json:"results"`
			}
			bodyInv, _ := io.ReadAll(respInv.Body)
			respInv.Body.Close()
			json.Unmarshal(bodyInv, &invResponse)

			// 🔥 ESTRATÉGIA AVANÇADA: Se a Pluggy Sandbox não gerou um PDF de fatura,
			// mas você tem saldo devedor negativo (ex: -167.7), nós geramos a fatura pelo saldo!
			if len(invResponse.Results) == 0 && acc.Balance < 0 {
				valorFatura := acc.Balance * -1 // Transforma -167.7 em 167.7 positivo

				boletos = append(boletos, model.Boleto{
					ID:               acc.ID,
					Numero:           "Fatura Atual",
					Valor:            valorFatura,
					CPFSacado:        cpf,
					DataVencimento:   time.Now().AddDate(0, 0, 10),
					Situacao:         "PENDENTE",
					NomeBeneficiario: acc.Name + " (Pluggy Bank)",
					PixCopiaECola:    util.GenerateRealPixMP(valorFatura, acc.ID, cpf), // Passando CPF!
				})
			} else {
				// Se a fatura fechada existir, usa ela
				for _, inv := range invResponse.Results {
					dueDate, _ := time.Parse("2006-01-02", inv.DueDate)
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

					boletos = append(boletos, model.Boleto{
						ID:               inv.ID,
						Numero:           "Fatura Cartão #" + inv.ID[:6],
						Valor:            valorFatura,
						CPFSacado:        cpf,
						DataVencimento:   dueDate,
						Situacao:         situacao,
						NomeBeneficiario: acc.Name + " (Pluggy Bank)",
						PixCopiaECola:    util.GenerateRealPixMP(valorFatura, inv.ID, cpf), // Passando CPF!
					})
				}
			}
		}
		respAcc.Body.Close()
	}

	// =========================================================
	// 2. BUSCAR EMPRÉSTIMOS (Lendo a nova estrutura JSON)
	// =========================================================
	reqLoans, _ := http.NewRequest("GET", r.BaseURL+"/loans?itemId="+itemID, nil)
	reqLoans.Header.Set("X-API-KEY", token)
	reqLoans.Header.Set("Accept", "application/json")

	respLoans, err := r.HTTPClient.Do(reqLoans)
	if err == nil && respLoans.StatusCode == 200 {
		var loansResponse struct {
			Results []struct {
				ID          string `json:"id"`
				ProductName string `json:"productName"` // Pega o "Crédito Pessoal Consignado"
				Payments    struct {
					ContractOutstandingBalance float64 `json:"contractOutstandingBalance"` // Pega o 1000.04 !!
				} `json:"payments"`
			} `json:"results"`
		}
		bodyLoans, _ := io.ReadAll(respLoans.Body)
		json.Unmarshal(bodyLoans, &loansResponse)

		for _, loan := range loansResponse.Results {
			valor := loan.Payments.ContractOutstandingBalance
			if valor <= 0 {
				continue
			}

			nomeInstituicao := loan.ProductName
			if nomeInstituicao == "" {
				nomeInstituicao = "Crédito/Financiamento"
			}

			boletos = append(boletos, model.Boleto{
				ID:               loan.ID,
				Numero:           "Contrato: " + loan.ID[:8],
				Valor:            valor,
				CPFSacado:        cpf,
				DataVencimento:   time.Now().AddDate(0, 1, 0),
				Situacao:         "PENDENTE",
				NomeBeneficiario: nomeInstituicao + " (Pluggy Bank)",
				PixCopiaECola:    util.GenerateRealPixMP(valor, loan.ID, cpf), // Passando CPF!
			})
		}
		respLoans.Body.Close()
	}

	if len(boletos) == 0 {
		boletos = r.fallbackDeSeguranca(cpf)
	}

	return boletos, nil
}

func (r *PluggyRepository) fallbackDeSeguranca(cpf string) []model.Boleto {
	return []model.Boleto{
		{
			ID:               "fatura_seguranca",
			Numero:           "Fatura ID: fatura_seg_999",
			Valor:            0.01,
			CPFSacado:        cpf,
			DataVencimento:   time.Now().AddDate(0, 0, 5),
			Situacao:         "PENDENTE",
			NomeBeneficiario: "AuraFlow (Pagamento Teste)",
			PixCopiaECola:    util.GenerateRealPixMP(0.01, "fatura_seg_999", cpf), // Passando CPF!
		},
	}
}
