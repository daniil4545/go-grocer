# Changelog

## [Unreleased]

### Added
- `tg-bot/fns.go`: клиент к proverkacheka.com API — `GetReceipt`, структуры `Item`, `Receipt`, `CheckData`, `CheckResponse`
- Приём фото QR и отправка в API как `multipart/form-data` (qrfile)
- Форматирование списка позиций чека с суммами в рублях и итогом

### Changed
- `tg-bot/main.go`: вместо размера файла бот возвращает список товаров из чека
- Архитектура: gozbar и OCR-fallback исключены, QR распознаётся на стороне proverkacheka.com

### Removed
- Зависимость от gozbar / libzbar0 (CGO) — проект теперь полностью pure Go
