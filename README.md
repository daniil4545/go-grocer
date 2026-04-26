# go-grocer

Telegram-бот для учёта продуктов и финансов через парсинг кассовых чеков.

## Как работает

1. Отправь боту фото QR-кода с чека (или скриншот чека из приложения доставки)
2. Бот распознаёт чек через [proverkacheka.com](https://proverkacheka.com) и возвращает список товаров с ценами

## Запуск

```bash
cp .env.example .env
# заполни BOT_TOKEN и FNS_TOKEN в .env
go run ./tg-bot/
```

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `BOT_TOKEN` | Токен Telegram-бота от @BotFather |
| `FNS_TOKEN` | API-токен от proverkacheka.com |

## Стек

- Go 1.22+, pure Go (без CGO)
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [proverkacheka.com API](https://proverkacheka.com) — распознавание QR и получение данных чека
