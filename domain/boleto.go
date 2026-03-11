package domain

type Boleto struct {
	LinhaDigitavel string
	Valor          float64
	Vencimento     string
	Descricao      string
	Status         string
}

type BoletoAPI interface {
	Consultar(cpf string) ([]Boleto, error)
}
