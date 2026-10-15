package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pikobot/internal/mis"
	"pikobot/internal/utils"
)

func (b *Bot) HandleMessage(message *tgbotapi.Message) error {
	userID := message.From.ID
	state := b.GetUserState(userID)

	if state == "waiting_logbook_kegiatan" {
		kegiatan := strings.TrimSpace(message.Text)
		if kegiatan == "" {
			return nil
		}

		var entry *mis.LogbookEntry
		if b.MISClient != nil {
			cfg := b.MISClient.GetConfig()
			entry = cfg.DefaultLogbookEntry(kegiatan)
		} else {
			entry = &mis.LogbookEntry{
				Tanggal:      time.Now().Format("2006-01-02"),
				JamMulai:     "08.00",
				JamSelesai:   "16.00",
				Kegiatan:     kegiatan,
				SesuaiKuliah: 2,
				Minggu:       16,
			}
		}

		return b.ShowLogbookPreview(message.Chat.ID, message.From, entry)
	}

	if state == "waiting_logbook_jam" {
		mulai, selesai, err := ParseJamInput(message.Text)
		if err != nil {
			prompt := "⚠️ Format jam tidak valid.\n" +
				"Gunakan format: `JJ.MM - JJ.MM` (Contoh: `08.00 - 16.00` atau `09.00 - 17.00`).\n\n" +
				"Silakan kirim ulang atau klik Batal."
			msg := tgbotapi.NewMessage(message.Chat.ID, prompt)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = GetCancelLogbookInlineKeyboard()
			_, sendErr := b.API.Send(msg)
			return sendErr
		}

		draft := b.GetLogbookDraft(userID)
		if draft == nil {
			draft = &mis.LogbookEntry{
				Tanggal:      time.Now().Format("2006-01-02"),
				Kegiatan:     "Belum diisi",
				SesuaiKuliah: 2,
				Minggu:       16,
			}
		}
		draft.JamMulai = mulai
		draft.JamSelesai = selesai

		return b.ShowLogbookPreview(message.Chat.ID, message.From, draft)
	}

	if state == "waiting_set_balance" {
		nominal, err := utils.ParseRupiah(message.Text)
		if err != nil {
			prompt := "⚠️ Format nominal tidak valid.\n" +
				"Contoh input yang benar: `500000`, `500.000`, `500k`, `1jt`.\n\n" +
				"Silakan kirim ulang nominal Anda atau gunakan tombol Batal di bawah."
			msg := tgbotapi.NewMessage(message.Chat.ID, prompt)
			msg.ReplyMarkup = GetCancelSaldoInlineKeyboard()
			_, sendErr := b.API.Send(msg)
			return sendErr
		}

		b.ClearUserState(userID)
		return b.applySetBalance(message.Chat.ID, message.From, nominal)
	}

	if state == "waiting_add_balance" {
		nominal, err := utils.ParseRupiah(message.Text)
		if err != nil {
			prompt := "⚠️ Format nominal tidak valid.\n" +
				"Contoh input yang benar: `50000`, `50.000`, `50k`, `100rb`.\n\n" +
				"Silakan kirim ulang nominal Anda atau gunakan tombol Batal di bawah."
			msg := tgbotapi.NewMessage(message.Chat.ID, prompt)
			msg.ReplyMarkup = GetCancelSaldoInlineKeyboard()
			_, sendErr := b.API.Send(msg)
			return sendErr
		}

		b.ClearUserState(userID)
		return b.applyAddBalance(message.Chat.ID, message.From, nominal)
	}

	if state == "waiting_qr_text" {
		b.ClearUserState(userID)
		return b.ExecuteGenerateQR(message.Chat.ID, message.Text)
	}

	if state == "waiting_url_short" {
		b.ClearUserState(userID)
		return b.ExecuteShortenURL(message.Chat.ID, message.Text)
	}

	if state == "waiting_url_redirect" {
		b.ClearUserState(userID)
		return b.ExecuteCheckRedirects(message.Chat.ID, message.Text)
	}

	if state == "waiting_url_meta" {
		b.ClearUserState(userID)
		return b.ExecuteFetchURLMetadata(message.Chat.ID, message.Text)
	}

	if strings.HasPrefix(state, "waiting_translate_text") {
		targetLang := "en"
		if strings.Contains(state, ":") {
			targetLang = strings.SplitN(state, ":", 2)[1]
		}
		b.ClearUserState(userID)
		return b.ExecuteTranslate(message.Chat.ID, message.Text, targetLang, "auto")
	}

	if state == "waiting_convert_file" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Silakan kirimkan *file gambar atau dokumen* yang ingin dikonversi, bukan teks biasa.")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
		_, err := b.API.Send(msg)
		return err
	}

	if state == "waiting_imagemeta_photo" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Silakan kirimkan *foto atau dokumen gambar* yang ingin dicek metadatanya.")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
		_, err := b.API.Send(msg)
		return err
	}

	if state == "waiting_compress_file" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Silakan kirimkan file *PDF atau Foto* yang ingin dikompresi.")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
		_, err := b.API.Send(msg)
		return err
	}

	switch message.Text {
	case "📝 Logbook KP", "Logbook KP", "Logbook":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_logbook", Message: message, From: message.From})
	case "📄 Progres PDF", "Progres KP", "Progres":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_progres", Message: message, From: message.From})
	case "💵 Keuangan", "Keuangan":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_keuangan", Message: message, From: message.From})
	case "💰 Catat Transaksi":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_catat", Message: message, From: message.From})
	case "👛 Saldo":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_saldo", Message: message, From: message.From})
	case "📜 Riwayat Transaksi":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_riwayat", Message: message, From: message.From})
	case "📊 Laporan Keuangan":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_laporan", Message: message, From: message.From})
	case "🗂️ Kategori":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_kategori", Message: message, From: message.From})
	case "🎯 Anggaran":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_anggaran", Message: message, From: message.From})
	case "🛠️ Menu Tools", "Menu Tools", "Tools":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_tools", Message: message, From: message.From})
	case "🔲 QR Code Generator", "QR Code":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "tools_qr", Message: message, From: message.From})
	case "🔄 File Converter", "File Converter":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "tools_converter", Message: message, From: message.From})
	case "🗜️ Compress PDF & Foto", "Compress":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "tools_compress", Message: message, From: message.From})
	case "🔗 URL Tools", "URL Tools":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "tools_url", Message: message, From: message.From})
	case "🌐 Translator", "Translator":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "tools_translate", Message: message, From: message.From})
	case "🖼️ Image Metadata", "Image Metadata":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "tools_imagemeta", Message: message, From: message.From})
	case "🔔 Pengingat":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_pengingat", Message: message, From: message.From})
	case "⚙️ Pengaturan":
		return b.HandleCallbackQuery(&tgbotapi.CallbackQuery{ID: "", Data: "menu_pengaturan", Message: message, From: message.From})
	}

	replyText := fmt.Sprintf(
		"Piko menerima pesanmu:\n\n\"%s\"\n\nSilakan pilih menu di bawah ini untuk melanjutkan aktivitas Anda:",
		message.Text,
	)
	msg := tgbotapi.NewMessage(message.Chat.ID, replyText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetMainMenuInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}
