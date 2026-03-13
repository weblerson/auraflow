package view

import (
	"fmt"

	"auraflow/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Greeting(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "Olá! Bem-vindo ao AuraFlow! 🤖\nEstou aqui para te ajudar com seus boletos.")
}

func CPFRequest(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "Por favor, informe o seu CPF:")
}

func CPFSuccess(chatID int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, "CPF registrado com sucesso!")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Consultar boletos"),
		),
	)
	return msg
}

func CPFNotFound(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "CPF não encontrado. Por favor, utilize /start para cadastrar seu CPF.")
}

func NoBoletosFound(chatID int64, cpf string) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "Nenhum boleto encontrado para o CPF: "+cpf)
}

func ErrorConsultingBoletos(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "Erro ao consultar boletos. Tente novamente mais tarde.")
}

func FormatBoletos(chatID int64, cpf string, boletos []model.Boleto) tgbotapi.MessageConfig {
	var text string

	if len(boletos) == 1 {
		text = fmt.Sprintf("📄 Boleto encontrado para CPF %s:\n\n", cpf)
	} else {
		text = fmt.Sprintf("📄 Boletos encontrados para CPF %s (%d):\n\n", cpf, len(boletos))
	}

	for i, b := range boletos {
		situacaoEmoji := getSituacaoEmoji(b.Situacao)
		text += fmt.Sprintf("%d. %s %s\n", i+1, situacaoEmoji, b.Situacao)
		text += fmt.Sprintf("   Valor: R$ %.2f\n", b.Valor)
		text += fmt.Sprintf("   Vencimento: %s\n", b.DataVencimento.Format("02/01/2006"))
		text += fmt.Sprintf("   Beneficiário: %s\n", b.NomeBeneficiario)
		if b.Situacao == "PAGO" && b.DataPagamento != nil {
			text += fmt.Sprintf("   Pago em: %s\n", b.DataPagamento.Format("02/01/2006"))
		}
		text += "\n"
	}

	return tgbotapi.NewMessage(chatID, text)
}

func getSituacaoEmoji(situacao string) string {
	switch situacao {
	case "PAGO":
		return "✅"
	case "PENDENTE":
		return "⏳"
	case "VENCIDO":
		return "⚠️"
	case "CANCELADO":
		return "❌"
	default:
		return "❓"
	}
}
