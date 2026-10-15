package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pikobot/internal/mis"
	"pikobot/internal/tools"
	"pikobot/internal/utils"
)

func (b *Bot) HandleCommand(message *tgbotapi.Message) error {
	switch message.Command() {
	case "start", "menu":
		return b.SendMenu(message.Chat.ID, message.From)

	case "logbook":
		args := strings.TrimSpace(message.CommandArguments())
		if args != "" {
			var entry *mis.LogbookEntry
			if b.MISClient != nil {
				cfg := b.MISClient.GetConfig()
				entry = cfg.DefaultLogbookEntry(args)
			} else {
				entry = &mis.LogbookEntry{
					Tanggal:      time.Now().Format("2006-01-02"),
					JamMulai:     "08.00",
					JamSelesai:   "16.00",
					Kegiatan:     args,
					SesuaiKuliah: 2,
					Minggu:       16,
				}
			}
			return b.ShowLogbookPreview(message.Chat.ID, message.From, entry)
		}
		return b.PromptLogbookKegiatan(message.Chat.ID, message.From)

	case "progres":
		return b.PromptAttachPDF(message.Chat.ID, message.From)

	case "keuangan":
		keuanganText := b.formatKeuanganDashboard(message.From)
		msg := tgbotapi.NewMessage(message.Chat.ID, keuanganText)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = GetKeuanganInlineKeyboard()
		_, err := b.API.Send(msg)
		return err

	case "saldo":
		cardText := b.formatSaldoCard(message.From)
		msg := tgbotapi.NewMessage(message.Chat.ID, cardText)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = GetSaldoInlineKeyboard()
		_, err := b.API.Send(msg)
		return err

	case "setsaldo":
		args := strings.TrimSpace(message.CommandArguments())
		if args == "" {
			b.SetUserState(message.From.ID, "waiting_set_balance")
			prompt := "━━━━━━━━━━━━━━━━━━━━━━\n" +
				"  ✏️ *ATUR ULANG SALDO*\n" +
				"━━━━━━━━━━━━━━━━━━━━━━\n" +
				"Ketik nominal saldo yang ingin Anda tetapkan.\n\n" +
				"💡 *Contoh:* `500000`, `500k`, `1.5jt`\n\n" +
				"Kirim nominal sekarang:"
			msg := tgbotapi.NewMessage(message.Chat.ID, prompt)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = GetCancelSaldoInlineKeyboard()
			_, err := b.API.Send(msg)
			return err
		}

		nominal, err := utils.ParseRupiah(args)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Nominal tidak valid. Contoh input: `500.000` atau `500k`")
			_, err := b.API.Send(msg)
			return err
		}

		return b.applySetBalance(message.Chat.ID, message.From, nominal)

	case "tambahsaldo":
		args := strings.TrimSpace(message.CommandArguments())
		if args == "" {
			b.SetUserState(message.From.ID, "waiting_add_balance")
			prompt := "━━━━━━━━━━━━━━━━━━━━━━\n" +
				"  💵 *TAMBAH SALDO*\n" +
				"━━━━━━━━━━━━━━━━━━━━━━\n" +
				"Ketik nominal yang ingin Anda tambahkan ke saldo saat ini.\n\n" +
				"💡 *Contoh:* `50000`, `50k`, `100rb`\n\n" +
				"Kirim nominal sekarang:"
			msg := tgbotapi.NewMessage(message.Chat.ID, prompt)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = GetCancelSaldoInlineKeyboard()
			_, err := b.API.Send(msg)
			return err
		}

		nominal, err := utils.ParseRupiah(args)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Nominal tidak valid. Contoh input: `50.000` atau `50k`")
			_, err := b.API.Send(msg)
			return err
		}

		return b.applyAddBalance(message.Chat.ID, message.From, nominal)

	case "tools":
		return b.SendToolsMenu(message.Chat.ID, message.From)

	case "qr":
		args := strings.TrimSpace(message.CommandArguments())
		if args != "" {
			return b.ExecuteGenerateQR(message.Chat.ID, args)
		}
		return b.PromptQR(message.Chat.ID, message.From)

	case "convert":
		args := strings.TrimSpace(message.CommandArguments())
		targetFormat, _ := tools.ParseConvertArgs(args)
		if targetFormat == "" {
			targetFormat = "png"
		}

		if message.ReplyToMessage != nil {
			if message.ReplyToMessage.Document != nil {
				return b.ExecuteConvertFile(message.Chat.ID, message.ReplyToMessage.Document.FileID, message.ReplyToMessage.Document.FileName, targetFormat)
			}
			if len(message.ReplyToMessage.Photo) > 0 {
				largest := message.ReplyToMessage.Photo[len(message.ReplyToMessage.Photo)-1]
				return b.ExecuteConvertFile(message.Chat.ID, largest.FileID, "photo.jpg", targetFormat)
			}
		}

		if message.Document != nil {
			return b.ExecuteConvertFile(message.Chat.ID, message.Document.FileID, message.Document.FileName, targetFormat)
		}
		if len(message.Photo) > 0 {
			largest := message.Photo[len(message.Photo)-1]
			return b.ExecuteConvertFile(message.Chat.ID, largest.FileID, "photo.jpg", targetFormat)
		}

		return b.PromptConverter(message.Chat.ID, message.From)

	case "compress":

		if message.ReplyToMessage != nil {
			if message.ReplyToMessage.Document != nil {
				return b.ExecuteCompressFile(message.Chat.ID, message.ReplyToMessage.Document.FileID, message.ReplyToMessage.Document.FileName, 0)
			}
			if len(message.ReplyToMessage.Photo) > 0 {
				largest := message.ReplyToMessage.Photo[len(message.ReplyToMessage.Photo)-1]
				return b.ExecuteCompressFile(message.Chat.ID, largest.FileID, "photo.jpg", 0)
			}
		}

		if message.Document != nil {
			return b.ExecuteCompressFile(message.Chat.ID, message.Document.FileID, message.Document.FileName, 0)
		}
		if len(message.Photo) > 0 {
			largest := message.Photo[len(message.Photo)-1]
			return b.ExecuteCompressFile(message.Chat.ID, largest.FileID, "photo.jpg", 0)
		}

		return b.PromptCompress(message.Chat.ID, message.From)

	case "url":
		args := strings.TrimSpace(message.CommandArguments())
		if args == "" {
			return b.PromptURLTools(message.Chat.ID, message.From)
		}
		lowerArgs := strings.ToLower(args)
		if strings.HasPrefix(lowerArgs, "short ") {
			return b.ExecuteShortenURL(message.Chat.ID, strings.TrimSpace(args[6:]))
		}
		if strings.HasPrefix(lowerArgs, "check ") {
			return b.ExecuteCheckRedirects(message.Chat.ID, strings.TrimSpace(args[6:]))
		}
		if strings.HasPrefix(lowerArgs, "redirect ") {
			return b.ExecuteCheckRedirects(message.Chat.ID, strings.TrimSpace(args[9:]))
		}
		if strings.HasPrefix(lowerArgs, "meta ") {
			return b.ExecuteFetchURLMetadata(message.Chat.ID, strings.TrimSpace(args[5:]))
		}
		return b.ExecuteShortenURL(message.Chat.ID, args)

	case "short", "shortlink":
		args := strings.TrimSpace(message.CommandArguments())
		if args != "" {
			return b.ExecuteShortenURL(message.Chat.ID, args)
		}
		b.SetUserState(message.From.ID, "waiting_url_short")
		msg := tgbotapi.NewMessage(message.Chat.ID, "🔗 *Masukkan URL yang ingin diperpendek:*\n\nContoh: `https://example.com/very/long/url`")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
		_, err := b.API.Send(msg)
		return err

	case "redirect", "expand":
		args := strings.TrimSpace(message.CommandArguments())
		if args != "" {
			return b.ExecuteCheckRedirects(message.Chat.ID, args)
		}
		b.SetUserState(message.From.ID, "waiting_url_redirect")
		msg := tgbotapi.NewMessage(message.Chat.ID, "🔀 *Masukkan URL yang ingin dicek redirect-nya:*\n\nContoh: `https://bit.ly/example`")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
		_, err := b.API.Send(msg)
		return err

	case "urlmeta":
		args := strings.TrimSpace(message.CommandArguments())
		if args != "" {
			return b.ExecuteFetchURLMetadata(message.Chat.ID, args)
		}
		b.SetUserState(message.From.ID, "waiting_url_meta")
		msg := tgbotapi.NewMessage(message.Chat.ID, "📋 *Masukkan URL yang ingin dicek metadatanya:*\n\nContoh: `https://github.com`")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
		_, err := b.API.Send(msg)
		return err

	case "translate", "tr":
		args := strings.TrimSpace(message.CommandArguments())

		if message.ReplyToMessage != nil && message.ReplyToMessage.Text != "" {
			targetLang := "en"
			if args != "" {
				targetLang = args
			}
			return b.ExecuteTranslate(message.Chat.ID, message.ReplyToMessage.Text, targetLang, "auto")
		}

		if args == "" {
			return b.PromptTranslator(message.Chat.ID, message.From, "en")
		}

		parts := strings.SplitN(args, " ", 2)
		if len(parts) == 1 {
			return b.ExecuteTranslate(message.Chat.ID, parts[0], "en", "auto")
		}
		return b.ExecuteTranslate(message.Chat.ID, parts[1], parts[0], "auto")

	case "metadata", "imgmeta":
		if message.ReplyToMessage != nil {
			if message.ReplyToMessage.Document != nil {
				return b.ExecuteImageMetadata(message.Chat.ID, message.ReplyToMessage.Document.FileID, message.ReplyToMessage.Document.FileName)
			}
			if len(message.ReplyToMessage.Photo) > 0 {
				largest := message.ReplyToMessage.Photo[len(message.ReplyToMessage.Photo)-1]
				return b.ExecuteImageMetadata(message.Chat.ID, largest.FileID, "photo.jpg")
			}
		}
		if message.Document != nil {
			return b.ExecuteImageMetadata(message.Chat.ID, message.Document.FileID, message.Document.FileName)
		}
		if len(message.Photo) > 0 {
			largest := message.Photo[len(message.Photo)-1]
			return b.ExecuteImageMetadata(message.Chat.ID, largest.FileID, "photo.jpg")
		}
		return b.PromptImageMetadata(message.Chat.ID, message.From)

	case "help":
		helpText := "📖 *PANDUAN PENGGUNAAN PIKO ASSISTANT*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━\n" +
			"🏠 *Navigasi & Dashboard:*\n" +
			"• /menu - Buka dashboard utama\n" +
			"• /tools - Menu kumpulan alat cerdas\n" +
			"• /help - Menampilkan panduan bantuan ini\n\n" +
			"🛠️ *Tools & Utilitas:*\n" +
			"• /qr <teks/link> - Buat gambar QR Code\n" +
			"• /convert <format> - Konversi file (reply file dg `/convert png`)\n" +
			"• /compress - Kompres ukuran file PDF atau foto\n" +
			"• /url - Buka menu network URL tools\n" +
			"• /short <url> - Buat tautan pendek (shortlink)\n" +
			"• /redirect <url> - Lacak rantai pengalihan URL\n" +
			"• /urlmeta <url> - Cek metadata & info web server\n" +
			"• /translate <target> <teks> - Terjemahkan bahasa (alias: `/tr`)\n" +
			"• /metadata - Periksa metadata resolusi & EXIF foto\n\n" +
			"💵 *Keuangan:*\n" +
			"• /saldo - Cek dan kelola saldo dompet\n" +
			"• /setsaldo <nominal> - Atur saldo (misal: `/setsaldo 500k`)\n" +
			"• /tambahsaldo <nominal> - Tambah saldo (misal: `/tambahsaldo 50k`)\n\n" +
			"📝 *Logbook KP:*\n" +
			"• /logbook - Isi catatan logbook harian\n" +
			"• /progres - Upload laporan mingguan PDF\n\n" +
			"⚙️ *Sistem & Info:*\n" +
			"• /stats - Cek statistik database MongoDB\n" +
			"• /ping - Tes respon koneksi bot\n\n" +
			"💡 Anda juga bisa langsung mengetuk tombol menu interaktif pada dashboard."
		msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
		msg.ParseMode = "Markdown"
		_, err := b.API.Send(msg)
		return err

	case "ping":
		msg := tgbotapi.NewMessage(message.Chat.ID, "🏓 *Pong!* Piko Assistant aktif dan siap membantu segala aktivitas Anda.")
		msg.ParseMode = "Markdown"
		_, err := b.API.Send(msg)
		return err

	case "stats":
		var statsText string
		if b.UserRepo == nil {
			statsText = "⚠️ Database MongoDB belum terhubung. Pastikan MONGODB_URI sudah diisi dengan benar di .env."
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			count, err := b.UserRepo.CountUsers(ctx)
			if err != nil {
				statsText = "⚠️ Gagal mengambil data statistik dari database."
			} else {
				user, _ := b.UserRepo.GetUserByTelegramID(ctx, message.From.ID)
				userMsgCount := int64(1)
				if user != nil {
					userMsgCount = user.TotalMsgs
				}
				statsText = fmt.Sprintf(
					"📊 *STATISTIK DATABASE MONGODB*\n"+
						"━━━━━━━━━━━━━━━━━━━━━━\n"+
						"👥 Total Pengguna Terdaftar: *%d*\n"+
						"💬 Total Interaksi Anda: *%d*\n"+
						"🍃 Status Database: *Connected (Atlas)*",
					count, userMsgCount,
				)
			}
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, statsText)
		msg.ParseMode = "Markdown"
		_, err := b.API.Send(msg)
		return err

	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Perintah tidak dikenali. Silakan buka menu utama atau bantuan.")
		_, err := b.API.Send(msg)
		return err
	}
}
