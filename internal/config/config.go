package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	TelegramBotUser  string
	MongoURI         string
	MongoDBName      string
	BannerURL        string
	AppEnv           string
	WebhookURL       string
	Port             string

	MISUsername   string
	MISPassword   string
	MISNRP        string
	MISKPDaftar   string
	MISMahasiswa  string
	MISTahun      string
	MISSemester   string
	MISJamMulai   string
	MISJamSelesai string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found or using system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "pikobot"
	}

	bannerURL := os.Getenv("BANNER_URL")
	if bannerURL == "" {
		bannerURL = "public/piko.png"
	}

	misNRP := os.Getenv("MIS_NRP")
	if misNRP == "" {
		misNRP = "3124521031"
	}

	misKPDaftar := os.Getenv("MIS_KP_DAFTAR")
	if misKPDaftar == "" {
		misKPDaftar = "7733"
	}

	misMahasiswa := os.Getenv("MIS_MAHASISWA")
	if misMahasiswa == "" {
		misMahasiswa = "25533"
	}

	misTahun := os.Getenv("MIS_TAHUN")
	if misTahun == "" {
		misTahun = "2026"
	}

	misSemester := os.Getenv("MIS_SEMESTER")
	if misSemester == "" {
		misSemester = "1"
	}

	misJamMulai := os.Getenv("MIS_JAM_MULAI")
	if misJamMulai == "" {
		misJamMulai = "08.00"
	}

	misJamSelesai := os.Getenv("MIS_JAM_SELESAI")
	if misJamSelesai == "" {
		misJamSelesai = "16.00"
	}

	return &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramBotUser:  os.Getenv("TELEGRAM_BOT_USERNAME"),
		MongoURI:         os.Getenv("MONGODB_URI"),
		MongoDBName:      dbName,
		BannerURL:        bannerURL,
		AppEnv:           appEnv,
		WebhookURL:       os.Getenv("WEBHOOK_URL"),
		Port:             port,

		MISUsername:   os.Getenv("MIS_USERNAME"),
		MISPassword:   os.Getenv("MIS_PASSWORD"),
		MISNRP:        misNRP,
		MISKPDaftar:   misKPDaftar,
		MISMahasiswa:  misMahasiswa,
		MISTahun:      misTahun,
		MISSemester:   misSemester,
		MISJamMulai:   misJamMulai,
		MISJamSelesai: misJamSelesai,
	}
}
