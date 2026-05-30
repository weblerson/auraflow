package main

import (
	"log"
	"os"
	"strings"

	"auraflow/controller"
	"auraflow/model"
	"auraflow/repository"
	"auraflow/util"
	"auraflow/view"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var financeCtrl *controller.FinanceController

// init() Roda antes de tudo para configurar as variáveis e dependências
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: Arquivo .env não encontrado, usando variáveis de ambiente do sistema.")
	}

	util.InitRedis()

	// Inicializa o Repositório da Pluggy usando credenciais limpas do .env
	clientID := os.Getenv("PLUGGY_CLIENT_ID")
	clientSecret := os.Getenv("PLUGGY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("PLUGGY_CLIENT_ID e PLUGGY_CLIENT_SECRET são obrigatórios no .env")
	}

	pluggyRepo := repository.NewPluggyRepository(clientID, clientSecret)
	financeCtrl = controller.NewFinanceController(pluggyRepo)
}

// ------------------------------------------------------------------
// HANDLERS (Controladores de cada ação do bot)
// ------------------------------------------------------------------

func handleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	bot.Send(view.Greeting(chatID))
	bot.Send(view.CPFRequest(chatID))

	if err := model.SetWaitingForCPF(util.Rdb, chatID, true); err != nil {
		log.Printf("Error setting waiting state: %v", err)
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
		log.Printf("Error consulting boletos: %v", err)
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

	// Por enquanto a API retorna tudo junto, mas enviamos para o formatador específico de Empréstimos
	boletos, err := financeCtrl.ConsultarBoletos(cpf)
	if err != nil || len(boletos) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Nenhum contrato de empréstimo ativo encontrado."))
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

	// Como o extrato é Mock, não precisa chamar a API da Pluggy aqui
	bot.Send(view.FormatMockExtrato(chatID))
}

// ------------------------------------------------------------------
// FUNÇÃO PRINCIPAL (Loop do Bot)
// ------------------------------------------------------------------

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

	// Fica escutando as mensagens do usuário
	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := strings.TrimSpace(update.Message.Text)

		// Verifica se o bot está esperando o usuário digitar o CPF
		if model.IsWaitingForCPF(util.Rdb, chatID) && !strings.HasPrefix(text, "/") {
			if err := model.StoreCPF(util.Rdb, util.EncryptionKey, chatID, text); err != nil {
				log.Printf("Error storing CPF: %v", err)
				bot.Send(tgbotapi.NewMessage(chatID, "Erro ao salvar CPF. Tente novamente."))
			} else {
				model.SetWaitingForCPF(util.Rdb, chatID, false)
				bot.Send(view.CPFSuccess(chatID)) // Isso vai mostrar o novo menu de botões!
			}
			continue
		}

		// Roteador de Comandos (O que o usuário clicou/digitou?)
		switch text {
		case "/start":
			handleStart(bot, update.Message)

		case "📄 Boletos e Faturas", "Consultar faturas", "/consultar":
			handleConsultarBoletos(bot, update.Message)

		case "💸 Empréstimos":
			handleConsultarEmprestimos(bot, update.Message)

		case "📊 Entradas e Saídas (Extrato)":
			handleConsultarExtrato(bot, update.Message)

		default:
			// Se ele mandar um texto aleatório que não seja um comando, ignoramos ou pedimos pra usar o menu
			if !model.IsWaitingForCPF(util.Rdb, chatID) {
				bot.Send(tgbotapi.NewMessage(chatID, "Comando não reconhecido. Por favor, use as opções do menu abaixo 👇"))
			}
		}
	}
}
