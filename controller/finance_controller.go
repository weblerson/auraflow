package controller

import (
	"auraflow/model"
	"auraflow/repository"
)

type FinanceController struct {
	repo repository.FinanceRepository
}

func NewFinanceController(repo repository.FinanceRepository) *FinanceController {
	return &FinanceController{
		repo: repo,
	}
}

func (c *FinanceController) ConsultarBoletos(cpf string) ([]model.Boleto, error) {
	return c.repo.ConsultarBoletos(cpf)
}
