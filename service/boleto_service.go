package service

import (
	"context"
	"fmt"

	"auraflow/crypto"
	"auraflow/domain"
	"auraflow/repository"
)

type BoletoService struct {
	api domain.BoletoAPI
}

func NewBoletoService(api domain.BoletoAPI) *BoletoService {
	return &BoletoService{api: api}
}

func (s *BoletoService) ConsultarBoletos(chatID int64) ([]domain.Boleto, error) {
	cpf, err := repository.GetCPF(chatID, crypto.Decrypt)
	if err != nil {
		return nil, err
	}

	if cpf == "" {
		return nil, fmt.Errorf("CPF não encontrado para este usuário")
	}

	return s.api.Consultar(context.Background(), cpf)
}

func (s *BoletoService) SalvarCPF(chatID int64, cpf string) error {
	return repository.StoreCPF(chatID, cpf, crypto.Encrypt)
}
