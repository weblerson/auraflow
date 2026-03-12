package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"auraflow/model"
)

var _ FinanceRepository = (*MockFinanceRepository)(nil)

type MockFinanceRepository struct {
	boletos []model.Boleto
}

func NewMockFinanceRepository(filePath string) (*MockFinanceRepository, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mock file: %w", err)
	}

	var mockData struct {
		Boletos []model.Boleto `json:"boletos"`
	}

	if err := json.Unmarshal(data, &mockData); err != nil {
		return nil, fmt.Errorf("failed to parse mock file: %w", err)
	}

	return &MockFinanceRepository{
		boletos: mockData.Boletos,
	}, nil
}

func (r *MockFinanceRepository) ConsultarBoletos(cpf string) ([]model.Boleto, error) {
	var result []model.Boleto
	for _, b := range r.boletos {
		if b.CPFSacado == cpf {
			result = append(result, b)
		}
	}
	return result, nil
}
