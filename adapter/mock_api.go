package adapter

import (
	"encoding/json"
	"os"

	"auraflow/domain"
)

type MockAPI struct{}

func NewMockAPI() *MockAPI {
	return &MockAPI{}
}

func (m *MockAPI) Consultar(cpf string) ([]domain.Boleto, error) {
	data, err := os.ReadFile("mock_boletos.json")
	if err != nil {
		return nil, err
	}

	var boletos []domain.Boleto
	if err := json.Unmarshal(data, &boletos); err != nil {
		return nil, err
	}

	return boletos, nil
}
