package model

import "time"

// Boleto/Fatura representa uma cobrança que o usuário precisa pagar
type Boleto struct {
	ID               string     `json:"id"`
	Numero           string     `json:"numero"` // Pode ser o ID da fatura ou linha digitável
	Valor            float64    `json:"valor"`
	NomeSacado       string     `json:"nome_sacado"`
	CPFSacado        string     `json:"cpf_sacado"`
	DataVencimento   time.Time  `json:"data_vencimento"`
	DataPagamento    *time.Time `json:"data_pagamento,omitempty"`
	ValorPago        float64    `json:"valor_pago,omitempty"`
	Situacao         string     `json:"situacao"`          // "PAGO", "PENDENTE", "VENCIDO"
	NomeBeneficiario string     `json:"nome_beneficiario"` // Ex: "Nubank", "Itaú"
	PixCopiaECola    string     `json:"pix_copia_e_cola"`  // Novo campo para o Pix
}
