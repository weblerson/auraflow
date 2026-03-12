package model

import "time"

// Boleto representa um boleto bancário
type Boleto struct {
	ID               string     `json:"id"`
	Numero           string     `json:"numero"`
	Valor            float64    `json:"valor"`
	NomeSacado       string     `json:"nome_sacado"`
	CPFSacado        string     `json:"cpf_sacado"`
	DataVencimento   time.Time  `json:"data_vencimento"`
	DataPagamento    *time.Time `json:"data_pagamento,omitempty"`
	ValorPago        float64    `json:"valor_pago,omitempty"`
	Situacao         string     `json:"situacao"` // "PAGO", "PENDENTE", "VENCIDO", "CANCELADO"
	NomeBeneficiario string     `json:"nome_beneficiario"`
}
