package bot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pikobot/internal/tools"
)

func formatToolsMenuText(user *tgbotapi.User) string {
	name := user.FirstName
	if name == "" {
		name = "Sobat Piko"
	}

	return fmt.Sprintf(
		"🛠️ *MENU TOOLS & UTILITAS CERDAS*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"Halo, *%s*! Pilih utilitas pintar yang ingin Anda gunakan:\n\n"+
			"1️⃣ *QR Code Generator*\n"+
			"└ Buat gambar QR Code dari teks atau link web.\n\n"+
			"2️⃣ *File Converter*\n"+
			"└ Konversi format gambar (WEBP, PNG, JPG, BMP) & dokumen teks.\n\n"+
			"3️⃣ *Compress PDF & Foto*\n"+
			"└ Perkecil ukuran file dokumen PDF atau foto agar hemat memori.\n\n"+
			"4️⃣ *URL Tools*\n"+
			"└ Buat shortlink, periksa jejak redirect, dan cek metadata web.\n\n"+
			"5️⃣ *Translator*\n"+
			"└ Terjemahkan teks multibahasa secara cepat dan akurat.\n\n"+
			"6️⃣ *Image Metadata*\n"+
			"└ Cek resolusi, EXIF kamera, ukuran file, & info detail foto.\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"👇 *PILIH ALAT DI BAWAH INI:*",
		name,
	)
}

func (b *Bot) SendToolsMenu(chatID int64, user *tgbotapi.User) error {
	b.ClearUserState(user.ID)
	text := formatToolsMenuText(user)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetToolsInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) PromptQR(chatID int64, from *tgbotapi.User) error {
	b.SetUserState(from.ID, "waiting_qr_text")
	prompt := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  🔲 *QR CODE GENERATOR*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"Kirim teks, nomor telepon, atau URL yang ingin Anda ubah menjadi QR Code.\n\n" +
		"💡 *Contoh:*\n" +
		"• `https://google.com`\n" +
		"• `WiFi:S:MyNetwork;T:WPA;P:MyPassword;;`\n" +
		"• `Halo dari Piko Bot!`\n\n" +
		"Silakan kirimkan teks atau tautan Anda sekarang:"

	msg := tgbotapi.NewMessage(chatID, prompt)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) ExecuteGenerateQR(chatID int64, text string) error {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Teks untuk QR Code tidak boleh kosong.")
		_, err := b.API.Send(msg)
		return err
	}

	qrBytes, err := tools.GenerateQRCode(trimmed, 512)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal membuat QR Code: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	displayContent := trimmed
	if len(displayContent) > 100 {
		displayContent = displayContent[:97] + "..."
	}

	caption := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  ✅ *QR CODE BERHASIL DIBUAT*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"📋 *Isi / Konten:*\n`%s`\n\n"+
			"📐 *Ukuran:* `512 × 512 px` (PNG High Quality)\n"+
			"💡 _Simpan atau bagikan gambar ini ke rekan Anda._",
		displayContent,
	)

	photoBytes := tgbotapi.FileBytes{
		Name:  "qrcode.png",
		Bytes: qrBytes,
	}

	photo := tgbotapi.NewPhoto(chatID, photoBytes)
	photo.Caption = caption
	photo.ParseMode = "Markdown"
	photo.ReplyMarkup = GetBackToToolsInlineKeyboard()

	_, err = b.API.Send(photo)
	return err
}

