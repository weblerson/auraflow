package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"auraflow/adapter"
	"auraflow/handler"
	"auraflow/repository"
	"auraflow/service"
)

func main() {
	godotenv.Load()

	if err := repository.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	mockAPI := adapter.NewMockAPI()
	boletoSvc := service.NewBoletoService(mockAPI)
	telegramHandler := handler.NewTelegramHandler(bot, boletoSvc)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		telegramHandler.HandleMessage(&update)
	}
}
