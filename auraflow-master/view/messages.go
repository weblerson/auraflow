package view

import (
	"fmt"
	"time"

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

	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📄 Boletos e Faturas"),
			tgbotapi.NewKeyboardButton("💸 Empréstimos"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Entradas e Saídas (Extrato)"),
			tgbotapi.NewKeyboardButton("📋 Prioridade de Pagamento"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💰 Simular Empréstimo"),
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

func FormatFaturas(chatID int64, boletos []model.Boleto) tgbotapi.MessageConfig {
	var text string
	text = fmt.Sprintf("📄 *Faturas de Cartão de Crédito (%d):*\n\n", len(boletos))

	for i, b := range boletos {
		situacaoEmoji := getSituacaoEmoji(b.Situacao)
		diasRestantes := int(time.Until(b.DataVencimento).Hours() / 24)
		prioridade := getPrioridade(diasRestantes, b.Situacao)

		text += fmt.Sprintf("*%d.* %s %s\n", i+1, situacaoEmoji, b.NomeBeneficiario)
		text += fmt.Sprintf("💰 *Valor:* R$ %.2f\n", b.Valor)
		text += fmt.Sprintf("📅 *Vencimento:* %s", b.DataVencimento.Format("02/01/2006"))
		if diasRestantes >= 0 {
			text += fmt.Sprintf(" (%d dias)", diasRestantes)
		} else {
			text += fmt.Sprintf(" (VENCIDO há %d dias)", -diasRestantes)
		}
		text += "\n"
		text += fmt.Sprintf("⚡ *Prioridade:* %s\n", prioridade)
		text += fmt.Sprintf("💠 *PIX:*\n`%s`\n", b.PixCopiaECola)
		text += "-----------------------\n"
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	return msg
}

func getPrioridade(diasRestantes int, situacao string) string {
	if situacao == "VENCIDO" || diasRestantes < 0 {
		return "🔴 URGENTE - VENCIDO"
	}
	if situacao == "PAGO" {
		return "🟢 PAGO"
	}
	if diasRestantes <= 3 {
		return "🔴 CRÍTICO - Vence em até 3 dias"
	}
	if diasRestantes <= 7 {
		return "🟡 ALTO - Vence em até 7 dias"
	}
	if diasRestantes <= 15 {
		return "🟠 MÉDIO - Vence em até 15 dias"
	}
	return "🟢 BAIXO - Vence em mais de 15 dias"
}

func FormatPrioridadePagamento(chatID int64, faturas []model.Boleto, emprestimos []model.Boleto) tgbotapi.MessageConfig {
	text := "📋 *Prioridade de Pagamento*\n\n"
	text += "📊 *O que pagar primeiro (ordenado por urgência):*\n\n"

	type itemPagamento struct {
		nome       string
		valor      float64
		vencimento time.Time
		situacao   string
		pix        string
		tipo       string
	}

	var todos []itemPagamento

	for _, f := range faturas {
		todos = append(todos, itemPagamento{
			nome:       f.NomeBeneficiario,
			valor:      f.Valor,
			vencimento: f.DataVencimento,
			situacao:   f.Situacao,
			pix:        f.PixCopiaECola,
			tipo:       "FATURA",
		})
	}

	for _, e := range emprestimos {
		todos = append(todos, itemPagamento{
			nome:       e.NomeBeneficiario,
			valor:      e.Valor,
			vencimento: e.DataVencimento,
			situacao:   e.Situacao,
			pix:        e.PixCopiaECola,
			tipo:       "EMPRÉSTIMO",
		})
	}

	for i, item := range todos {
		diasRestantes := int(time.Until(item.vencimento).Hours() / 24)
		prioridade := getPrioridade(diasRestantes, item.situacao)
		situacaoEmoji := getSituacaoEmoji(item.situacao)

		text += fmt.Sprintf("*%d. %s %s*\n", i+1, situacaoEmoji, item.nome)
		text += fmt.Sprintf("📌 *Tipo:* %s\n", item.tipo)
		text += fmt.Sprintf("💰 *Valor:* R$ %.2f\n", item.valor)
		text += fmt.Sprintf("📅 *Vence:* %s", item.vencimento.Format("02/01/2006"))
		if diasRestantes >= 0 {
			text += fmt.Sprintf(" (%d dias)", diasRestantes)
		} else {
			text += fmt.Sprintf(" (VENCIDO)")
		}
		text += "\n"
		text += fmt.Sprintf("⚡ *Prioridade:* %s\n", prioridade)
		text += fmt.Sprintf("💠 *PIX:* `%s`\n", item.pix)
		text += "-----------------------\n"
	}

	text += "\n💡 *Dica:* Pague primeiro os itens vermelhos (URGENTE/CRÍTICO)!"

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

func LoanValueRequest(chatID int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, "💰 *Simular Empréstimo*\n\nDigite o valor que deseja simular (ex: 5000):")
}

func LoanSimulation(chatID int64, valor float64) tgbotapi.MessageConfig {
	text := fmt.Sprintf("💰 *Simulação de Empréstimo*\n\n")
	text += fmt.Sprintf("Valor solicitado: *R$ %.2f*\n\n", valor)
	text += "📊 *Opções de parcelamento:*\n\n"

	parcelas := []int{6, 12, 24, 48, 60}
	for _, p := range parcelas {
		taxa := 0.0199
		valorComJuros := valor * (1 + taxa*float64(p)/12)
		parcelaComJuros := valorComJuros / float64(p)
		text += fmt.Sprintf("📋 *%dx* de R$ %.2f (taxa ~1,99%% a.m.)\n", p, parcelaComJuros)
		text += fmt.Sprintf("   Total: R$ %.2f\n\n", valorComJuros)
	}

	text += "💡 *Dica:* Empréstimos com menos parcelas têm juros menores.\n"
	text += "Para contratar, acesse seu banco pelo Open Finance."

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
