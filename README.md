# go-grocer

Telegram-бот для учёта продуктов и финансов через парсинг кассовых чеков.

MVP: бот принимает фото QR-кода или документ с чеком, получает позиции через proverkacheka.com, сохраняет чек в SQLite и показывает базовую статистику прямо в Telegram.

## Как работает

1. Отправь боту фото QR-кода с чека (или скриншот чека из приложения доставки)
2. Бот распознаёт чек через [proverkacheka.com](https://proverkacheka.com)
3. Чек и позиции сохраняются в локальную SQLite-базу
4. Команды `/last` и `/stats` показывают последний чек и статистику трат

## Команды

| Команда | Что делает |
|---------|------------|
| `/start` | Показывает короткую инструкцию |
| `/last` | Показывает последний сохранённый чек |
| `/stats` | Показывает количество чеков, общую сумму и топ товаров |

## Запуск

```bash
cp .env.example .env
# заполни BOT_TOKEN, FNS_TOKEN и ALLOWED_CHAT_ID в .env
go run .
```

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `BOT_TOKEN` | Токен Telegram-бота от @BotFather |
| `FNS_TOKEN` | API-токен от proverkacheka.com |
| `ALLOWED_CHAT_ID` | Telegram chat_id единственного разрешённого пользователя |
| `DB_PATH` | Путь до SQLite-файла, по умолчанию `./grocer.db` |

## Проверка

```bash
go test ./...
go vet ./...
```

## Стек

- Go 1.22+, pure Go (без CGO)
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [modernc.org/sqlite](https://modernc.org/sqlite) — SQLite без CGO
- [proverkacheka.com API](https://proverkacheka.com) — распознавание QR и получение данных чека
