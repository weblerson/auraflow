package main

import (
	"log"
	"os"

	"auraflow/controller"
	"auraflow/model"
	"auraflow/repository"
	"auraflow/util"
	"auraflow/view"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var financeCtrl *controller.FinanceController

func init() {
	godotenv.Load()
	util.InitRedis()

	repo, err := repository.NewMockFinanceRepository("mock/boletos.json")
	if err != nil {
		log.Printf("Warning: failed to load mock repository: %v", err)
	} else {
		financeCtrl = controller.NewFinanceController(repo)
	}
}

func handleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	if _, err := bot.Send(view.Greeting(chatID)); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	if _, err := bot.Send(view.CPFRequest(chatID)); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	if err := model.SetWaitingForCPF(util.Rdb, chatID, true); err != nil {
		log.Printf("Error setting waiting state: %v", err)
	}
}

func handleConsultar(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	cpf, err := model.GetCPF(util.Rdb, util.EncryptionKey, message.Chat.ID)
	if err != nil {
		log.Printf("Error getting CPF: %v", err)
		cpf = ""
	}

	if cpf == "" {
		if _, err := bot.Send(view.CPFNotFound(message.Chat.ID)); err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}

	boletos, err := financeCtrl.ConsultarBoletos(cpf)
	if err != nil {
		log.Printf("Error consulting boletos: %v", err)
		if _, err := bot.Send(view.ErrorConsultingBoletos(message.Chat.ID)); err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}

	if len(boletos) == 0 {
		if _, err := bot.Send(view.NoBoletosFound(message.Chat.ID, cpf)); err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}

	if _, err := bot.Send(view.FormatBoletos(message.Chat.ID, cpf, boletos)); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func main() {
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

		if model.IsWaitingForCPF(util.Rdb, chatID) && update.Message.Command() == "" {
			if err := model.StoreCPF(util.Rdb, util.EncryptionKey, chatID, update.Message.Text); err != nil {
				log.Printf("Error storing CPF: %v", err)
			}
			model.SetWaitingForCPF(util.Rdb, chatID, false)

			if _, err := bot.Send(view.CPFSuccess(chatID)); err != nil {
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
