package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelDebug})))

	err := godotenv.Load()
	if err != nil {
		slog.Error("failed to load .env", "err", err)
		os.Exit(1)
	}

	botToken := os.Getenv("BOT_TOKEN")
	fnstoken := os.Getenv("FNS_TOKEN")
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./grocer.db"
	}

	allowedChatID, err := strconv.ParseInt(os.Getenv("ALLOWED_CHAT_ID"), 10, 64)
	if err != nil {
		slog.Error("invalid ALLOWED_CHAT_ID", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		slog.Error("failed to open db", "err", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	store := NewStore(db)
	if err := store.InitSchema(ctx); err != nil {
		slog.Error("failed to init db schema", "err", err)
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		slog.Error("failed to create bot", "err", err)
		os.Exit(1)
	}

	bot.Debug = true

	slog.Info("authorized", "username", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat.ID != allowedChatID {
			slog.Warn("unauthorized message", "chat_id", update.Message.Chat.ID)
			continue
		}

		slog.Info("message received", "user", update.Message.From.UserName, "text", update.Message.Text)

		if text, ok, err := HandleCommand(ctx, store, update.Message.Text); ok {
			if err != nil {
				slog.Error("failed to handle command", "err", err)
				text = "Не удалось выполнить команду. Попробуй ещё раз."
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			if _, err := bot.Send(msg); err != nil {
				slog.Error("failed to send message", "err", err)
			}
			continue
		}

		var fileID string
		if len(update.Message.Photo) > 0 {
			fileID = update.Message.Photo[len(update.Message.Photo)-1].FileID
		} else if update.Message.Document != nil {
			fileID = update.Message.Document.FileID
		}

		if fileID == "" {
			continue
		}

		photoBytes, err := DownloadFile(bot, fileID)
		if err != nil {
			slog.Error("failed to download file", "err", err)
			continue
		}

		receipt, err := GetReceipt(ctx, fnstoken, photoBytes)
		if err != nil {
			slog.Error("failed to get receipt", "err", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить чек. Попробуй ещё раз.")
			bot.Send(msg) //nolint:errcheck
			continue
		}

		if _, err := store.SaveReceipt(ctx, receipt); err != nil {
			slog.Error("failed to save receipt", "err", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Чек получен, но не удалось сохранить его в базу.")
			bot.Send(msg) //nolint:errcheck
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, FormatSavedReceipt(receipt))
		if _, err := bot.Send(msg); err != nil {
			slog.Error("failed to send message", "err", err)
		}
	}
}

func DownloadFile(bot *tgbotapi.BotAPI, fileID string) ([]byte, error) {
	photoURL, err := bot.GetFileDirectURL(fileID)
	if err != nil {
		return nil, fmt.Errorf("get file url: %w", err)
	}

	resp, err := http.Get(photoURL) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("download photo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return data, nil
}
