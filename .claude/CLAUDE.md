# grocer

Telegram-бот для учёта продуктов и финансов через парсинг кассовых чеков (QR-код → ФНС API → нормализация LLM → БД → аналитика).

## Стек

- Language: go
- Runtime: Go 1.22+
- Telegram: `go-telegram-bot-api` или `telebot v3`
- БД: SQLite (`modernc.org/sqlite`, pure Go, без CGO)
- Query builder: `sqlx`
- LLM нормализация: Claude API (claude-haiku), через `net/http`
- Деплой: Docker Compose (VPS или RPi 4)

## Ключевые файлы

- `main.go`: точка входа, bot loop, обработка фото и документов
- `fns.go`: клиент к proverkacheka.com API, структуры ответа, retry-логика
- `docker-compose.yml`: описание окружения

## Архитектура

Основной поток: фото QR → proverkacheka.com API (формат 4, qrfile) → сырые позиции JSON → сохранение → LLM нормализация → `product_mappings` кэш → аналитика.

QR распознаётся на стороне API — gozbar не нужен, CGO-зависимости отсутствуют.

### proverkacheka.com API

Endpoint: `POST https://proverkacheka.com/api/v1/check/get`

Используемый формат (4 — qrfile):
- Content-Type: `multipart/form-data`
- Поля: `token` (из env), `qrfile` (байты фото)

Коды ответа (`code`):
- `1` — успех, данные в `data.json`
- `2` — данные ещё не получены, нужен retry
- `4` — подождать перед retry
- `0`, `3`, `5` — ошибка

Ключевые поля `data.json`:
- `items[i].name`, `items[i].price`, `items[i].quantity`, `items[i].sum` — позиции
- `totalSum` — итого
- `user` — организация, `userInn` — ИНН, `retailPlaceAddress` — адрес, `ticketDate` — дата

Схема БД:
- `receipts` — чеки (магазин, ИНН, дата, сумма)
- `receipt_items` — сырые позиции из чека
- `products` — нормализованные товары (каноническое имя, категория)
- `product_mappings` — кэш нормализации (raw_name → product_id)

Авторизация — одиночный пользователь, chat_id из env.

## Майлстоуны

- M1: приём фото QR, декодирование, запрос ФНС, сохранение в БД
- M2: LLM нормализация, ревью через бота, кэш в `product_mappings`
- M3: аналитика (траты, динамика цен, топ товаров)
- M4: предсказание заканчивающихся товаров, список покупок

## Gotchas

- Все суммы в API (`items[i].sum`, `totalSum`, `cashTotalSum`) — в **копейках**, не в рублях
- `code=2` — данные ещё не готовы, нужен retry; `code=4` — ждать перед retry; не путать с ошибкой
- `first=1` — чек получен впервые, `first=0` — повторный запрос (кэш на стороне API)
- Нормализация LLM — ключевой нетривиальный шаг; без неё аналитика даёт мусор
- `modernc.org/sqlite` — pure Go, не путать с `mattn/go-sqlite3` (CGO)
