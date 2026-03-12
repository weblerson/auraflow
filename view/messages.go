package view

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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

func ConsultarBoletos(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "Consultando boletos")
}
