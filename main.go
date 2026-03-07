package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func handleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	greeting := tgbotapi.NewMessage(chatID, "Olá! Bem-vindo ao AuraFlow! 🤖\nEstou aqui para te ajudar com seus boletos.")
	if _, err := bot.Send(greeting); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	cpfRequest := tgbotapi.NewMessage(chatID, "Por favor, informe o seu CPF:")
	if _, err := bot.Send(cpfRequest); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	if err := setWaitingForCPF(chatID, true); err != nil {
		log.Printf("Error setting waiting state: %v", err)
	}
}

func handleConsultar(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Consultando boletos")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func main() {
	godotenv.Load()
	initRedis()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		if isWaitingForCPF(chatID) && update.Message.Command() == "" {
			if err := storeCPF(chatID, update.Message.Text); err != nil {
				log.Printf("Error storing CPF: %v", err)
			}
			setWaitingForCPF(chatID, false)

			msg := tgbotapi.NewMessage(chatID, "CPF registrado com sucesso!")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Consultar boletos"),
				),
			)
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
			continue
		}

		switch update.Message.Command() {
		case "start":
			handleStart(bot, update.Message)
		case "consultar":
			handleConsultar(bot, update.Message)
		}

		if update.Message.Text == "Consultar boletos" {
			handleConsultar(bot, update.Message)
		}
	}
}
