package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pikobot/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Penggunaan: go run cmd/setwebhook/main.go <URL_VERCEL_ANDA>")
		fmt.Println("Contoh:    go run cmd/setwebhook/main.go https://pikobot.vercel.app")
		os.Exit(1)
	}

	vercelURL := os.Args[1]

	webhookURL := vercelURL + "/"

	cfg := config.Load()
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set in environment or .env file")
	}

	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Gagal terhubung ke Telegram Bot API: %v", err)
	}

	whConfig, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		log.Fatalf("URL Webhook tidak valid: %v", err)
	}

	_, err = botAPI.Request(whConfig)
	if err != nil {
		log.Fatalf("Gagal mendaftarkan webhook: %v", err)
	}

	info, err := botAPI.GetWebhookInfo()
	if err != nil {
		log.Printf("Peringatan: Gagal mendapatkan info webhook: %v", err)
	} else {
		log.Printf("✅ Webhook berhasil diatur ke: %s", info.URL)
		log.Printf("📊 Pending update count: %d", info.PendingUpdateCount)
	}
}
