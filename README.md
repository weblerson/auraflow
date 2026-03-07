# AuraFlow

Telegram bot for consulting boletos (payment slips).

## Prerequisites

- **Go** 1.25+
- **Redis** server running
- A **Telegram Bot Token** (create one via [BotFather](https://t.me/BotFather))

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/your-user/auraflow.git
cd auraflow
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Configure environment variables

Create a `.env` file in the project root:

```env
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
ENCRYPTION_KEY=your-64-hex-char-key
```

| Variable | Description | Required | Default |
|---|---|---|---|
| `TELEGRAM_BOT_TOKEN` | Telegram bot API token | Yes | — |
| `REDIS_ADDR` | Redis server address | No | `localhost:6379` |
| `REDIS_PASSWORD` | Redis authentication password | No | — |
| `ENCRYPTION_KEY` | 32-byte AES-256 key, hex-encoded (64 characters) | Yes | — |

#### Generating an encryption key

```bash
openssl rand -hex 32
```

### 4. Start Redis

```bash
# Using Docker
docker run -d --name redis -p 6379:6379 redis

# Or using systemd
sudo systemctl start redis
```

### 5. Run the bot

```bash
go run .
```

To build and run a binary:

```bash
go build -o auraflow .
./auraflow
```

## Bot commands

| Command | Description |
|---|---|
| `/start` | Greets the user and asks for their CPF |
| `/consultar` | Consults boletos |

The "Consultar boletos" button is also available in the chat keyboard after registering a CPF.

## Security

- CPF data is encrypted with **AES-256-GCM** before being stored in Redis.
- Stored CPFs automatically expire after **24 hours**.
- The CPF input waiting state expires after **5 minutes**.
- The `.env` file is included in `.gitignore` — never commit it to version control.
