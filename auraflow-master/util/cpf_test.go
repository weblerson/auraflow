package util

import (
	"testing"
)

func TestValidarCPF(t *testing.T) {
	tests := []struct {
		name    string
		cpf     string
		wantErr bool
	}{
		{
			name:    "CPF válido com pontuação",
			cpf:     "529.982.247-25",
			wantErr: false,
		},
		{
			name:    "CPF válido sem pontuação",
			cpf:     "52998224725",
			wantErr: false,
		},
		{
			name:    "CPF válido com espaços",
			cpf:     "529 982 247 25",
			wantErr: false,
		},
		{
			name:    "CPF inválido - todos dígitos iguais",
			cpf:     "111.111.111-11",
			wantErr: true,
		},
		{
			name:    "CPF inválido - dígito verificador errado",
			cpf:     "529.982.247-26",
			wantErr: true,
		},
		{
			name:    "CPF inválido - muito curto",
			cpf:     "1234567890",
			wantErr: true,
		},
		{
			name:    "CPF inválido - muito longo",
			cpf:     "123456789012",
			wantErr: true,
		},
		{
			name:    "CPF inválido - com letras",
			cpf:     "abc.def.ghi-jk",
			wantErr: true,
		},
		{
			name:    "CPF vazio",
			cpf:     "",
			wantErr: true,
		},
		{
			name:    "CPF zero zero zero",
			cpf:     "000.000.000-00",
			wantErr: true,
		},
		{
			name:    "CPF válido 2",
			cpf:     "529.982.247-25",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidarCPF(tt.cpf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidarCPF() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatarCPF(t *testing.T) {
	tests := []struct {
		name string
		cpf  string
		want string
	}{
		{
			name: "CPF sem pontuação",
			cpf:  "52998224725",
			want: "529.982.247-25",
		},
		{
			name: "CPF já formatado",
			cpf:  "529.982.247-25",
			want: "529.982.247-25",
		},
		{
			name: "CPF com espaços",
			cpf:  "529 982 247 25",
			want: "529.982.247-25",
		},
		{
			name: "CPF curto (inválido)",
			cpf:  "12345",
			want: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatarCPF(tt.cpf); got != tt.want {
				t.Errorf("FormatarCPF() = %v, want %v", got, tt.want)
			}
		})
	}
}
