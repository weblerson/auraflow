package handler

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"auraflow/repository"
	"auraflow/service"
)

type TelegramHandler struct {
	bot           *tgbotapi.BotAPI
	boletoService *service.BoletoService
}

func NewTelegramHandler(bot *tgbotapi.BotAPI, svc *service.BoletoService) *TelegramHandler {
	return &TelegramHandler{
		bot:           bot,
		boletoService: svc,
	}
}

func (h *TelegramHandler) HandleStart(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	greeting := tgbotapi.NewMessage(chatID, "Olá! Bem-vindo ao AuraFlow! 🤖\nEstou aqui para te ajudar com seus boletos.")
	if _, err := h.bot.Send(greeting); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	cpfRequest := tgbotapi.NewMessage(chatID, "Por favor, informe o seu CPF:")
	if _, err := h.bot.Send(cpfRequest); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	if err := repository.SetWaitingForCPF(chatID, true); err != nil {
		log.Printf("Error setting waiting state: %v", err)
	}
}

func (h *TelegramHandler) HandleConsultar(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	boletos, err := h.boletoService.ConsultarBoletos(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Erro ao consultar boletos: %v", err))
		h.bot.Send(msg)
		return
	}

	if len(boletos) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Nenhum boleto encontrado para o seu CPF.")
		h.bot.Send(msg)
		return
	}

	for _, b := range boletos {
		text := fmt.Sprintf("*%s*\nValor: R$ %.2f\nVencimento: %s\nStatus: %s",
			b.Descricao, b.Valor, b.Vencimento, b.Status)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		h.bot.Send(msg)
	}
}

func (h *TelegramHandler) HandleCPF(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	cpf := message.Text

	if err := h.boletoService.SalvarCPF(chatID, cpf); err != nil {
		log.Printf("Error storing CPF: %v", err)
		return
	}

	repository.SetWaitingForCPF(chatID, false)

	msg := tgbotapi.NewMessage(chatID, "CPF registrado com sucesso!")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Consultar boletos"),
		),
	)
	h.bot.Send(msg)
}

func (h *TelegramHandler) HandleMessage(update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	if repository.IsWaitingForCPF(chatID) && update.Message.Command() == "" {
		h.HandleCPF(update.Message)
		return
	}

	switch update.Message.Command() {
	case "start":
		h.HandleStart(update.Message)
	case "consultar":
		h.HandleConsultar(update.Message)
	}

	if update.Message.Text == "Consultar boletos" {
		h.HandleConsultar(update.Message)
	}
}
