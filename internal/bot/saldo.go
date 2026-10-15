package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pikobot/internal/database"
	"pikobot/internal/utils"
)

func (b *Bot) formatKeuanganDashboard(userFrom *tgbotapi.User) string {
	name := userFrom.FirstName
	if name == "" {
		name = "Sobat Piko"
	}

	var currentBalance int64 = 0
	if b.UserRepo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if u, err := b.UserRepo.GetUserByTelegramID(ctx, userFrom.ID); err == nil && u != nil {
			currentBalance = u.Balance
		}
	}

	formattedBalance := utils.FormatRupiah(currentBalance)
	now := time.Now().Format("02 Jan 2006, 15:04 WIB")

	return fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  💵 *DASHBOARD KEUANGAN*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"👋 Halo, *%s*!\n"+
			"Kelola pencatatan dan arus kas keuangan Anda di sini.\n\n"+
			"💰 *Saldo Aktif Saat Ini:*\n"+
			"👉 *`%s`*\n\n"+
			"🗓️ _Status per: %s_\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"👇 *PILIH MENU KEUANGAN:*",
		name, formattedBalance, now,
	)
}

func (b *Bot) formatSaldoCard(userFrom *tgbotapi.User) string {
	name := userFrom.FirstName
	if name == "" {
		name = "Sobat Piko"
	}

	var currentBalance int64 = 0
	if b.UserRepo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if u, err := b.UserRepo.GetUserByTelegramID(ctx, userFrom.ID); err == nil && u != nil {
			currentBalance = u.Balance
		}
	}

	formattedBalance := utils.FormatRupiah(currentBalance)
	now := time.Now().Format("02 Jan 2006, 15:04 WIB")

	return fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  👛 *INFORMASI SALDO ANDA*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"👋 Halo, *%s*!\n"+
			"Berikut adalah total saldo aktif Anda saat ini:\n\n"+
			"💰 *TOTAL SALDO:*\n"+
			"👉 *`%s`*\n\n"+
			"🗓️ _Diperbarui: %s_\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"👇 *PILIH TINDAKAN SALDO:*",
		name, formattedBalance, now,
	)
}

func (b *Bot) applySetBalance(chatID int64, userFrom *tgbotapi.User, nominal int64) error {
	if b.UserRepo == nil {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Database MongoDB belum terhubung. Konfigurasi MONGODB_URI di .env.")
		_, err := b.API.Send(msg)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := b.UserRepo.SetUserBalance(ctx, userFrom.ID, nominal); err != nil {
		log.Printf("[ERROR] Gagal menyimpan saldo: %v", err)
		msg := tgbotapi.NewMessage(chatID, "⚠️ Gagal memperbarui saldo di database.")
		_, err := b.API.Send(msg)
		return err
	}

	if b.TxRepo != nil {
		_ = b.TxRepo.Record(ctx, &database.Transaction{
			TelegramID:   userFrom.ID,
			Type:         database.TypeSetBalance,
			Amount:       nominal,
			BalanceAfter: nominal,
			Description:  "Atur ulang saldo",
		})
	}

	successText := fmt.Sprintf(
		"✅ *SALDO BERHASIL DIATUR!*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"💰 Total Saldo Baru: *`%s`*\n"+
			"🗓️ Waktu: %s\n\n"+
			"Saldo Anda telah berhasil diperbarui di database MongoDB.",
		utils.FormatRupiah(nominal),
		time.Now().Format("02 Jan 2006, 15:04 WIB"),
	)

	msg := tgbotapi.NewMessage(chatID, successText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetSaldoInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) applyAddBalance(chatID int64, userFrom *tgbotapi.User, nominal int64) error {
	if b.UserRepo == nil {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Database MongoDB belum terhubung. Konfigurasi MONGODB_URI di .env.")
		_, err := b.API.Send(msg)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newBalance, err := b.UserRepo.AddUserBalance(ctx, userFrom.ID, nominal)
	if err != nil {
		log.Printf("[ERROR] Gagal menambah saldo: %v", err)
		msg := tgbotapi.NewMessage(chatID, "⚠️ Gagal menambah saldo di database.")
		_, err := b.API.Send(msg)
		return err
	}

	if b.TxRepo != nil {
		_ = b.TxRepo.Record(ctx, &database.Transaction{
			TelegramID:   userFrom.ID,
			Type:         database.TypeAddBalance,
			Amount:       nominal,
			BalanceAfter: newBalance,
			Description:  "Tambah saldo",
		})
	}

	successText := fmt.Sprintf(
		"✅ *SALDO BERHASIL DITAMBAHKAN!*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"➕ Nominal Masuk: *`+%s`*\n"+
			"💰 Total Saldo Sekarang: *`%s`*\n"+
			"🗓️ Waktu: %s\n\n"+
			"Penambahan saldo telah tercatat di MongoDB.",
		utils.FormatRupiah(nominal),
		utils.FormatRupiah(newBalance),
		time.Now().Format("02 Jan 2006, 15:04 WIB"),
	)

	msg := tgbotapi.NewMessage(chatID, successText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetSaldoInlineKeyboard()
	_, err = b.API.Send(msg)
	return err
}
