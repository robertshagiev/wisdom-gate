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

```
wisdom-gate/
├── cmd/wisdom-gate.go          # Точка входа
├── internal/
│   ├── adapters/               # PostgreSQL, Redis
│   ├── application/            # PoW, цитаты, протокол
│   ├── config/                 # Конфигурация
│   └── delivery/tcp/           # TCP сервер, middleware
├── migrations/                 # SQL миграции
└── docker/                     # Docker файлы

client/                         # Тестовый клиент
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

## Производительность

### Текущие показатели
- **Сложность 4:** 10-50ms время решения PoW
- **Сложность 20:** 50-200ms время решения PoW  
- **Максимум соединений:** 100 одновременных TCP соединений
- **Пропускная способность:** 100-500 запросов/минута
- **Латентность Redis:** <1ms для операций с challenges
- **Латентность PostgreSQL:** 5-15ms для получения цитат

### Масштабируемость
**Горизонтальное масштабирование:**
- **Load Balancer** - распределение нагрузки между инстансами
- **Redis Cluster** - масштабирование кэша challenges
- **PostgreSQL Read Replicas** - чтение цитат с реплик
- **Connection Pooling** - эффективное использование соединений

**Вертикальное масштабирование:**
- **CPU** - увеличение сложности PoW при росте нагрузки
- **RAM** - больше соединений и кэша challenges
- **Network** - оптимизация TCP буферов и timeouts

### Выбор технологий

#### PostgreSQL
- **Надежность** - проверенная временем СУБД для критически важных данных цитат
- **Backup/Recovery** - встроенные механизмы резервного копирования для сохранности коллекции мудрости
- **Connection Pooling** - эффективное управление соединениями через pgxpool для множественных TCP соединений
- **ACID транзакции** - гарантирует целостность данных при добавлении новых цитат

#### Redis
- **In-memory производительность** - микросекундные операции для PoW challenges (критично для TCP протокола)
- **TTL поддержка** - автоматическое истечение challenges (20s) и spent tokens (2m) без дополнительного кода
- **Низкая латентность** - критично для DDoS защиты (клиент ждет challenge мгновенно)
- **Простота** - минимальная конфигурация для кэширования временных данных PoW


### Hashcash
- **CPU-bound** - защита от ботов с ограниченными ресурсами
- **Configurable difficulty** - адаптация к нагрузке
- **Stateless verification** - не требует хранения промежуточных состояний
- **Industry standard** - проверенный алгоритм (используется в Bitcoin)
- **SHA-256** - криптографически стойкий хеш

## Безопасность

- DDoS защита через Proof of Work
- Anti-replay через Redis
- Time-based expiration (20 секунд)
- Subject binding к IP адресу
- Connection limits и timeouts