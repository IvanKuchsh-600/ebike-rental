# Ebike Rental

Сервис учёта аренды электровелосипедов: велосипеды, аренды, поломки, выплаты партнёру.

## Стек

- **Backend:** Go 1.25, gin, pgx v5, JWT, bcrypt
- **БД:** PostgreSQL 16
- **Frontend:** Go templates, HTMX, Pico.css
- **Миграции:** golang-migrate
- **Инфраструктура:** Docker, docker-compose

## Быстрый старт (локально)

### 1. Клонировать репозиторий

```bash
git clone https://github.com/IvanKuchsh-600/ebike-rental.git
cd ebike-rental
```

### 2. Настроить окружение

Скопировать `.env.example` в `.env` и заполнить:

```bash
cp .env.example .env
```

Сгенерировать `JWT_SECRET`:

```bash
openssl rand -base64 48
```

### 3. Поднять Postgres

```bash
docker compose up -d
```

### 4. Накатить миграции

```bash
docker compose up migrate
```

### 5. Создать пользователей

```bash
go run ./cmd/seed \
  -owner-email=owner@example.com \
  -owner-password=strongPassword123 \
  -partner-email=partner@example.com \
  -partner-password=strongPassword456
```

### 6. Запустить API

```bash
go run ./cmd/api
```

Открыть: http://localhost:8080/web/login

## Роли

- **owner** — полный доступ: CRUD велосипедов, аренд, поломок, выплат.
- **partner** — только просмотр: дашборд, велосипеды, аренды, поломки, выплаты.

## Основные ручки

### Публичные

- `POST /auth/login` — вход, выдаёт JWT
- `GET /health` — healthcheck

### Только для owner

- `GET|POST|PUT|DELETE /bikes` — велосипеды
- `POST|DELETE /bikes/{id}/photo` — фото велосипеда
- `GET|POST|PUT|DELETE /rentals` — аренды
- `POST /rentals/{id}/return` — вернуть велосипед
- `POST /rentals/{id}/pay` — отметить оплату
- `GET|POST|PUT|DELETE /breakdowns` — поломки
- `POST /breakdowns/{id}/fix` — отметить починенным
- `GET|POST|DELETE /payouts` — выплаты
- `GET /payouts/preview?month=YYYY-MM` — предварительный расчёт

### Только для partner

- `GET /partner/dashboard` — сводка
- `GET /partner/bikes` — велосипеды
- `GET /partner/rentals` — аренды
- `GET /partner/breakdowns` — поломки
- `GET /partner/payouts` — выплаты
- `GET /partner/payouts/preview?month=YYYY-MM` — preview

## Бизнес-логика

- **Партнёр получает 10%** от суммы оплаченных аренд (настраивается в `.env`).
- **Процент считается по `started_at`** — аренда за апрель попадает в апрельский payout.
- **Можно удалить:** велосипед без аренд, активную неоплаченную аренду, непочиненную поломку, невыплаченную выплату.
- **Нельзя удалить:** велосипед с арендами, оплаченную/возвращённую аренду, починенную поломку, выплаченную выплату.
- **Залог** — возвратная сумма, не входит в базу для процента партнёра.

## Структура

```
cmd/api/          — точка входа API
cmd/seed/         — создание пользователей
internal/         — весь внутренний код
  config/         — конфиг из .env
  domain/         — доменные модели и ошибки
  repository/     — доступ к данным
  storage/        — подключение к Postgres
  usecase/        — бизнес-логика
  http/           — хендлеры, роутер, middleware
migrations/       — SQL-миграции
web/              — шаблоны и статика
```

## Разработка

- Тесты: `go test ./...`
- Миграции: `migrate -path migrations -database "$DATABASE_URL" up`
- Логи: stdout в JSON