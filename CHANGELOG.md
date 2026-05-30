# Changelog

## [Unreleased]

### Added
- `fns.go`: клиент к proverkacheka.com API — `GetReceipt`, структуры `Item`, `Receipt`, `CheckData`, `CheckResponse`
- Retry-логика для code=2 (2s) и code=4 (8s) с учётом отмены контекста (`sleepWithContext`)
- Поддержка `Message.Document` — PNG/JPEG можно отправлять как файл, не только как фото
- Фильтр по `ALLOWED_CHAT_ID` — бот отвечает только авторизованному пользователю
- Пользователь получает сообщение об ошибке если чек не удалось получить
- SQLite-хранилище чеков и позиций на `modernc.org/sqlite`
- Команды `/start`, `/last`, `/stats`
- Тесты storage, форматирования и command handling без сети и Telegram API
- `DB_PATH` для настройки пути к SQLite-файлу

### Changed
- Все файлы перенесены из `tg-bot/` в корень проекта
- `HandlePhoto` переименован в `DownloadFile(fileID string)` — принимает любой file_id
- `Data` в `CheckResponse` теперь `json.RawMessage` — корректно обрабатывает ответы API с нечисловым `data`
- Форматирование позиций через `strings.Builder` вместо `+=` в цикле
- После успешного распознавания бот сохраняет чек в БД и отвечает кратким подтверждением

### Removed
- Зависимость от gozbar / libzbar0 (CGO) — проект полностью pure Go
