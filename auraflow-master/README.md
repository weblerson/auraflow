# AuraFlow

Bot de Telegram para consulta de boletos, faturas e empréstimos via Open Finance.

## Pré-requisitos

- **Go** 1.22+
- **Redis** rodando
- Token de bot **Telegram** (crie via [BotFather](https://t.me/BotFather))
- Conta no [Pluggy](https://pluggy.ai) (Open Finance)
- Conta no [Mercado Pago](https://mercadopago.com) (para gerar PIX)

## Configuração

### 1. Clone o repositório

```bash
git clone https://github.com/seu-usuario/auraflow.git
cd auraflow
```

### 2. Instale as dependências

```bash
go mod download
```

### 3. Configure as variáveis de ambiente

Copie o arquivo de exemplo e preencha seus dados:

```bash
cp .env.example .env
```

| Variável | Descrição | Obrigatória | Padrão |
|---|---|---|---|
| `TELEGRAM_BOT_TOKEN` | Token do bot Telegram | Sim | — |
| `REDIS_ADDR` | Endereço do servidor Redis | Não | `localhost:6379` |
| `REDIS_PASSWORD` | Senha de autenticação do Redis | Não | — |
| `ENCRYPTION_KEY` | Chave AES-256 (32 bytes, 64 caracteres hex) | Sim | — |
| `PLUGGY_CLIENT_ID` | Client ID da API Pluggy | Sim | — |
| `PLUGGY_CLIENT_SECRET` | Client Secret da API Pluggy | Sim | — |
| `MP_ACCESS_TOKEN` | Access Token do Mercado Pago | Sim | — |
| `PLUGGY_ITEM_ID` | Item ID do Open Finance na Pluggy | Sim | — |

#### Gerando a chave de criptografia

```bash
openssl rand -hex 32
```

### 4. Inicie o Redis

```bash
# Usando Docker
docker run -d --name redis -p 6379:6379 redis

# Ou usando systemd
sudo systemctl start redis
```

### 5. Execute o bot

```bash
go run .
```

Para compilar e executar um binário:

```bash
go build -o auraflow .
./auraflow
```

## Comandos do Bot

| Comando | Descrição |
|---|---|
| `/start` | Cumprimenta o usuário e solicita o CPF |
| `/consultar` | Consulta boletos e faturas |

Após cadastrar o CPF, o bot apresenta um menu com as opções:
- **📄 Boletos e Faturas** - Consulta faturas de cartão e boletos
- **💸 Empréstimos** - Consulta contratos de crédito ativos
- **📊 Entradas e Saídas (Extrato)** - Extrato mockado

## Segurança

- CPFs são criptografados com **AES-256-GCM** antes de serem armazenados no Redis
- CPFs armazenados expiram automaticamente após **24 horas**
- O estado de espera por CPF expira após **5 minutos**
- O arquivo `.env` está no `.gitignore` — nunca o committe em versionamento
- **Nunca** compartilhe suas credenciais de API

## Estrutura do Projeto

```
auraflow/
├── main.go                    # Ponto de entrada e loop do bot
├── controller/
│   └── finance_controller.go  # Lógica de negócio
├── repository/
│   ├── finance_repository.go  # Interface de repositório
│   ├── pluggy_repository.go   # Implementação via API Pluggy
│   └── mock_finance_repository.go # Mock para testes
├── model/
│   ├── boleto.go              # Estrutura de boleto
│   └── cpf.go                 # Operações de CPF no Redis
├── view/
│   └── messages.go            # Mensagens formatadas para Telegram
├── util/
│   ├── crypto.go              # Criptografia AES-256-GCM
│   ├── env.go                 # Utilitário de variáveis de ambiente
│   ├── pix.go                 # Geração de PIX via Mercado Pago
│   └── redis.go               # Conexão Redis
├── mock/
│   └── boletos.json           # Dados de teste
├── .env.example               # Template de variáveis de ambiente
├── go.mod                     # Dependências Go
└── docker-compose.yml         # Setup com Docker
```

## License

MIT