func (b *Bot) PromptConverter(chatID int64, from *tgbotapi.User) error {
	b.SetUserState(from.ID, "waiting_convert_file")
	text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  🔄 *FILE CONVERTER*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"Konversi format gambar, dokumen, atau file teks dengan mudah.\n\n" +
		"📁 *Format yang didukung:*\n" +
		"• *Gambar:* WEBP, PNG, JPG/JPEG, BMP, GIF\n" +
		"• *Dokumen/Data:* JSON, CSV, HTML, TXT\n\n" +
		"Silakan kirimkan file atau foto yang ingin Anda ubah formatnya sekarang via chat ini:"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) ExecuteConvertFile(chatID int64, fileID string, origFilename string, targetFormat string) error {
	targetFormat = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(targetFormat), "."))
	if targetFormat == "" {
		targetFormat = "png"
	}

	loadingMsg, _ := b.API.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("⏳ *Sedang mengunduh & mengonversi file ke .%s...*", targetFormat)))

	fileURL, err := b.API.GetFileDirectURL(fileID)
	if err != nil {
		if loadingMsg.MessageID != 0 {
			b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal mengakses file dari server Telegram: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		if loadingMsg.MessageID != 0 {
			b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal mengunduh file: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if loadingMsg.MessageID != 0 {
			b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal membaca data file: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	convertedData, outFilename, _, convErr := tools.ConvertFile(data, origFilename, targetFormat)
	if loadingMsg.MessageID != 0 {
		b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
	}

	if convErr != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ *Konversi Gagal:*\n%v", convErr))
		msg.ParseMode = "Markdown"
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	caption := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  ✅ *FILE BERHASIL DIKONVERSI*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"📁 *Nama File:* `%s`\n"+
			"📦 *Ukuran:* `%s`\n"+
			"🔄 *Format Tujuan:* `%s`\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"File siap diunduh di atas 👆",
		outFilename, tools.FormatFileSize(int64(len(convertedData))), strings.ToUpper(targetFormat),
	)

	docFile := tgbotapi.FileBytes{
		Name:  outFilename,
		Bytes: convertedData,
	}

	doc := tgbotapi.NewDocument(chatID, docFile)
	doc.Caption = caption
	doc.ParseMode = "Markdown"
	doc.ReplyMarkup = GetBackToToolsInlineKeyboard()

	_, err = b.API.Send(doc)
	return err
}

func (b *Bot) PromptCompress(chatID int64, from *tgbotapi.User) error {
	b.SetUserState(from.ID, "waiting_compress_file")
	text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  🗜️ *KOMPRES PDF & GAMBAR*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"Perkecil ukuran file PDF atau foto agar lebih hemat memori & mudah diunggah.\n\n" +
		"📋 *Format yang didukung:*\n" +
		"• *Dokumen:* File `.PDF` (Optimasi stream & objek)\n" +
		"• *Gambar:* `.JPG`, `.PNG`, `.WEBP`, `.BMP`\n\n" +
		"Silakan kirimkan file dokumen PDF atau foto yang ingin Anda perkecil ukurannya sekarang via chat ini:"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) ExecuteCompressFile(chatID int64, fileID string, origFilename string, targetKB int) error {
	if origFilename == "" {
		origFilename = "file"
	}

	loadingMsg, _ := b.API.Send(tgbotapi.NewMessage(chatID, "⏳ *Sedang mengunduh & mengompresi file...*"))

	fileURL, err := b.API.GetFileDirectURL(fileID)
	if err != nil {
		if loadingMsg.MessageID != 0 {
			b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal mengakses file dari server Telegram: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		if loadingMsg.MessageID != 0 {
			b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal mengunduh file: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if loadingMsg.MessageID != 0 {
			b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal membaca data file: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	var result *tools.CompressionResult
	var compErr error

	isPDF := strings.HasSuffix(strings.ToLower(origFilename), ".pdf")
	if isPDF {
		result, compErr = tools.CompressPDF(ctx, data, origFilename)
	} else {
		opts := tools.CompressImageOptions{
			Quality:   70,
			MaxDim:    1920,
			TargetKB:  targetKB,
			ForceJPEG: true,
		}
		result, compErr = tools.CompressImage(data, origFilename, opts)
	}

	if loadingMsg.MessageID != 0 {
		b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
	}

	if compErr != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ *Kompresi Gagal:*\n%v", compErr))
		msg.ParseMode = "Markdown"
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	caption := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  ✅ *KOMPRESI BERHASIL*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"📁 *Nama File:* `%s`\n"+
			"📦 *Ukuran Awal:* `%s`\n"+
			"🗜️ *Ukuran Baru:* `%s`\n"+
			"🎉 *Hemat Ukuran:* `%.1f%%` (`%s`)\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"File siap diunduh di atas 👆",
		result.Filename, result.OriginalStr, result.CompressedStr, result.SavedPercent, tools.FormatFileSize(result.SavedBytes),
	)

	docFile := tgbotapi.FileBytes{
		Name:  result.Filename,
		Bytes: result.Data,
	}

	doc := tgbotapi.NewDocument(chatID, docFile)
	doc.Caption = caption
	doc.ParseMode = "Markdown"
	doc.ReplyMarkup = GetBackToToolsInlineKeyboard()

	_, err = b.API.Send(doc)
	return err
}

func (b *Bot) PromptURLTools(chatID int64, from *tgbotapi.User) error {
	b.ClearUserState(from.ID)
	text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  🔗 *URL TOOLS & NETWORK UTILITY*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"Kelola dan telusuri informasi tautan URL web secara mendalam:\n\n" +
		"1️⃣ *Shortlink Generator*\n" +
		"└ Perpendek URL panjang menjadi tautan ringkas dan rapi.\n\n" +
		"2️⃣ *Redirect Checker*\n" +
		"└ Periksa rantai pengalihan URL (HTTP 301, 302) hingga tujuan akhir.\n\n" +
		"3️⃣ *URL Metadata Inspector*\n" +
		"└ Ambil judul, deskripsi, OpenGraph, dan informasi web server.\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"👇 Pilih fitur yang ingin Anda gunakan di bawah:"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetURLToolsInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) ExecuteShortenURL(chatID int64, targetURL string) error {
	trimmed := strings.TrimSpace(targetURL)
	if trimmed == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Masukkan URL yang valid untuk diperpendek.")
		_, err := b.API.Send(msg)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	shortURL, err := tools.ShortenURL(ctx, trimmed)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal memendekkan URL: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	resultText := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  ✂️ *SHORT LINK BERHASIL DIBUAT*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"🔗 *Short URL:*\n`%s`\n\n"+
			"🌐 *Original URL:*\n`%s`\n\n"+
			"💡 _Klik link di atas untuk menyalin langsung._",
		shortURL, trimmed,
	)

	msg := tgbotapi.NewMessage(chatID, resultText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetBackToURLToolsInlineKeyboard()
	_, err = b.API.Send(msg)
	return err
}

func (b *Bot) ExecuteCheckRedirects(chatID int64, targetURL string) error {
	trimmed := strings.TrimSpace(targetURL)
	if trimmed == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Masukkan URL yang valid untuk dicek redirect-nya.")
		_, err := b.API.Send(msg)
		return err
	}

	loadingMsg, _ := b.API.Send(tgbotapi.NewMessage(chatID, "🔍 *Memeriksa rantai pengalihan URL...*"))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := tools.CheckRedirects(ctx, trimmed)
	if loadingMsg.MessageID != 0 {
		b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
	}

	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal memeriksa redirect URL: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	var sb strings.Builder
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("  🔀 *HASIL PEMERIKSAAN REDIRECT*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("🌐 *URL Awal:* `%s`\n\n", res.InitialURL))
	sb.WriteString("📋 *Rantai Pengalihan (Hops):*\n")

	for _, h := range res.Hops {
		icon := "➡️"
		if h.StatusCode >= 200 && h.StatusCode < 300 {
			icon = "🏁"
		} else if h.StatusCode >= 300 && h.StatusCode < 400 {
			icon = "🔀"
		} else if h.StatusCode >= 400 {
			icon = "⚠️"
		}
		sb.WriteString(fmt.Sprintf("%s *Tahap %d:* `%d %s`\n└ `%s`\n", icon, h.Step, h.StatusCode, h.StatusText, h.URL))
	}

	sb.WriteString(fmt.Sprintf("\n🎯 *URL Akhir:* `%s`\n", res.FinalURL))
	if res.IsRedirect {
		sb.WriteString(fmt.Sprintf("📊 *Total Pengalihan:* %d kali lompatan\n", res.TotalHops-1))
	} else {
		sb.WriteString("ℹ️ _Tidak ada pengalihan (langsung ke tujuan)._\n")
	}

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetBackToURLToolsInlineKeyboard()
	_, err = b.API.Send(msg)
	return err
}

func (b *Bot) ExecuteFetchURLMetadata(chatID int64, targetURL string) error {
	trimmed := strings.TrimSpace(targetURL)
	if trimmed == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Masukkan URL website yang valid untuk dicek metadatanya.")
		_, err := b.API.Send(msg)
		return err
	}

	loadingMsg, _ := b.API.Send(tgbotapi.NewMessage(chatID, "🔍 *Mengambil metadata URL...*"))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	meta, err := tools.FetchURLMetadata(ctx, trimmed)
	if loadingMsg.MessageID != 0 {
		b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
	}

	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal mengambil metadata URL: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	title := meta.Title
	if title == "" {
		title = "Tidak ada judul"
	}
	desc := meta.Description
	if desc == "" {
		desc = "Tidak ada deskripsi meta"
	} else if len(desc) > 300 {
		desc = desc[:297] + "..."
	}

	sizeStr := "-"
	if meta.ContentSize > 0 {
		sizeStr = tools.FormatFileSize(meta.ContentSize)
	}

	serverStr := meta.Server
	if serverStr == "" {
		serverStr = "-"
	}

	cType := meta.ContentType
	if cType == "" {
		cType = "-"
	}

	text := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  📋 *METADATA URL WEB*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"🌐 *URL:* `%s`\n\n"+
			"📌 *Judul Halaman:*\n*%s*\n\n"+
			"📝 *Deskripsi:*\n_%s_\n\n"+
			"⚙️ *Info Server & Protokol:*\n"+
			"├ 🚦 Status HTTP: `%d %s`\n"+
			"├ 📦 Ukuran Konten: `%s`\n"+
			"├ 📄 Tipe MIME: `%s`\n"+
			"└ 🖥️ Web Server: `%s`",
		meta.URL, title, desc, meta.StatusCode, http.StatusText(meta.StatusCode), sizeStr, cType, serverStr,
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetBackToURLToolsInlineKeyboard()
	_, err = b.API.Send(msg)
	return err
}

func (b *Bot) PromptTranslator(chatID int64, from *tgbotapi.User, defaultTarget string) error {
	if defaultTarget == "" {
		defaultTarget = "en"
	}
	b.SetUserState(from.ID, fmt.Sprintf("waiting_translate_text:%s", defaultTarget))

	text := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  🌐 *TRANSLATOR MULTI-BAHASA*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"Terjemahkan pesan secara akurat antarbahasa.\n\n"+
			"🎯 *Target Bahasa Aktif:* %s\n\n"+
			"Silakan kirimkan kalimat atau teks yang ingin Anda terjemahkan sekarang via chat ini.\n\n"+
			"👇 Atau pilih bahasa tujuan lain di bawah:",
		tools.GetLanguageName(defaultTarget),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetTranslatorInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) ExecuteTranslate(chatID int64, text, targetLang, sourceLang string) error {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Masukkan teks yang ingin diterjemahkan.")
		_, err := b.API.Send(msg)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	res, err := tools.TranslateText(ctx, trimmed, targetLang, sourceLang)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal menerjemahkan teks: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	resultText := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  🌐 *HASIL PENERJEMAHAN*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"📤 *Sumber (%s):*\n\"%s\"\n\n"+
			"📥 *Hasil Terjemahan (%s):*\n\"%s\"\n"+
			"━━━━━━━━━━━━━━━━━━━━━━",
		res.SourceLangName, res.OriginalText, res.TargetLangName, res.TranslatedText,
	)

	msg := tgbotapi.NewMessage(chatID, resultText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetBackToToolsInlineKeyboard()
	_, err = b.API.Send(msg)
	return err
}

func (b *Bot) PromptImageMetadata(chatID int64, from *tgbotapi.User) error {
	b.SetUserState(from.ID, "waiting_imagemeta_photo")

	text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  🖼️ *IMAGE METADATA INSPECTOR*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"Kirimkan foto atau gambar (JPG, PNG, WEBP) untuk melihat informasi lengkapnya:\n\n" +
		"🔍 *Informasi yang akan ditampilkan:*\n" +
		"• 📐 Resolusi (Lebar × Tinggi) & Megapixels\n" +
		"• 📦 Ukuran file asli & format kompresi\n" +
		"• 📷 Data EXIF kamera, model HP, lensa, ISO, aperture\n" +
		"• 🗺️ Koordinat lokasi GPS (jika tersedia di file)\n\n" +
		"Kirim foto sekarang via chat ini:"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetCancelToolsInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) ExecuteImageMetadata(chatID int64, fileID string, origFilename string) error {
	loadingMsg, _ := b.API.Send(tgbotapi.NewMessage(chatID, "🔍 *Mengekstrak metadata gambar...*"))

	fileURL, err := b.API.GetFileDirectURL(fileID)
	if err != nil {
		if loadingMsg.MessageID != 0 {
			b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal mengakses gambar dari Telegram: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		if loadingMsg.MessageID != 0 {
			b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal mengunduh gambar: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if loadingMsg.MessageID != 0 {
			b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal membaca data gambar: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	meta, err := tools.ExtractImageMetadata(data, origFilename)
	if loadingMsg.MessageID != 0 {
		b.API.Send(tgbotapi.NewDeleteMessage(chatID, loadingMsg.MessageID))
	}

	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Gagal membaca metadata: %v", err))
		_, sendErr := b.API.Send(msg)
		return sendErr
	}

	var sb strings.Builder
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("  🖼️ *METADATA INFORMASI GAMBAR*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("📁 *Informasi File:*\n")
	sb.WriteString(fmt.Sprintf("├ 🏷️ *Nama File:* `%s`\n", meta.Filename))
	sb.WriteString(fmt.Sprintf("├ 📦 *Ukuran File:* `%s`\n", meta.FileSizeStr))
	colorModelStr := meta.ColorModel
	if colorModelStr == "" {
		colorModelStr = "Standard"
	}
	sb.WriteString(fmt.Sprintf("└ 🎨 *Format:* `%s (%s)`\n\n", meta.Format, colorModelStr))

	sb.WriteString("📐 *Dimensi & Resolusi:*\n")
	if meta.Width > 0 && meta.Height > 0 {
		sb.WriteString(fmt.Sprintf("├ 📏 *Dimensi:* `%d × %d px`\n", meta.Width, meta.Height))
		sb.WriteString(fmt.Sprintf("├ 🔍 *Resolusi:* `%.2f MP (Megapixels)`\n", meta.Megapixels))
		sb.WriteString(fmt.Sprintf("└ 📐 *Aspek Rasio:* `%s`\n\n", meta.AspectRatio))
	} else {
		sb.WriteString("└ ⚠️ _Dimensi tidak dapat dibaca dari header file._\n\n")
	}

	if meta.HasEXIF {
		sb.WriteString("📷 *Data Kamera & EXIF:*\n")
		deviceStr := meta.CameraModel
		if meta.CameraMake != "" {
			deviceStr = meta.CameraMake + " " + meta.CameraModel
		}
		if deviceStr == "" {
			deviceStr = "-"
		}
		sb.WriteString(fmt.Sprintf("├ 📱 *Perangkat:* `%s`\n", deviceStr))

		if meta.DateTime != "" {
			sb.WriteString(fmt.Sprintf("├ 🗓️ *Waktu Jepret:* `%s`\n", meta.DateTime))
		}

		var cameraSettings []string
		if meta.FNumber != "" {
			cameraSettings = append(cameraSettings, meta.FNumber)
		}
		if meta.ExposureTime != "" {
			cameraSettings = append(cameraSettings, meta.ExposureTime)
		}
		if meta.ISO != "" {
			cameraSettings = append(cameraSettings, meta.ISO)
		}
		if meta.FocalLength != "" {
			cameraSettings = append(cameraSettings, meta.FocalLength)
		}
		if len(cameraSettings) > 0 {
			sb.WriteString(fmt.Sprintf("├ ⚙️ *Setting:* `%s`\n", strings.Join(cameraSettings, " | ")))
		}

		if meta.Software != "" {
			sb.WriteString(fmt.Sprintf("├ 💻 *Software:* `%s`\n", meta.Software))
		}

		if meta.HasGPS {
			sb.WriteString(fmt.Sprintf("└ 🗺️ *Lokasi GPS:* `%.5f, %.5f` ([Buka Maps](%s))\n", meta.Latitude, meta.Longitude, meta.GoogleMapsURL))
		} else {
			sb.WriteString("└ 🗺️ *Lokasi GPS:* _Tidak tersimpan di file_\n")
		}
	} else {
		sb.WriteString("ℹ️ _Data EXIF kamera tidak ditemukan (biasanya otomatis dihapus oleh aplikasi chat/kompresi)._\n")
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━")

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetBackToToolsInlineKeyboard()
	_, err = b.API.Send(msg)
	return err
}
