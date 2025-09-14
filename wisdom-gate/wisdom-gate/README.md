## Test Task for Server Engineer

Design and implement "Word of Wisdom" TCP server.

• TCP server should be protected from DDOS attacks with the Proof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.
• The choice of the POW algorithm should be explained.
• After Proof Of Work verification, server should send one of the quotes from "word of wisdom" book or any other collection of the quotes.
• Docker file should be provided both for the server and for the client that solves the POW challenge

Environment variables:

```bash
# Server
SERVER_PORT=8080
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
IDLE_TIMEOUT=60s
MAX_CONNECTIONS=100

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
CHALLENGE_TTL=20s
SPENT_TTL=2m

# Proof of Work
POW_DIFFICULTY=20

# Quotes
QUOTES_SOURCE=internal
```

## Инструкции по запуску

### Docker Compose

```bash
cd wisdom-gate

docker-compose up --build

docker-compose run client
```

**Остановка сервисов:**
```bash
docker-compose down
```

```bash
# Запускаем Redis
docker run -d -p 6379:6379 --name redis-test redis:7-alpine

# Запускаем PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_DB=wisdom_gate \
  -e POSTGRES_USER=wisdom_user \
  -e POSTGRES_PASSWORD=wisdom_pass \
  --name postgres-test postgres:15-alpine

# Ждем пока PostgreSQL запустится (10-15 секунд)
sleep 15
```

**Сборка сервера**
```bash
cd wisdom-gate
go build -o wisdom-gate ./cmd/wisdom-gate.go
```

**Сборка клиента**
```bash
cd ../client
go build -o client ./main.go
```

**Запуск сервера**
```bash
cd ../wisdom-gate

export DBSTRING="postgres://wisdom_user:wisdom_pass@localhost:5432/wisdom_gate?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export POW_DIFFICULTY=4
export MIGRATION_PATH="./migrations"

./wisdom-gate
```

**Запуск клиента в новом терминале**
```bash
cd /path/to/wisdom-gate/client
./client 127.0.0.1:8080
```

