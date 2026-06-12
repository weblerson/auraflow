package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewMockFinanceRepository(t *testing.T) {
	mockJSON := `{
		"boletos": [
			{
				"id": "BOL001",
				"numero": "12345.67890.00000.000000.000001 1",
				"valor": 150.00,
				"nome_sacado": "JOAO SILVA",
				"cpf_sacado": "12345678900",
				"data_vencimento": "2026-03-15T00:00:00Z",
				"situacao": "PENDENTE",
				"nome_beneficiario": "TECHBANK S.A."
			},
			{
				"id": "BOL002",
				"numero": "12345.67890.00000.000000.000002 2",
				"valor": 275.50,
				"nome_sacado": "JOAO SILVA",
				"cpf_sacado": "12345678900",
				"data_vencimento": "2026-03-20T00:00:00Z",
				"situacao": "PENDENTE",
				"nome_beneficiario": "TECHBANK S.A."
			}
		]
	}`

	tmpDir := t.TempDir()
	mockFile := filepath.Join(tmpDir, "boletos.json")
	if err := os.WriteFile(mockFile, []byte(mockJSON), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo mock: %v", err)
	}

	repo, err := NewMockFinanceRepository(mockFile)
	if err != nil {
		t.Fatalf("NewMockFinanceRepository() error = %v", err)
	}

	if repo == nil {
		t.Fatal("NewMockFinanceRepository() retornou nil")
	}
}

func TestMockFinanceRepository_ConsultarBoletos(t *testing.T) {
	mockJSON := `{
		"boletos": [
			{
				"id": "BOL001",
				"numero": "12345.67890.00000.000000.000001 1",
				"valor": 150.00,
				"nome_sacado": "JOAO SILVA",
				"cpf_sacado": "12345678900",
				"data_vencimento": "2026-03-15T00:00:00Z",
				"situacao": "PENDENTE",
				"nome_beneficiario": "TECHBANK S.A."
			},
			{
				"id": "BOL002",
				"numero": "12345.67890.00000.000000.000002 2",
				"valor": 275.50,
				"nome_sacado": "JOAO SILVA",
				"cpf_sacado": "12345678900",
				"data_vencimento": "2026-03-20T00:00:00Z",
				"situacao": "PENDENTE",
				"nome_beneficiario": "TECHBANK S.A."
			}
		]
	}`

	tmpDir := t.TempDir()
	mockFile := filepath.Join(tmpDir, "boletos.json")
	if err := os.WriteFile(mockFile, []byte(mockJSON), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo mock: %v", err)
	}

	repo, err := NewMockFinanceRepository(mockFile)
	if err != nil {
		t.Fatalf("NewMockFinanceRepository() error = %v", err)
	}

	tests := []struct {
		name      string
		cpf       string
		expected  int
		expectErr bool
	}{
		{
			name:     "CPF com dois boletos",
			cpf:      "12345678900",
			expected: 2,
		},
		{
			name:     "CPF sem boletos",
			cpf:      "00000000000",
			expected: 0,
		},
		{
			name:     "CPF inexistente",
			cpf:      "11111111111",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boletos, err := repo.ConsultarBoletos(tt.cpf)
			if (err != nil) != tt.expectErr {
				t.Errorf("ConsultarBoletos() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if len(boletos) != tt.expected {
				t.Errorf("ConsultarBoletos() retornou %d boletos, esperado %d", len(boletos), tt.expected)
			}
		})
	}
}

func TestNewMockFinanceRepository_ArquivoNaoExiste(t *testing.T) {
	_, err := NewMockFinanceRepository("/caminho/inexistente/boletos.json")
	if err == nil {
		t.Error("NewMockFinanceRepository() deveria retornar erro para arquivo inexistente")
	}
}

func TestNewMockFinanceRepository_JSONInvalido(t *testing.T) {
	tmpDir := t.TempDir()
	mockFile := filepath.Join(tmpDir, "invalido.json")
	if err := os.WriteFile(mockFile, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo mock: %v", err)
	}

	_, err := NewMockFinanceRepository(mockFile)
	if err == nil {
		t.Error("NewMockFinanceRepository() deveria retornar erro para JSON inválido")
	}
}
