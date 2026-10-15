package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) editCallbackMessage(query *tgbotapi.CallbackQuery, text string, markup *tgbotapi.InlineKeyboardMarkup) error {
	if query == nil || query.Message == nil {
		return nil
	}

	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	isMedia := query.Message.Text == "" ||
		len(query.Message.Photo) > 0 ||
		query.Message.Document != nil ||
		query.Message.Animation != nil ||
		query.Message.Video != nil ||
		query.Message.Audio != nil ||
		query.Message.Voice != nil ||
		query.Message.Caption != ""

	if isMedia {
		editCap := tgbotapi.NewEditMessageCaption(chatID, messageID, text)
		editCap.ParseMode = "Markdown"
		editCap.ReplyMarkup = markup
		if _, err := b.API.Send(editCap); err == nil {
			return nil
		} else if strings.Contains(strings.ToLower(err.Error()), "not modified") {
			return nil
		}
	}

	editText := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editText.ParseMode = "Markdown"
	if markup != nil {
		editText.ReplyMarkup = markup
	}
	if _, err := b.API.Send(editText); err == nil {
		return nil
	} else if strings.Contains(strings.ToLower(err.Error()), "not modified") {
		return nil
	}

	if !isMedia {
		editCap := tgbotapi.NewEditMessageCaption(chatID, messageID, text)
		editCap.ParseMode = "Markdown"
		editCap.ReplyMarkup = markup
		if _, err := b.API.Send(editCap); err == nil {
			return nil
		} else if strings.Contains(strings.ToLower(err.Error()), "not modified") {
			return nil
		}
	}

	sendMsg := tgbotapi.NewMessage(chatID, text)
	sendMsg.ParseMode = "Markdown"
	if markup != nil {
		sendMsg.ReplyMarkup = markup
	}
	_, err := b.API.Send(sendMsg)
	return err
}

func (b *Bot) removeReplyMarkup(chatID int64, messageID int) {
	editMarkup := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	})
	_, _ = b.API.Send(editMarkup)
}

func isToolResultMessage(msg *tgbotapi.Message) bool {
	if msg == nil {
		return false
	}
	if msg.Document != nil {
		return true
	}
	text := msg.Text
	if text == "" {
		text = msg.Caption
	}
	return strings.Contains(text, "KOMPRESI BERHASIL") ||
		strings.Contains(text, "QR CODE BERHASIL DIBUAT") ||
		strings.Contains(text, "FILE BERHASIL DIKONVERSI") ||
		strings.Contains(text, "IMAGE METADATA") ||
		strings.Contains(text, "HASIL TERJEMAHAN") ||
		strings.Contains(text, "SHORT LINK BERHASIL DIBUAT") ||
		strings.Contains(text, "PEMERIKSAAN REDIRECT") ||
		strings.Contains(text, "METADATA URL")
}

