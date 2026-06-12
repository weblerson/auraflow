package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: Arquivo .env não encontrado, usando variáveis de ambiente do sistema.")
	}

	util.InitRedis()

	clientID := os.Getenv("PLUGGY_CLIENT_ID")
	clientSecret := os.Getenv("PLUGGY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("PLUGGY_CLIENT_ID e PLUGGY_CLIENT_SECRET são obrigatórios no .env")
	}

	pluggyRepo := repository.NewPluggyRepository(clientID, clientSecret)
	financeCtrl = controller.NewFinanceController(pluggyRepo)
}

func handleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	bot.Send(view.Greeting(chatID))
	bot.Send(view.CPFRequest(chatID))

	if err := model.SetWaitingForCPF(util.Rdb, chatID, true); err != nil {
		log.Printf("Erro ao definir estado de espera: %v", err)
	}
}

func handleConsultarBoletos(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	cpf, err := model.GetCPF(util.Rdb, util.EncryptionKey, chatID)
	if err != nil || cpf == "" {
		bot.Send(view.CPFNotFound(chatID))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "🔄 Conectando via Open Finance para buscar faturas... Aguarde."))

	boletos, err := financeCtrl.ConsultarBoletos(cpf)
	if err != nil {
		log.Printf("Erro ao consultar boletos: %v", err)
		bot.Send(view.ErrorConsultingBoletos(chatID))
		return
	}

	if len(boletos) == 0 {
		bot.Send(view.NoBoletosFound(chatID, cpf))
		return
	}

	bot.Send(view.FormatBoletos(chatID, cpf, boletos))
}

func handleConsultarEmprestimos(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	cpf, err := model.GetCPF(util.Rdb, util.EncryptionKey, chatID)
	if err != nil || cpf == "" {
		bot.Send(view.CPFNotFound(chatID))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "🔄 Buscando contratos de crédito e financiamentos..."))

	boletos, err := financeCtrl.ConsultarEmprestimos(cpf)
	if err != nil || len(boletos) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Nenhum contrato de empréstimo ativo encontrado.\n\nUse 💰 *Simular Empréstimo* para ver opções de parcelamento."))
		return
	}

	bot.Send(view.FormatLoans(chatID, boletos))
}

func handleConsultarExtrato(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	_, err := model.GetCPF(util.Rdb, util.EncryptionKey, chatID)
	if err != nil {
		bot.Send(view.CPFNotFound(chatID))
		return
	}

	bot.Send(view.FormatMockExtrato(chatID))
}

func handlePrioridadePagamento(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	cpf, err := model.GetCPF(util.Rdb, util.EncryptionKey, chatID)
	if err != nil || cpf == "" {
		bot.Send(view.CPFNotFound(chatID))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "🔄 Buscando todas as suas faturas e empréstimos para calcular prioridade..."))

	faturas, err := financeCtrl.ConsultarFaturas(cpf)
	if err != nil {
		log.Printf("Aviso ao consultar faturas: %v", err)
	}

	emprestimos, err := financeCtrl.ConsultarEmprestimos(cpf)
	if err != nil {
		log.Printf("Aviso ao consultar empréstimos: %v", err)
	}

	if len(faturas) == 0 && len(emprestimos) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Nenhuma fatura ou empréstimo encontrado para suas contas conectadas."))
		return
	}

	bot.Send(view.FormatPrioridadePagamento(chatID, faturas, emprestimos))
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

	log.Printf("✅ Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("🛑 Encerrando bot...")
		cancel()
		bot.StopReceivingUpdates()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot encerrado graciosamente.")
			return
		case update, ok := <-updates:
			if !ok {
				return
			}

			if update.Message == nil {
				continue
			}

			chatID := update.Message.Chat.ID
			text := strings.TrimSpace(update.Message.Text)

			if model.IsWaitingForLoanValue(util.Rdb, chatID) && !strings.HasPrefix(text, "/") {
				var valor float64
				_, err := fmt.Sscanf(text, "%f", &valor)
				if err != nil || valor <= 0 {
					bot.Send(tgbotapi.NewMessage(chatID, "❌ Valor inválido. Por favor, digite um número (ex: 5000):"))
					continue
				}

				model.SetWaitingForLoanValue(util.Rdb, chatID, false)
				bot.Send(view.LoanSimulation(chatID, valor))
				continue
			}

			if model.IsWaitingForCPF(util.Rdb, chatID) && !strings.HasPrefix(text, "/") {
				if err := util.ValidarCPF(text); err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "❌ CPF inválido: "+err.Error()+". Por favor, digite novamente."))
					continue
				}

				if err := model.StoreCPF(util.Rdb, util.EncryptionKey, chatID, text); err != nil {
					log.Printf("Erro ao salvar CPF: %v", err)
					bot.Send(tgbotapi.NewMessage(chatID, "Erro ao salvar CPF. Tente novamente."))
				} else {
					model.SetWaitingForCPF(util.Rdb, chatID, false)
					bot.Send(view.CPFSuccess(chatID))
				}
				continue
			}

			switch text {
			case "/start":
				handleStart(bot, update.Message)

			case "📄 Boletos e Faturas", "Consultar faturas", "/consultar":
				handleConsultarBoletos(bot, update.Message)

			case "💸 Empréstimos":
				handleConsultarEmprestimos(bot, update.Message)

		case "📊 Entradas e Saídas (Extrato)":
			handleConsultarExtrato(bot, update.Message)

		case "📋 Prioridade de Pagamento":
			handlePrioridadePagamento(bot, update.Message)

		case "💰 Simular Empréstimo":
			bot.Send(view.LoanValueRequest(chatID))
			if err := model.SetWaitingForLoanValue(util.Rdb, chatID, true); err != nil {
				log.Printf("Erro ao definir estado de espera: %v", err)
			}

		default:
				if !model.IsWaitingForCPF(util.Rdb, chatID) {
					bot.Send(tgbotapi.NewMessage(chatID, "Comando não reconhecido. Por favor, use as opções do menu abaixo 👇"))
				}
			}
		}
	}
}
