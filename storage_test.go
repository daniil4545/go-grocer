package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "grocer.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})

	store := NewStore(db)
	if err := store.InitSchema(context.Background()); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	return store
}

func TestStore_SaveReceiptAndLastReceipt(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	receipt := Receipt{
		User:       "ВкусВилл",
		UserInn:    "7704217370",
		Address:    "Новосибирск, Красный проспект, 1",
		TicketDate: "2026-05-30T10:15:00",
		TotalSum:   112345,
		Items: []Item{
			{Name: "Молоко", Price: 8990, Quantity: 1, Sum: 8990},
			{Name: "Сыр", Price: 103355, Quantity: 0.5, Sum: 103355},
		},
	}

	id, err := store.SaveReceipt(ctx, &receipt)
	if err != nil {
		t.Fatalf("save receipt: %v", err)
	}
	got, err := store.LastReceipt(ctx)
	if err != nil {
		t.Fatalf("last receipt: %v", err)
	}

	if got.ID != id {
		t.Fatalf("receipt id = %d, want %d", got.ID, id)
	}
	if got.User != "ВкусВилл" {
		t.Fatalf("user = %q, want %q", got.User, "ВкусВилл")
	}
	if got.UserInn != "7704217370" {
		t.Fatalf("user inn = %q, want %q", got.UserInn, "7704217370")
	}
	if got.Address != "Новосибирск, Красный проспект, 1" {
		t.Fatalf("address = %q, want %q", got.Address, "Новосибирск, Красный проспект, 1")
	}
	if got.TicketDate != "2026-05-30T10:15:00" {
		t.Fatalf("ticket date = %q, want %q", got.TicketDate, "2026-05-30T10:15:00")
	}
	if got.TotalSum != 112345 {
		t.Fatalf("total sum = %d, want %d", got.TotalSum, 112345)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items len = %d, want 2", len(got.Items))
	}
	if got.Items[1].Quantity != 0.5 {
		t.Fatalf("second quantity = %v, want 0.5", got.Items[1].Quantity)
	}
	if got.Items[1].Sum != 103355 {
		t.Fatalf("second sum = %d, want %d", got.Items[1].Sum, 103355)
	}
}

func TestStore_Stats(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	fixtures := []*Receipt{
		{
			User:     "Пятёрочка",
			TotalSum: 13000,
			Items: []Item{
				{Name: "Хлеб", Price: 5000, Quantity: 1, Sum: 5000},
				{Name: "Молоко", Price: 8000, Quantity: 1, Sum: 8000},
			},
		},
		{
			User:     "Магнит",
			TotalSum: 22000,
			Items: []Item{
				{Name: "Молоко", Price: 9000, Quantity: 1, Sum: 9000},
				{Name: "Сыр", Price: 13000, Quantity: 1, Sum: 13000},
			},
		},
	}
	for _, receipt := range fixtures {
		if _, err := store.SaveReceipt(ctx, receipt); err != nil {
			t.Fatalf("save receipt: %v", err)
		}
	}

	stats, err := store.Stats(ctx, 2)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}

	if stats.ReceiptCount != 2 {
		t.Fatalf("receipt count = %d, want 2", stats.ReceiptCount)
	}
	if stats.TotalSum != 35000 {
		t.Fatalf("total sum = %d, want %d", stats.TotalSum, 35000)
	}
	if len(stats.TopItems) != 2 {
		t.Fatalf("top items len = %d, want 2", len(stats.TopItems))
	}
	if stats.TopItems[0].Name != "Молоко" || stats.TopItems[0].TotalSum != 17000 {
		t.Fatalf("top item 0 = %+v, want Молоко 17000", stats.TopItems[0])
	}
	if stats.TopItems[1].Name != "Сыр" || stats.TopItems[1].TotalSum != 13000 {
		t.Fatalf("top item 1 = %+v, want Сыр 13000", stats.TopItems[1])
	}
}
