# go-grocer

[![CI](https://github.com/daniil4545/go-grocer/actions/workflows/ci.yml/badge.svg)](https://github.com/daniil4545/go-grocer/actions/workflows/ci.yml)

Telegram-бот для учёта покупок по кассовым чекам: фото QR-кода, позиции из API ФНС, статистика трат в SQLite. Pure Go, без CGO.

Бот принимает фото QR-кода или документ с чеком, распознаёт чек через [proverkacheka.com](https://proverkacheka.com), сохраняет чек и позиции в локальную SQLite-базу и показывает статистику командами прямо в Telegram.

## Быстрый старт

```bash
git clone https://github.com/daniil4545/go-grocer.git
cd go-grocer
cp .env.example .env   # заполнить BOT_TOKEN, FNS_TOKEN, ALLOWED_CHAT_ID
go run .
```

Проверка:

```bash
go test ./...
go vet ./...
```

## Команды

| Команда | Что делает |
|---------|------------|
| `/start` | Короткая инструкция |
| `/last` | Последний сохранённый чек с позициями |
| `/stats` | Количество чеков, общая сумма, топ товаров |

## Инженерные решения

- **Pure Go без CGO.** SQLite через `modernc.org/sqlite`: сборка не требует C-toolchain, бинарник кросс-компилируется штатным `go build`.
- **Allowlist доступа.** Бот обслуживает единственный `chat_id` из `ALLOWED_CHAT_ID`; сообщения остальных пользователей логируются и игнорируются.
- **Retry-протокол внешнего API.** proverkacheka.com отвечает кодами «чек ещё в обработке»; клиент повторяет запрос до 5 раз с паузой под конкретный код, ожидание отменяемо через `context`.
- **Деньги в копейках.** Суммы хранятся как `int64`, в рубли конвертируются только при форматировании вывода.
- **Транзакционная запись и тестируемость.** Чек и его позиции пишутся в одной транзакции; хендлеры команд зависят от узкого интерфейса `ReceiptReader` и покрыты table-driven тестами.

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `BOT_TOKEN` | Токен Telegram-бота от @BotFather |
| `FNS_TOKEN` | API-токен от proverkacheka.com |
| `ALLOWED_CHAT_ID` | Telegram chat_id единственного разрешённого пользователя |
| `DB_PATH` | Путь до SQLite-файла, по умолчанию `./grocer.db` |
| `LOG_LEVEL` | Уровень логирования: `debug`/`info`/`warn`/`error`, по умолчанию `info` |
| `BOT_DEBUG` | Debug-режим `go-telegram-bot-api` (лог всех запросов к Telegram API), по умолчанию `false` |

## Стек

Go 1.25, `modernc.org/sqlite`, `go-telegram-bot-api/v5`, `slog`.

## О разработке

Часть работы велась через AI-агента (Claude Code) с открытым внутренним roadmap в `.claude/` и `docs/state.md`. Реализован и протестирован M1 (приём и сохранение чеков, статистика) — это и есть то, что описано в README выше. M2-M4 в roadmap — не начатая работа, а не забытая недоделка: файлы оставлены в репозитории намеренно, как рабочий журнал.

## Лицензия

MIT, см. [LICENSE](LICENSE).
