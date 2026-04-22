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

func FormatBoleto(chatID int64, boleto model.Boleto) tgbotapi.MessageConfig {
	situacaoEmoji := getSituacaoEmoji(boleto.Situacao)
	text := fmt.Sprintf("%s %s\n", situacaoEmoji, boleto.Situacao)
	text += fmt.Sprintf("Valor: R$ %.2f\n", boleto.Valor)
	text += fmt.Sprintf("Vencimento: %s\n", boleto.DataVencimento.Format("02/01/2006"))
	text += fmt.Sprintf("Beneficiário: %s\n", boleto.NomeBeneficiario)
	if boleto.Situacao == "PAGO" && boleto.DataPagamento != nil {
		text += fmt.Sprintf("Pago em: %s\n", boleto.DataPagamento.Format("02/01/2006"))
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Pagar", "pagar:"+boleto.ID),
			tgbotapi.NewInlineKeyboardButtonData("Ignorar", "ignorar:"+boleto.ID),
		),
	)
	return msg
}

func BoletoPago(chatID int64, boletoID string) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, fmt.Sprintf("Boleto %s pago.", boletoID))
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
