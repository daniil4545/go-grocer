package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type ReceiptReader interface {
	LastReceipt(context.Context) (*StoredReceipt, error)
	Stats(context.Context, int) (*ReceiptStats, error)
}

func HandleCommand(ctx context.Context, reader ReceiptReader, text string) (string, bool, error) {
	switch strings.TrimSpace(text) {
	case "/start":
		return "Привет! Отправь фото QR-кода или документ с чеком, а я сохраню покупки и покажу статистику.", true, nil
	case "/last":
		receipt, err := reader.LastReceipt(ctx)
		if errors.Is(err, ErrNotFound) {
			return emptyStorageMessage(), true, nil
		}
		if err != nil {
			return "", true, err
		}
		return FormatLastReceipt(receipt), true, nil
	case "/stats":
		stats, err := reader.Stats(ctx, 5)
		if err != nil {
			return "", true, err
		}
		if stats.ReceiptCount == 0 {
			return emptyStorageMessage(), true, nil
		}
		return FormatStats(stats), true, nil
	default:
		return "", false, nil
	}
}

func FormatSavedReceipt(receipt *Receipt) string {
	store := receipt.User
	if store == "" {
		store = "магазин не указан"
	}
	return fmt.Sprintf(
		"Чек сохранён: %s\nПозиций: %d\nИтого: %s руб.",
		store,
		len(receipt.Items),
		formatMoney(receipt.TotalSum),
	)
}

func FormatLastReceipt(receipt *StoredReceipt) string {
	store := receipt.User
	if store == "" {
		store = "Магазин не указан"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Последний чек #%d\n%s", receipt.ID, store)
	if receipt.TicketDate != "" {
		fmt.Fprintf(&sb, "\nДата: %s", receipt.TicketDate)
	}
	fmt.Fprintf(&sb, "\nИтого: %s руб.", formatMoney(receipt.TotalSum))

	if len(receipt.Items) > 0 {
		sb.WriteString("\n\nПозиции:")
		for i, item := range receipt.Items {
			if i == 5 {
				fmt.Fprintf(&sb, "\n...ещё %d", len(receipt.Items)-i)
				break
			}
			fmt.Fprintf(&sb, "\n%s — %s руб.", item.Name, formatMoney(item.Sum))
		}
	}

	return sb.String()
}

func FormatStats(stats *ReceiptStats) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Статистика\nЧеков: %d\nВсего: %s руб.", stats.ReceiptCount, formatMoney(stats.TotalSum))

	if len(stats.TopItems) > 0 {
		sb.WriteString("\n\nТоп товаров:")
		for i, item := range stats.TopItems {
			fmt.Fprintf(&sb, "\n%d. %s — %s руб.", i+1, item.Name, formatMoney(item.TotalSum))
		}
	}

	return sb.String()
}

func formatMoney(kopecks int64) string {
	return fmt.Sprintf("%.2f", float64(kopecks)/100)
}

func emptyStorageMessage() string {
	return "Пока нет сохранённых чеков."
}
