package main

import (
	"context"
	"errors"
	"testing"
)

type fakeReceiptReader struct {
	lastReceipt *StoredReceipt
	stats       ReceiptStats
	lastErr     error
	statsErr    error
}

func (f fakeReceiptReader) LastReceipt(context.Context) (*StoredReceipt, error) {
	if f.lastErr != nil {
		return nil, f.lastErr
	}
	return f.lastReceipt, nil
}

func (f fakeReceiptReader) Stats(context.Context, int) (*ReceiptStats, error) {
	if f.statsErr != nil {
		return nil, f.statsErr
	}
	return &f.stats, nil
}

func TestFormatSavedReceipt(t *testing.T) {
	text := FormatSavedReceipt(&Receipt{
		User:     "ВкусВилл",
		TotalSum: 12345,
		Items: []Item{
			{Name: "Молоко", Sum: 8990},
			{Name: "Хлеб", Sum: 3355},
		},
	})

	want := "Чек сохранён: ВкусВилл\nПозиций: 2\nИтого: 123.45 руб."
	if text != want {
		t.Fatalf("text = %q, want %q", text, want)
	}
}

func TestHandleCommand(t *testing.T) {
	reader := fakeReceiptReader{
		lastReceipt: &StoredReceipt{
			ID:         7,
			User:       "Магнит",
			TicketDate: "2026-05-30T10:15:00",
			TotalSum:   22000,
			Items: []Item{
				{Name: "Молоко", Sum: 9000},
				{Name: "Сыр", Sum: 13000},
			},
		},
		stats: ReceiptStats{
			ReceiptCount: 2,
			TotalSum:     35000,
			TopItems: []TopItem{
				{Name: "Молоко", TotalSum: 17000},
				{Name: "Сыр", TotalSum: 13000},
			},
		},
	}

	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "start",
			text: "/start",
			want: "Привет! Отправь фото QR-кода или документ с чеком, а я сохраню покупки и покажу статистику.",
		},
		{
			name: "last",
			text: "/last",
			want: "Последний чек #7\nМагнит\nДата: 2026-05-30T10:15:00\nИтого: 220.00 руб.\n\nПозиции:\nМолоко — 90.00 руб.\nСыр — 130.00 руб.",
		},
		{
			name: "stats",
			text: "/stats",
			want: "Статистика\nЧеков: 2\nВсего: 350.00 руб.\n\nТоп товаров:\n1. Молоко — 170.00 руб.\n2. Сыр — 130.00 руб.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := HandleCommand(context.Background(), reader, tt.text)
			if err != nil {
				t.Fatalf("handle command: %v", err)
			}
			if !ok {
				t.Fatalf("ok = false, want true")
			}
			if got != tt.want {
				t.Fatalf("text = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandleCommand_EmptyStorage(t *testing.T) {
	reader := fakeReceiptReader{
		lastErr: ErrNotFound,
		stats:   ReceiptStats{},
	}

	got, ok, err := HandleCommand(context.Background(), reader, "/last")
	if err != nil {
		t.Fatalf("handle last: %v", err)
	}
	if !ok {
		t.Fatalf("ok = false, want true")
	}
	if got != "Пока нет сохранённых чеков." {
		t.Fatalf("last text = %q", got)
	}

	got, ok, err = HandleCommand(context.Background(), reader, "/stats")
	if err != nil {
		t.Fatalf("handle stats: %v", err)
	}
	if !ok {
		t.Fatalf("ok = false, want true")
	}
	if got != "Пока нет сохранённых чеков." {
		t.Fatalf("stats text = %q", got)
	}
}

func TestHandleCommand_IgnoresUnknownText(t *testing.T) {
	got, ok, err := HandleCommand(context.Background(), fakeReceiptReader{}, "hello")
	if err != nil {
		t.Fatalf("handle command: %v", err)
	}
	if ok {
		t.Fatalf("ok = true, want false")
	}
	if got != "" {
		t.Fatalf("text = %q, want empty", got)
	}
}

func TestHandleCommand_ReturnsUnexpectedStorageError(t *testing.T) {
	reader := fakeReceiptReader{lastErr: errors.New("db is closed")}

	_, _, err := HandleCommand(context.Background(), reader, "/last")
	if err == nil {
		t.Fatalf("err = nil, want error")
	}
}
