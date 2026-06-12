package util

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidarCPF verifica se um CPF é válido according to Receita Federal rules
func ValidarCPF(cpf string) error {
	cpf = strings.ReplaceAll(cpf, ".", "")
	cpf = strings.ReplaceAll(cpf, "-", "")
	cpf = strings.ReplaceAll(cpf, " ", "")

	if len(cpf) != 11 {
		return fmt.Errorf("CPF deve ter 11 dígitos, recebido %d", len(cpf))
	}

	matched, _ := regexp.MatchString(`^\d{11}$`, cpf)
	if !matched {
		return fmt.Errorf("CPF deve conter apenas dígitos")
	}

	// Verificar se todos os dígitos são iguais (CPF inválido)
	allSame := true
	for i := 1; i < 11; i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return fmt.Errorf("CPF inválido: todos os dígitos são iguais")
	}

	// Validar primeiro dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(cpf[i]-'0') * (10 - i)
	}
	remainder := sum % 11
	digit1 := 0
	if remainder >= 2 {
		digit1 = 11 - remainder
	}
	if int(cpf[9]-'0') != digit1 {
		return fmt.Errorf("CPF inválido: primeiro dígito verificador incorreto")
	}

	// Validar segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(cpf[i]-'0') * (11 - i)
	}
	remainder = sum % 11
	digit2 := 0
	if remainder >= 2 {
		digit2 = 11 - remainder
	}
	if int(cpf[10]-'0') != digit2 {
		return fmt.Errorf("CPF inválido: segundo dígito verificador incorreto")
	}

	return nil
}

// FormatarCPF formata o CPF para exibição (XXX.XXX.XXX-XX)
func FormatarCPF(cpf string) string {
	cpf = strings.ReplaceAll(cpf, ".", "")
	cpf = strings.ReplaceAll(cpf, "-", "")
	cpf = strings.ReplaceAll(cpf, " ", "")

	if len(cpf) != 11 {
		return cpf
	}

	return fmt.Sprintf("%s.%s.%s-%s", cpf[:3], cpf[3:6], cpf[6:9], cpf[9:])
}
