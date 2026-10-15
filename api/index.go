package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pikobot/internal/bot"
	"pikobot/internal/config"
	"pikobot/internal/database"
	"pikobot/internal/mis"
)

var (
	botInstance *bot.Bot
	mongoDB     *database.MongoDB
	botOnce     sync.Once
	initErr     error
)

func initializeServices() (*bot.Bot, *database.MongoDB, error) {
	botOnce.Do(func() {
		cfg := config.Load()
		if cfg.TelegramBotToken == "" {
			initErr = fmt.Errorf("TELEGRAM_BOT_TOKEN is not set")
			return
		}

		var userRepo *database.UserRepository
		var txRepo *database.TransactionRepository
		if cfg.MongoURI != "" {
			var err error
			mongoDB, err = database.Connect(cfg.MongoURI, cfg.MongoDBName)
			if err != nil {
				log.Printf("[WARN] MongoDB connection failed: %v", err)
			} else {
				userRepo = database.NewUserRepository(mongoDB.Database)
				txRepo = database.NewTransactionRepository(mongoDB.Database)
			}
		}

		botInstance, initErr = bot.New(cfg.TelegramBotToken, userRepo, txRepo, cfg.BannerURL)
		if initErr == nil {
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
			if misClient, err := mis.NewClient(misCfg); err == nil {
				botInstance.SetMISClient(misClient)
			}
			log.Printf("[INFO] Bot authorized as @%s", botInstance.API.Self.UserName)
		}
	})
	return botInstance, mongoDB, initErr
}

func Handler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		b, db, err := initializeServices()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		mongoStatus := "disconnected / not configured"
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := db.Ping(ctx); err == nil {
				mongoStatus = "connected"
			} else {
				mongoStatus = fmt.Sprintf("error: %v", err)
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"bot_name": b.API.Self.UserName,
			"bot_id":   b.API.Self.ID,
			"mongodb":  mongoStatus,
			"message":  "Piko Bot is active on Vercel Serverless!",
		})
		return
	}

	if r.Method == http.MethodPost {
		b, _, err := initializeServices()
		if err != nil {
			log.Printf("[ERROR] Service init failed: %v", err)
			http.Error(w, "Bot initialization failed", http.StatusInternalServerError)
			return
		}

		var update tgbotapi.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			log.Printf("[ERROR] Could not decode update: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		b.ProcessUpdate(update)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
