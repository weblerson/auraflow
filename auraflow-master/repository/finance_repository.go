package repository

import (
	"auraflow/model"
)

type FinanceRepository interface {
	ConsultarBoletos(cpf string) ([]model.Boleto, error)
	ConsultarEmprestimos(cpf string) ([]model.Boleto, error)
	ConsultarFaturas(cpf string) ([]model.Boleto, error)
}
