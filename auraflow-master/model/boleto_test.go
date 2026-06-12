package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBoletoJSON(t *testing.T) {
	vencimento := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	pagamento := time.Date(2026, 3, 14, 14, 30, 0, 0, time.UTC)

	boleto := Boleto{
		ID:               "BOL001",
		Numero:           "12345.67890.00000.000000.000001 1",
		Valor:            150.00,
		NomeSacado:       "JOAO SILVA",
		CPFSacado:        "12345678900",
		DataVencimento:   vencimento,
		DataPagamento:    &pagamento,
		ValorPago:        150.00,
		Situacao:         "PAGO",
		NomeBeneficiario: "TECHBANK S.A.",
		PixCopiaECola:    "00020126580014br.gov.bcb.pix0136123e4567-e12b-4a56-8000-123456789abc5204000053039865405150.005802BR5913JOAO SILVA6008SAO PAULO62070503***63041D4D",
	}

	jsonData, err := json.Marshal(boleto)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var boletoDeserializado Boleto
	if err := json.Unmarshal(jsonData, &boletoDeserializado); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if boletoDeserializado.ID != boleto.ID {
		t.Errorf("ID = %v, want %v", boletoDeserializado.ID, boleto.ID)
	}
	if boletoDeserializado.Valor != boleto.Valor {
		t.Errorf("Valor = %v, want %v", boletoDeserializado.Valor, boleto.Valor)
	}
	if boletoDeserializado.Situacao != boleto.Situacao {
		t.Errorf("Situacao = %v, want %v", boletoDeserializado.Situacao, boleto.Situacao)
	}
	if boletoDeserializado.DataVencimento != boleto.DataVencimento {
		t.Errorf("DataVencimento = %v, want %v", boletoDeserializado.DataVencimento, boleto.DataVencimento)
	}
	if boletoDeserializado.DataPagamento == nil {
		t.Error("DataPagamento não deveria ser nil")
	} else if !boletoDeserializado.DataPagamento.Equal(*boleto.DataPagamento) {
		t.Errorf("DataPagamento = %v, want %v", *boletoDeserializado.DataPagamento, *boleto.DataPagamento)
	}
}

func TestBoletoSemPagamento(t *testing.T) {
	boleto := Boleto{
		ID:               "BOL002",
		Numero:           "12345.67890.00000.000000.000002 2",
		Valor:            275.50,
		NomeSacado:       "MARIA SANTOS",
		CPFSacado:        "98765432100",
		DataVencimento:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		DataPagamento:    nil,
		ValorPago:        0,
		Situacao:         "PENDENTE",
		NomeBeneficiario: "ITAU UNIBANCO",
		PixCopiaECola:    "pix-copia-e-cola-aqui",
	}

	jsonData, err := json.Marshal(boleto)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	jsonStr := string(jsonData)
	if jsonStr == "" {
		t.Error("JSON não deveria ser vazio")
	}

	var boletoDeserializado Boleto
	if err := json.Unmarshal(jsonData, &boletoDeserializado); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if boletoDeserializado.DataPagamento != nil {
		t.Error("DataPagamento deveria ser nil para boleto não pago")
	}
}

func TestBoletoJSONCompleto(t *testing.T) {
	jsonInput := `{
		"id": "BOL003",
		"numero": "99999.88888.77777.66666.55555.44444.33333.22222.11111.00000.12345",
		"valor": 1234.56,
		"nome_sacado": "PEDRO ALMEIDA",
		"cpf_sacado": "11122233344",
		"data_vencimento": "2026-06-01T00:00:00Z",
		"data_pagamento": "2026-05-28T10:30:00Z",
		"valor_pago": 1234.56,
		"situacao": "PAGO",
		"nome_beneficiario": "Banco do Brasil",
		"pix_copia_e_cola": "00020126580014br.gov.bcb.pix0136abcdef1234567890abcdef123456789052040000530398654061234.565802BR5913PEDRO ALMEIDA6009SAO PAULO62070503***63041A2B"
	}`

	var boleto Boleto
	if err := json.Unmarshal([]byte(jsonInput), &boleto); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if boleto.ID != "BOL003" {
		t.Errorf("ID = %v, want BOL003", boleto.ID)
	}
	if boleto.Valor != 1234.56 {
		t.Errorf("Valor = %v, want 1234.56", boleto.Valor)
	}
	if boleto.Situacao != "PAGO" {
		t.Errorf("Situacao = %v, want PAGO", boleto.Situacao)
	}
	if boleto.NomeBeneficiario != "Banco do Brasil" {
		t.Errorf("NomeBeneficiario = %v, want Banco do Brasil", boleto.NomeBeneficiario)
	}
}
