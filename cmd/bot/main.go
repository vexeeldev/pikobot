package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pikobot/internal/bot"
	"pikobot/internal/config"
	"pikobot/internal/database"
	"pikobot/internal/mis"
)

func main() {
	cfg := config.Load()
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set in environment or .env file")
	}

	var userRepo *database.UserRepository
	var txRepo *database.TransactionRepository
	if cfg.MongoURI != "" {
		mongoDB, err := database.Connect(cfg.MongoURI, cfg.MongoDBName)
		if err != nil {
			log.Printf("⚠️ Warning: MongoDB connection error: %v", err)
		} else {
			userRepo = database.NewUserRepository(mongoDB.Database)
			txRepo = database.NewTransactionRepository(mongoDB.Database)
			log.Println("🍃 MongoDB successfully connected and repositories ready")
		}
	} else {
		log.Println("ℹ️ MongoDB URI is empty, running in memory-only mode")
	}

	b, err := bot.New(cfg.TelegramBotToken, userRepo, txRepo, cfg.BannerURL)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	misCfg := mis.Config{
		Username:   cfg.MISUsername,
		Password:   cfg.MISPassword,
		NRP:        cfg.MISNRP,
		KPDaftar:   cfg.MISKPDaftar,
		Mahasiswa:  cfg.MISMahasiswa,
		Tahun:      cfg.MISTahun,
		Semester:   cfg.MISSemester,
		JamMulai:   cfg.MISJamMulai,
		JamSelesai: cfg.MISJamSelesai,
	}
	misClient, err := mis.NewClient(misCfg)
	if err != nil {
		log.Printf("⚠️ Warning: MIS Client initialization error: %v", err)
	} else {
		b.SetMISClient(misClient)
		log.Println("🎓 MIS PENS Logbook Client initialized successfully")
	}

	log.Printf("🤖 Authorized on account @%s (Local Polling Mode)", b.API.Self.UserName)

	_, _ = b.API.Request(tgbotapi.DeleteWebhookConfig{})

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.API.GetUpdatesChan(u)

	log.Println("⚡ Bot is listening for incoming messages... (Press Ctrl+C to stop)")
	for update := range updates {
		b.ProcessUpdate(update)
	}
}
