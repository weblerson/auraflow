package view

import (
	"fmt"

	"auraflow/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Greeting(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "Olá! Bem-vindo ao AuraFlow! 🤖\nEstou aqui para centralizar e pagar suas faturas.")
}

func CPFRequest(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "Por favor, informe o seu CPF:")
}

func CPFSuccess(chatID int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, "CPF registrado com sucesso! 🎉\n\nO que você deseja consultar no seu Open Finance?")
	msg.ParseMode = "Markdown"

	// Menu customizado com botões separados
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📄 Boletos e Faturas"),
			tgbotapi.NewKeyboardButton("💸 Empréstimos"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Entradas e Saídas (Extrato)"),
		),
	)
	return msg
}

func CPFNotFound(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "CPF não encontrado. Por favor, utilize /start para cadastrar seu CPF.")
}

func NoBoletosFound(chatID int64, cpf string) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "Nenhuma fatura pendente ou contrato encontrado para suas contas conectadas.")
}

func ErrorConsultingBoletos(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "Erro ao conectar com as instituições financeiras. Tente novamente mais tarde.")
}

func getSituacaoEmoji(situacao string) string {
	switch situacao {
	case "PAGO":
		return "✅"
	case "VENCIDO":
		return "❌"
	default:
		return "⏳"
	}
}

func FormatBoletos(chatID int64, cpf string, boletos []model.Boleto) tgbotapi.MessageConfig {
	var text string
	text = fmt.Sprintf("📄 *Faturas e Boletos encontrados (%d):*\n\n", len(boletos))

	for i, b := range boletos {
		situacaoEmoji := getSituacaoEmoji(b.Situacao)
		text += fmt.Sprintf("*%d.* %s %s\n", i+1, situacaoEmoji, b.Situacao)
		text += fmt.Sprintf("🏦 *Instituição:* %s\n", b.NomeBeneficiario)
		text += fmt.Sprintf("💰 *Valor:* R$ %.2f\n", b.Valor)
		text += fmt.Sprintf("📅 *Vencimento:* %s\n", b.DataVencimento.Format("02/01/2006"))
		text += fmt.Sprintf("💠 *PIX Copia e Cola para Pagamento:*\n`%s`\n", b.PixCopiaECola)
		text += "-----------------------\n"
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	return msg
}

func FormatLoans(chatID int64, boletos []model.Boleto) tgbotapi.MessageConfig {
	var text string
	text = fmt.Sprintf("💸 *Empréstimos e Financiamentos (%d):*\n\n", len(boletos))

	for i, b := range boletos {
		text += fmt.Sprintf("*%d. ⏳ CONTRATO ATIVO*\n", i+1)
		text += fmt.Sprintf("🏦 *Modalidade:* %s\n", b.NomeBeneficiario)
		text += fmt.Sprintf("💰 *Saldo Devedor Atual:* R$ %.2f\n", b.Valor)
		text += fmt.Sprintf("📅 *Próxima Parcela:* %s\n", b.DataVencimento.Format("02/01/2006"))
		text += fmt.Sprintf("💠 *Quitar / Amortizar via PIX:*\n`%s`\n", b.PixCopiaECola)
		text += "-----------------------\n"
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	return msg
}

func FormatMockExtrato(chatID int64) tgbotapi.MessageConfig {
	text := "📊 *Entradas e Saídas - Últimos 7 dias*\n\n"
	text += "🟢 *RECEBIDO (Entradas):*\n"
	text += "• 15/05 - PIX Recebido (Sérgio S.): +R$ 1.500,00\n"
	text += "• 12/05 - TED Recebida (AuraCorp): +R$ 3.450,20\n\n"
	text += "🔴 *PERDEU / GASTOU (Saídas):*\n"
	text += "• 16/05 - PIX Enviado (Mercado Livre): -R$ 149,90\n"
	text += "• 14/05 - Débito (Posto Shell): -R$ 80,00\n"
	text += "• 11/05 - Tarifa Bancária: -R$ 19,90\n\n"
	text += "💰 *Saldo Consolidado:* R$ 4.700,40"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	return msg
}
