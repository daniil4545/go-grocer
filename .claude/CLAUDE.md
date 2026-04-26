# grocer

Telegram-бот для учёта продуктов и финансов через парсинг кассовых чеков (QR-код → ФНС API → нормализация LLM → БД → аналитика).

## Стек

- Language: go
- Runtime: Go 1.22+
- Telegram: `go-telegram-bot-api` или `telebot v3`
- QR декодирование: `gozbar` (CGO), требует `libzbar0`
- БД: SQLite (`modernc.org/sqlite`, pure Go, без CGO)
- Query builder: `sqlx`
- LLM нормализация: Claude API (claude-haiku), через `net/http`
- Деплой: Docker Compose (VPS или RPi 4)

## Ключевые файлы

- `cmd/bot/main.go`: точка входа, инициализация зависимостей
- `internal/telegram/`: обработчики команд и фото от пользователя
- `internal/fns/`: клиент к ФНС API (`proverkacheka.nalog.ru`)
- `internal/qr/`: декодирование QR-кода из фото (gozbar)
- `internal/llm/`: пайплайн нормализации названий товаров через Claude
- `internal/db/`: схема, миграции, репозитории (sqlx)
- `docker-compose.yml`: описание окружения

## Архитектура

Основной поток: фото QR → декодирование параметров (`t`, `s`, `fn`, `i`, `fp`) → запрос к ФНС API → сохранение сырых позиций → LLM нормализация → `product_mappings` кэш → аналитика.

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

- ФНС API (`proverkacheka.nalog.ru`) требует регистрации/авторизации — изучить перед стартом M1
- `gozbar` требует CGO и системную зависимость `libzbar0` — учесть в Dockerfile
- Нормализация LLM — ключевой нетривиальный шаг; без неё аналитика даёт мусор
- `modernc.org/sqlite` — pure Go, не путать с `mattn/go-sqlite3` (CGO)
