package bot

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	cachedBannerFileID string
	cachedBannerMutex  sync.RWMutex
)

func findBannerPath(customPath string) string {
	candidates := []string{}
	if customPath != "" && !strings.HasPrefix(customPath, "http://") && !strings.HasPrefix(customPath, "https://") {
		candidates = append(candidates, customPath)
	}
	candidates = append(candidates,
		"public/piko.png",
		"../public/piko.png",
		"../../public/piko.png",
		"/home/malakul-tech/project/pikobot/public/piko.png",
	)
	for _, p := range candidates {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	return ""
}

func formatDashboardText(user *tgbotapi.User) string {
	name := user.FirstName
	if name == "" {
		name = "Sobat Piko"
	}

	username := user.UserName
	if username == "" {
		username = "-"
	} else {
		username = "@" + username
	}

	now := time.Now().Format("02 Jan 2006")

	return fmt.Sprintf(
		"✨ *PIKO ASSISTANT* ✨\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"👋 Halo, *%s*!\n"+
			"_Your Smart All-in-One Digital Assistant_\n\n"+
			"📊 *Informasi Akun:*\n"+
			"├ 👤 *Akun:* %s\n"+
			"├ 🆔 *ID Pengguna:* `%d`\n"+
			"└ 🗓️ *Tanggal:* %s\n\n"+
			"💡 _\"Satu asisten pintar untuk segala kebutuhan harian, produktivitas, dan aktivitas Anda.\"_\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"👇 *PILIH MENU:*",
		name, username, user.ID, now,
	)
}

func (b *Bot) SendMenu(chatID int64, user *tgbotapi.User) error {
	caption := formatDashboardText(user)

	cachedBannerMutex.RLock()
	fileID := cachedBannerFileID
	cachedBannerMutex.RUnlock()

	if fileID != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
		photo.Caption = caption
		photo.ParseMode = "Markdown"
		photo.ReplyMarkup = GetMainMenuInlineKeyboard()
		if _, err := b.API.Send(photo); err == nil {
			return nil
		}
		cachedBannerMutex.Lock()
		cachedBannerFileID = ""
		cachedBannerMutex.Unlock()
	}

	localPath := findBannerPath(b.BannerURL)
	if localPath != "" {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(localPath))
		photo.Caption = caption
		photo.ParseMode = "Markdown"
		photo.ReplyMarkup = GetMainMenuInlineKeyboard()

		sentMsg, err := b.API.Send(photo)
		if err == nil {
			if len(sentMsg.Photo) > 0 {
				bestPhoto := sentMsg.Photo[len(sentMsg.Photo)-1]
				cachedBannerMutex.Lock()
				cachedBannerFileID = bestPhoto.FileID
				cachedBannerMutex.Unlock()
				log.Printf("[INFO] Banner %s terkirim & di-cache di Telegram FileID: %s", localPath, bestPhoto.FileID)
			}
			return nil
		}
		log.Printf("[WARN] Gagal mengirim banner file %s: %v", localPath, err)
	}

	if strings.HasPrefix(b.BannerURL, "http://") || strings.HasPrefix(b.BannerURL, "https://") {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(b.BannerURL))
		photo.Caption = caption
		photo.ParseMode = "Markdown"
		photo.ReplyMarkup = GetMainMenuInlineKeyboard()

		sentMsg, err := b.API.Send(photo)
		if err == nil {
			if len(sentMsg.Photo) > 0 {
				bestPhoto := sentMsg.Photo[len(sentMsg.Photo)-1]
				cachedBannerMutex.Lock()
				cachedBannerFileID = bestPhoto.FileID
				cachedBannerMutex.Unlock()
			}
			return nil
		}
		log.Printf("[WARN] Gagal mengirim banner URL %s: %v", b.BannerURL, err)
	}

	msg := tgbotapi.NewMessage(chatID, caption)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetMainMenuInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}
