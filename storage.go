package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *sql.DB
}

type StoredReceipt struct {
	ID         int64
	User       string
	UserInn    string
	Address    string
	TicketDate string
	TotalSum   int64
	CreatedAt  time.Time
	Items      []Item
}

type ReceiptStats struct {
	ReceiptCount int64
	TotalSum     int64
	TopItems     []TopItem
}

type TopItem struct {
	Name     string
	TotalSum int64
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) InitSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS receipts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	store TEXT NOT NULL,
	user_inn TEXT NOT NULL DEFAULT '',
	retail_place_address TEXT NOT NULL DEFAULT '',
	ticket_date TEXT NOT NULL DEFAULT '',
	total_sum INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS receipt_items (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	receipt_id INTEGER NOT NULL REFERENCES receipts(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	price INTEGER NOT NULL,
	quantity REAL NOT NULL,
	sum INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_receipt_items_receipt_id ON receipt_items(receipt_id);
CREATE INDEX IF NOT EXISTS idx_receipts_created_at ON receipts(created_at);
`)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

func (s *Store) SaveReceipt(ctx context.Context, receipt *Receipt) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, `
INSERT INTO receipts (store, user_inn, retail_place_address, ticket_date, total_sum)
VALUES (?, ?, ?, ?, ?)`,
		receipt.User,
		receipt.UserInn,
		receipt.Address,
		receipt.TicketDate,
		receipt.TotalSum,
	)
	if err != nil {
		return 0, fmt.Errorf("insert receipt: %w", err)
	}

	receiptID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("receipt id: %w", err)
	}

	for _, item := range receipt.Items {
		_, err := tx.ExecContext(ctx, `
INSERT INTO receipt_items (receipt_id, name, price, quantity, sum)
VALUES (?, ?, ?, ?, ?)`,
			receiptID,
			item.Name,
			item.Price,
			item.Quantity,
			item.Sum,
		)
		if err != nil {
			return 0, fmt.Errorf("insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit receipt: %w", err)
	}

	return receiptID, nil
}

func (s *Store) LastReceipt(ctx context.Context) (*StoredReceipt, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, store, user_inn, retail_place_address, ticket_date, total_sum, created_at
FROM receipts
ORDER BY id DESC
LIMIT 1`)

	var receipt StoredReceipt
	var createdAt string
	if err := row.Scan(
		&receipt.ID,
		&receipt.User,
		&receipt.UserInn,
		&receipt.Address,
		&receipt.TicketDate,
		&receipt.TotalSum,
		&createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan receipt: %w", err)
	}
	receipt.CreatedAt = parseSQLiteTime(createdAt)

	items, err := s.receiptItems(ctx, receipt.ID)
	if err != nil {
		return nil, err
	}
	receipt.Items = items

	return &receipt, nil
}

func (s *Store) receiptItems(ctx context.Context, receiptID int64) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT name, price, quantity, sum
FROM receipt_items
WHERE receipt_id = ?
ORDER BY id`, receiptID)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Name, &item.Price, &item.Quantity, &item.Sum); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

func (s *Store) Stats(ctx context.Context, topLimit int) (*ReceiptStats, error) {
	var stats ReceiptStats
	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*), COALESCE(SUM(total_sum), 0)
FROM receipts`).Scan(&stats.ReceiptCount, &stats.TotalSum); err != nil {
		return nil, fmt.Errorf("scan stats: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT name, SUM(sum) AS total_sum
FROM receipt_items
GROUP BY name
ORDER BY total_sum DESC, name ASC
LIMIT ?`, topLimit)
	if err != nil {
		return nil, fmt.Errorf("query top items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var item TopItem
		if err := rows.Scan(&item.Name, &item.TotalSum); err != nil {
			return nil, fmt.Errorf("scan top item: %w", err)
		}
		stats.TopItems = append(stats.TopItems, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top items: %w", err)
	}

	return &stats, nil
}

func parseSQLiteTime(value string) time.Time {
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}
