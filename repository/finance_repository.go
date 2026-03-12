package repository

import (
	"auraflow/model"
)

type FinanceRepository interface {
	ConsultarBoletos(cpf string) ([]model.Boleto, error)
}