func (b *Bot) HandleCallbackQuery(query *tgbotapi.CallbackQuery) error {
	callback := tgbotapi.NewCallback(query.ID, "")
	_, _ = b.API.Request(callback)

	chatID := query.Message.Chat.ID

	if query.Data == "tools_back_to_tools" {
		b.ClearUserState(query.From.ID)
		b.removeReplyMarkup(chatID, query.Message.MessageID)
		return b.SendToolsMenu(chatID, query.From)
	}

	if query.Data == "tools_back_to_main" {
		b.ClearUserState(query.From.ID)
		b.removeReplyMarkup(chatID, query.Message.MessageID)
		return b.SendMenu(chatID, query.From)
	}

	if query.Data == "tools_back_to_url" {
		b.ClearUserState(query.From.ID)
		b.removeReplyMarkup(chatID, query.Message.MessageID)
		return b.PromptURLTools(chatID, query.From)
	}

	if query.Data == "menu_main" {
		b.ClearUserState(query.From.ID)
		if isToolResultMessage(query.Message) {
			b.removeReplyMarkup(chatID, query.Message.MessageID)
			return b.SendMenu(chatID, query.From)
		}
		dashboardText := formatDashboardText(query.From)
		mainMarkup := GetMainMenuInlineKeyboard()
		return b.editCallbackMessage(query, dashboardText, &mainMarkup)
	}

	if query.Data == "menu_logbook" {
		return b.PromptLogbookKegiatan(chatID, query.From)
	}

	if query.Data == "logbook_submit" {
		return b.HandleSubmitLogbook(query)
	}

	if query.Data == "logbook_change_time" {
		return b.PromptChangeJam(chatID, query.From)
	}

	if query.Data == "logbook_prompt_photo" {
		return b.PromptAttachPhoto(chatID, query.From)
	}

	if query.Data == "logbook_remove_photo" {
		return b.HandleRemovePhoto(query)
	}

	if query.Data == "logbook_prompt_pdf" {
		return b.PromptAttachPDF(chatID, query.From)
	}

	if query.Data == "logbook_remove_pdf" {
		return b.HandleRemovePDF(query)
	}

	if query.Data == "logbook_cancel" {
		b.ClearUserState(query.From.ID)
		b.ClearLogbookDraft(query.From.ID)

		cancelText := "❌ *Pengisian logbook dibatalkan.*\nBuka menu utama kapan saja untuk mengisi kembali."
		cancelMarkup := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "menu_main"),
			),
		)
		return b.editCallbackMessage(query, cancelText, &cancelMarkup)
	}

	if query.Data == "menu_keuangan" {
		b.ClearUserState(query.From.ID)
		keuanganText := b.formatKeuanganDashboard(query.From)
		keuanganMarkup := GetKeuanganInlineKeyboard()
		return b.editCallbackMessage(query, keuanganText, &keuanganMarkup)
	}

	if query.Data == "menu_saldo" {
		b.ClearUserState(query.From.ID)
		saldoCard := b.formatSaldoCard(query.From)
		saldoMarkup := GetSaldoInlineKeyboard()
		return b.editCallbackMessage(query, saldoCard, &saldoMarkup)
	}

	if query.Data == "saldo_set" {
		b.SetUserState(query.From.ID, "waiting_set_balance")
		prompt := "━━━━━━━━━━━━━━━━━━━━━━\n" +
			"  ✏️ *ATUR ULANG SALDO*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━\n\n" +
			"Silakan ketik nominal saldo baru Anda.\n\n" +
			"💡 *Contoh format yang didukung:*\n" +
			"• `500000` atau `500.000`\n" +
			"• `Rp 500.000`\n" +
			"• `500k` atau `1.5jt`\n\n" +
			"Kirim nominal sekarang via chat:"
		cancelMarkup := GetCancelSaldoInlineKeyboard()
		return b.editCallbackMessage(query, prompt, &cancelMarkup)
	}

	if query.Data == "saldo_add" {
		b.SetUserState(query.From.ID, "waiting_add_balance")
		prompt := "━━━━━━━━━━━━━━━━━━━━━━\n" +
			"  💵 *TAMBAH SALDO*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━\n\n" +
			"Berapa nominal yang ingin Anda tambahkan ke saldo saat ini?\n\n" +
			"💡 *Contoh format yang didukung:*\n" +
			"• `50000` atau `50.000`\n" +
			"• `50k` atau `100rb`\n" +
			"• `1jt`\n\n" +
			"Kirim nominal sekarang via chat:"
		cancelMarkup := GetCancelSaldoInlineKeyboard()
		return b.editCallbackMessage(query, prompt, &cancelMarkup)
	}

	if query.Data == "saldo_cancel" {
		b.ClearUserState(query.From.ID)
		saldoCard := b.formatSaldoCard(query.From)
		saldoMarkup := GetSaldoInlineKeyboard()
		return b.editCallbackMessage(query, saldoCard, &saldoMarkup)
	}

	if query.Data == "menu_tools" {
		b.ClearUserState(query.From.ID)
		if isToolResultMessage(query.Message) {
			b.removeReplyMarkup(chatID, query.Message.MessageID)
			return b.SendToolsMenu(chatID, query.From)
		}
		toolsText := formatToolsMenuText(query.From)
		toolsMarkup := GetToolsInlineKeyboard()
		return b.editCallbackMessage(query, toolsText, &toolsMarkup)
	}

	if query.Data == "tools_qr" {
		return b.PromptQR(chatID, query.From)
	}

	if query.Data == "tools_converter" {
		return b.PromptConverter(chatID, query.From)
	}

	if query.Data == "tools_compress" {
		return b.PromptCompress(chatID, query.From)
	}

	if query.Data == "tools_url" {
		return b.PromptURLTools(chatID, query.From)
	}

	if query.Data == "tools_url_short" {
		b.SetUserState(query.From.ID, "waiting_url_short")
		prompt := "━━━━━━━━━━━━━━━━━━━━━━\n" +
			"  ✂️ *BUAT SHORT LINK*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━\n" +
			"Kirimkan tautan URL panjang yang ingin Anda perpendek:\n\n" +
			"💡 *Contoh:* `https://github.com/google/gemini`\n\n" +
			"Kirim link sekarang via chat:"
		cancelMarkup := GetCancelToolsInlineKeyboard()
		return b.editCallbackMessage(query, prompt, &cancelMarkup)
	}

	if query.Data == "tools_url_redirect" {
		b.SetUserState(query.From.ID, "waiting_url_redirect")
		prompt := "━━━━━━━━━━━━━━━━━━━━━━\n" +
			"  🔀 *PERIKSA REDIRECT URL*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━\n" +
			"Kirimkan tautan URL yang ingin Anda periksa jejak pengalihannya:\n\n" +
			"💡 *Contoh:* `https://bit.ly/example` atau `t.co/abc`\n\n" +
			"Kirim link sekarang via chat:"
		cancelMarkup := GetCancelToolsInlineKeyboard()
		return b.editCallbackMessage(query, prompt, &cancelMarkup)
	}

	if query.Data == "tools_url_meta" {
		b.SetUserState(query.From.ID, "waiting_url_meta")
		prompt := "━━━━━━━━━━━━━━━━━━━━━━\n" +
			"  📋 *CEK METADATA URL WEB*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━\n" +
			"Kirimkan tautan URL website yang ingin Anda telusuri judul, deskripsi, dan info servernya:\n\n" +
			"💡 *Contoh:* `https://pens.ac.id` atau `https://github.com`\n\n" +
			"Kirim link sekarang via chat:"
		cancelMarkup := GetCancelToolsInlineKeyboard()
		return b.editCallbackMessage(query, prompt, &cancelMarkup)
	}

	if query.Data == "tools_translate" {
		return b.PromptTranslator(chatID, query.From, "en")
	}

	if strings.HasPrefix(query.Data, "tools_tr_") {
		lang := strings.TrimPrefix(query.Data, "tools_tr_")
		return b.PromptTranslator(chatID, query.From, lang)
	}

	if query.Data == "tools_imagemeta" {
		return b.PromptImageMetadata(chatID, query.From)
	}

	if query.Data == "tools_cancel" {
		b.ClearUserState(query.From.ID)
		toolsText := formatToolsMenuText(query.From)
		toolsMarkup := GetToolsInlineKeyboard()
		return b.editCallbackMessage(query, toolsText, &toolsMarkup)
	}

	var title, desc string
	isFinanceMenu := false

	switch query.Data {
	case "menu_catat":
		title = "💰 CATAT TRANSAKSI"
		desc = "Catat pemasukan dan pengeluaran harian Anda secara cepat.\n\n" +
			"💡 *Format cepat yang akan segera hadir:*\n" +
			"• `+50000 Gaji sampingan` _(Pemasukan)_\n" +
			"• `-25000 Makan siang warteg` _(Pengeluaran)_"
		isFinanceMenu = true
	case "menu_riwayat":
		title = "📜 RIWAYAT TRANSAKSI TERAKHIR"
		desc = "Daftar catatan transaksi yang pernah Anda input lengkap dengan nominal, tanggal, dan kategorinya."
		isFinanceMenu = true
	case "menu_laporan":
		title = "📊 LAPORAN & REKAP KEUANGAN"
		desc = "Ringkasan analisis keuangan Anda:\n\n" +
			"• 📈 *Arus Kas Bulanan*\n" +
			"• 🥧 *Distribusi Pengeluaran per Kategori*\n" +
			"• 📥 *Ekspor Laporan PDF / Spreadsheet*"
		isFinanceMenu = true
	case "menu_kategori":
		title = "🗂️ KELOLA KATEGORI"
		desc = "Atur dan kelompokkan pos keuangan Anda:\n\n" +
			"• 🍔 Makanan & Minuman\n" +
			"• 🚗 Transportasi & Bensin\n" +
			"• 🏠 Tagihan Rumah & Utilitas\n" +
			"• 🛍️ Belanja & Kebutuhan Harian"
		isFinanceMenu = true
	case "menu_anggaran":
		title = "🎯 ANGGARAN & BUDGETING"
		desc = "Tentukan batas pengeluaran bulanan agar keuangan tetap sehat. Dapatkan alert saat pengeluaran menyentuh 80% dari batas anggaran."
		isFinanceMenu = true
	case "menu_pengingat":
		title = "🔔 PENGINGAT & NOTIFIKASI"
		desc = "Atur pengingat otomatis:\n\n" +
			"• ⏰ *Jadwal Catat Rutin* (setiap malam)\n" +
			"• 📅 *Jatuh Tempo Tagihan* (WiFi, Listrik, Cicilan)"
	case "menu_catatan":
		title = "📝 CATATAN & TO-DO LIST"
		desc = "Simpan ide cepat, daftar belanjaan, atau tugas harian Anda bersama Piko.\n\n" +
			"• 📋 Daftar tugas harian\n" +
			"• 📌 Catatan penting / memo cepat"
	case "menu_pengaturan":
		title = "⚙️ PENGATURAN AKUN"
		desc = "Sesuaikan preferensi akun Piko Anda:\n\n" +
			"• 💱 Mata uang utama (IDR, USD)\n" +
			"• 🕒 Zona waktu lokal (WIB, WITA, WIT)\n" +
			"• 🔄 Backup & Reset Data"
	case "menu_help":
		title = "📖 PANDUAN PENGGUNAAN"
		desc = "Perintah cepat yang dapat Anda gunakan kapan saja:\n\n" +
			"• `/menu` - Buka dashboard utama\n" +
			"• `/tools` - Kumpulan alat pintar\n" +
			"• `/saldo` - Cek saldo aktif Anda\n" +
			"• `/setsaldo <nominal>` - Atur ulang saldo\n" +
			"• `/tambahsaldo <nominal>` - Tambah saldo\n" +
			"• `/stats` - Cek info & statistik database\n" +
			"• `/ping` - Periksa koneksi bot"
	default:
		title = "📋 MENU PIKO"
		desc = "Menu ini akan segera hadir pada pembaruan mendatang."
	}

	content := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  *%s*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"%s\n\n"+
			"⏳ _Status: Tampilan interface aktif (fitur logika data sedang disiapkan)._",
		title, desc,
	)

	var backMarkup tgbotapi.InlineKeyboardMarkup
	if isFinanceMenu {
		backMarkup = GetBackToKeuanganInlineKeyboard()
	} else {
		backMarkup = GetBackToMenuInlineKeyboard()
	}

	return b.editCallbackMessage(query, content, &backMarkup)
}
