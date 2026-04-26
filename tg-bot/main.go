package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

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

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
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
		if update.Message != nil {
			slog.Info("message received", "user", update.Message.From.UserName, "text", update.Message.Text)

			if len(update.Message.Photo) > 0 {
				data, err := HandlePhoto(bot, update.Message.Photo[len(update.Message.Photo)-1])
				if err != nil {
					slog.Error("failed to handle photo", "err", err)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Размер: %d", len(data)))
				if _, err := bot.Send(msg); err != nil {
					slog.Error("failed to send message", "err", err)
				}
			}
		}
	}
}

func HandlePhoto(bot *tgbotapi.BotAPI, photo tgbotapi.PhotoSize) ([]byte, error) {
	photoURL, err := bot.GetFileDirectURL(photo.FileID)
	if err != nil {
		return nil, fmt.Errorf("get file url: %w", err)
	}

	resp, err := http.Get(photoURL) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("download photo: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return data, nil
}
