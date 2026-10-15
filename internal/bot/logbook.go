package bot

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pikobot/internal/mis"
)

var reTimeRange = regexp.MustCompile(`^(\d{1,2}[:.]\d{2})\s*(?:[-–]|s/?d|sampai|\s)\s*(\d{1,2}[:.]\d{2})$`)

func (b *Bot) PromptLogbookKegiatan(chatID int64, from *tgbotapi.User) error {
	b.SetUserState(from.ID, "waiting_logbook_kegiatan")

	todayStr := time.Now().Format("02-01-2006")
	defaultJam := "08.00 - 16.00"
	if b.MISClient != nil {
		cfg := b.MISClient.GetConfig()
		if cfg.JamMulai != "" && cfg.JamSelesai != "" {
			defaultJam = fmt.Sprintf("%s - %s", cfg.JamMulai, cfg.JamSelesai)
		}
	}

	text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  📝 *LOGBOOK HARIAN KP*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"📅 *Tanggal* : `" + todayStr + "`\n" +
		"⏰ *Jam Kerja* : `" + defaultJam + "` _(Otomatis)_\n\n" +
		"Silakan ketik kegiatan atau pekerjaan yang kamu lakukan hari ini:\n\n" +
		"💡 *Contoh:* `mengecek performa server dan memperbaiki bug alert whatsapp`"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetCancelLogbookInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) ShowLogbookPreview(chatID int64, from *tgbotapi.User, entry *mis.LogbookEntry) error {
	b.SetLogbookDraft(from.ID, entry)
	b.SetUserState(from.ID, "confirm_logbook")

	hasPhoto := len(entry.FotoBytes) > 0
	hasPDF := len(entry.PDFBytes) > 0

	photoStatus := "📸 *Foto Kegiatan* : _Belum dilampirkan_"
	if hasPhoto {
		photoStatus = fmt.Sprintf("📸 *Foto Kegiatan* : Terlampir ✅ (`%s`)", entry.FotoFilename)
	}

	pdfStatus := "📄 *Laporan PDF*   : _Belum dilampirkan_"
	if hasPDF {
		pdfStatus = fmt.Sprintf("📄 *Laporan PDF*   : Terlampir ✅ (`%s`)", entry.PDFFilename)
	}

	text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  📋 *KONFIRMASI LOGBOOK KP*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"📅 *Tanggal* : `" + entry.Tanggal + "`\n" +
		"⏰ *Jam Kerja* : `" + entry.JamMulai + " - " + entry.JamSelesai + "`\n" +
		photoStatus + "\n" +
		pdfStatus + "\n\n" +
		"📝 *Kegiatan*:\n" +
		"\"" + entry.Kegiatan + "\"\n\n" +
		"💡 _Kirim foto dokumentasi (JPG) atau dokumen laporan (PDF) langsung ke chat untuk melampirkannya._"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetLogbookConfirmInlineKeyboard(hasPhoto, hasPDF, entry.JamMulai, entry.JamSelesai)
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) HandleSubmitLogbook(query *tgbotapi.CallbackQuery) error {
	draft := b.GetLogbookDraft(query.From.ID)
	if draft == nil {
		ans := tgbotapi.NewCallback(query.ID, "Draft logbook kosong. Silakan buka menu Logbook untuk mulai.")
		b.API.Request(ans)
		return nil
	}

	if b.MISClient == nil {
		ans := tgbotapi.NewCallback(query.ID, "Klien MIS belum diinisialisasi.")
		b.API.Request(ans)
		return nil
	}

	ans := tgbotapi.NewCallback(query.ID, "Sedang mengirim ke MIS PENS...")
	b.API.Request(ans)

	statusMsg := "⏳ *Sedang menghubungkan ke MIS PENS & menyimpan logbook...*\nMohon tunggu beberapa detik."
	if len(draft.FotoBytes) > 0 && len(draft.PDFBytes) > 0 {
		statusMsg = "⏳ *Sedang menyimpan logbook, foto kegiatan & laporan PDF ke MIS PENS...*\nMohon tunggu beberapa detik."
	} else if len(draft.FotoBytes) > 0 {
		statusMsg = "⏳ *Sedang menyimpan logbook & mengunggah foto kegiatan ke MIS PENS...*\nMohon tunggu beberapa detik."
	} else if len(draft.PDFBytes) > 0 {
		statusMsg = "⏳ *Sedang menyimpan logbook & mengunggah laporan PDF ke MIS PENS...*\nMohon tunggu beberapa detik."
	}
	_ = b.editCallbackMessage(query, statusMsg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	err := b.MISClient.SubmitLogbook(ctx, draft)
	if err != nil {
		errorText := fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━\n"+
			"  ⚠️ *GAGAL MENYIMPAN LOGBOOK*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"MIS PENS memberikan respons error:\n`%v`\n\n"+
			"💡 *Tips Perbaikan:*\n"+
			"1. Pastikan `MIS_USERNAME` dan `MIS_PASSWORD` di `.env` sudah benar.\n"+
			"2. Periksa apakah jaringan / server MIS PENS sedang maintenance.", err)

		errorMarkup := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔄 Coba Lagi", "logbook_submit"),
				tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "logbook_cancel"),
			),
		)
		return b.editCallbackMessage(query, errorText, &errorMarkup)
	}

	photoUploaded := false
	var photoErr error
	if len(draft.FotoBytes) > 0 {
		photoErr = b.MISClient.UploadFotoKegiatan(ctx, draft)
		if photoErr == nil {
			photoUploaded = true
		} else {
			log.Printf("[WARN] Failed to upload logbook photo: %v", photoErr)
		}
	}

	pdfUploaded := false
	var pdfErr error
	if len(draft.PDFBytes) > 0 {
		pdfErr = b.MISClient.UploadLaporanProgres(ctx, draft)
		if pdfErr == nil {
			pdfUploaded = true
		} else {
			log.Printf("[WARN] Failed to upload logbook PDF report: %v", pdfErr)
		}
	}

	b.ClearLogbookDraft(query.From.ID)
	b.ClearUserState(query.From.ID)

	var uploadNotes []string
	if photoUploaded {
		uploadNotes = append(uploadNotes, "📸 *Foto Kegiatan* : Berhasil diunggah ke MIS PENS! ✅")
	} else if photoErr != nil {
		uploadNotes = append(uploadNotes, fmt.Sprintf("⚠️ *Foto Kegiatan* : Gagal diunggah (`%v`)", photoErr))
	}

	if pdfUploaded {
		uploadNotes = append(uploadNotes, "📄 *Laporan PDF*   : Berhasil diunggah ke MIS PENS! ✅")
	} else if pdfErr != nil {
		uploadNotes = append(uploadNotes, fmt.Sprintf("⚠️ *Laporan PDF*   : Gagal diunggah (`%v`)", pdfErr))
	}

	notesStr := ""
	if len(uploadNotes) > 0 {
		notesStr = "\n" + strings.Join(uploadNotes, "\n")
	}

	successText := fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━\n"+
		"  ✅ *LOGBOOK RESMI TERSIMPAN!*\n"+
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
		"📅 *Tanggal* : `%s`\n"+
		"⏰ *Jam Kerja* : `%s - %s`\n"+
		"%s\n\n"+
		"📝 *Kegiatan*:\n\"%s\"\n\n"+
		"🎉 Semua data logbook kamu telah resmi masuk ke database MIS PENS!",
		draft.Tanggal, draft.JamMulai, draft.JamSelesai, notesStr, draft.Kegiatan)

	successMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "menu_main"),
		),
	)
	return b.editCallbackMessage(query, successText, &successMarkup)
}

func (b *Bot) PromptAttachPhoto(chatID int64, from *tgbotapi.User) error {
	b.SetUserState(from.ID, "waiting_logbook_photo")

	text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  📸 *LAMPIRKAN FOTO KEGIATAN*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"Silakan kirim foto dokumentasi kegiatan kerja praktek kamu sekarang.\n\n" +
		"💡 *Catatan:* Batas server MIS PENS adalah 500 KB (bot akan otomatis mengompres foto jika ukurannya besar)."

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetCancelLogbookInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) HandleRemovePhoto(query *tgbotapi.CallbackQuery) error {
	draft := b.GetLogbookDraft(query.From.ID)
	if draft != nil {
		draft.FotoBytes = nil
		draft.FotoFilename = ""
		b.SetLogbookDraft(query.From.ID, draft)
	}

	ans := tgbotapi.NewCallback(query.ID, "Foto berhasil dihapus dari draft.")
	b.API.Request(ans)

	if draft != nil {
		return b.ShowLogbookPreview(query.Message.Chat.ID, query.From, draft)
	}
	return nil
}

func (b *Bot) HandlePhoto(message *tgbotapi.Message) error {
	userID := message.From.ID
	if len(message.Photo) == 0 {
		return nil
	}

	userState := b.GetUserState(userID)
	captionLower := strings.ToLower(strings.TrimSpace(message.Caption))
	largestPhoto := message.Photo[len(message.Photo)-1]

	if userState == "waiting_imagemeta_photo" || strings.HasPrefix(captionLower, "/metadata") || strings.HasPrefix(captionLower, "/imgmeta") {
		b.ClearUserState(userID)
		return b.ExecuteImageMetadata(message.Chat.ID, largestPhoto.FileID, "photo.jpg")
	}

	if userState == "waiting_convert_file" || strings.HasPrefix(captionLower, "/convert") {
		b.ClearUserState(userID)
		targetFormat := "png"
		if strings.HasPrefix(captionLower, "/convert") {
			parts := strings.Fields(captionLower)
			if len(parts) > 1 {
				targetFormat = parts[len(parts)-1]
			}
		}
		return b.ExecuteConvertFile(message.Chat.ID, largestPhoto.FileID, "photo.jpg", targetFormat)
	}

	if userState == "waiting_compress_file" || strings.HasPrefix(captionLower, "/compress") {
		b.ClearUserState(userID)
		return b.ExecuteCompressFile(message.Chat.ID, largestPhoto.FileID, "photo.jpg", 0)
	}

	fileURL, err := b.API.GetFileDirectURL(largestPhoto.FileID)
	if err != nil {
		return fmt.Errorf("failed to get photo URL: %w", err)
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("failed to download photo: %w", err)
	}
	defer resp.Body.Close()

	photoData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read photo bytes: %w", err)
	}

	compressed := CompressJPEG(photoData, 480*1024)

	draft := b.GetLogbookDraft(userID)
	if draft == nil {
		kegiatan := strings.TrimSpace(message.Caption)
		if kegiatan == "" {
			kegiatan = "Mengerjakan kegiatan kerja praktek"
		}
		if b.MISClient != nil {
			draft = b.MISClient.GetConfig().DefaultLogbookEntry(kegiatan)
		} else {
			draft = &mis.LogbookEntry{
				Tanggal:      time.Now().Format("2006-01-02"),
				JamMulai:     "08.00",
				JamSelesai:   "16.00",
				Kegiatan:     kegiatan,
				SesuaiKuliah: 2,
				Minggu:       16,
			}
		}
	} else if message.Caption != "" && (draft.Kegiatan == "" || draft.Kegiatan == "Belum diisi") {
		draft.Kegiatan = message.Caption
	}

	draft.FotoBytes = compressed
	draft.FotoFilename = fmt.Sprintf("kegiatan_%s.jpg", time.Now().Format("20060102_150405"))
	b.SetLogbookDraft(userID, draft)

	infoMsg := tgbotapi.NewMessage(message.Chat.ID, "📸 *Foto dokumentasi kegiatan berhasil dilampirkan!*")
	infoMsg.ParseMode = "Markdown"
	_, _ = b.API.Send(infoMsg)

	return b.ShowLogbookPreview(message.Chat.ID, message.From, draft)
}

func (b *Bot) PromptChangeJam(chatID int64, from *tgbotapi.User) error {
	b.SetUserState(from.ID, "waiting_logbook_jam")

	text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  ⏰ *ATUR JAM KERJA LOGBOOK*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"Ketik jam mulai dan jam selesai dengan format:\n" +
		"`JJ.MM - JJ.MM`\n\n" +
		"💡 *Contoh:* `08.00 - 16.00` atau `09.00 - 17.00`"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetCancelLogbookInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func ParseJamInput(input string) (string, string, error) {
	trimmed := strings.TrimSpace(input)
	normalized := strings.ReplaceAll(trimmed, ":", ".")

	matches := reTimeRange.FindStringSubmatch(normalized)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("format tidak cocok. Gunakan contoh `08.00 - 16.00`")
	}

	mulai := formatTwoDigitsTime(matches[1])
	selesai := formatTwoDigitsTime(matches[2])
	return mulai, selesai, nil
}

func formatTwoDigitsTime(t string) string {
	parts := strings.Split(t, ".")
	if len(parts) == 2 {
		h := parts[0]
		m := parts[1]
		if len(h) == 1 {
			h = "0" + h
		}
		if len(m) == 1 {
			m = m + "0"
		}
		return h + "." + m
	}
	return t
}

func CompressJPEG(data []byte, maxBytes int) []byte {
	if len(data) <= maxBytes {
		return data
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}

	for quality := 85; quality >= 25; quality -= 15 {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err == nil {
			if buf.Len() <= maxBytes || quality == 25 {
				return buf.Bytes()
			}
		}
	}

	return data
}

func (b *Bot) HandleDocument(message *tgbotapi.Message) error {
	doc := message.Document
	if doc == nil {
		return nil
	}

	userID := message.From.ID
	userState := b.GetUserState(userID)
	captionLower := strings.ToLower(strings.TrimSpace(message.Caption))

	if userState == "waiting_imagemeta_photo" || strings.HasPrefix(captionLower, "/metadata") || strings.HasPrefix(captionLower, "/imgmeta") {
		b.ClearUserState(userID)
		return b.ExecuteImageMetadata(message.Chat.ID, doc.FileID, doc.FileName)
	}

	if userState == "waiting_convert_file" || strings.HasPrefix(captionLower, "/convert") {
		b.ClearUserState(userID)
		targetFormat := "png"
		if strings.HasPrefix(captionLower, "/convert") {
			parts := strings.Fields(captionLower)
			if len(parts) > 1 {
				targetFormat = parts[len(parts)-1]
			}
		}
		return b.ExecuteConvertFile(message.Chat.ID, doc.FileID, doc.FileName, targetFormat)
	}

	if userState == "waiting_compress_file" || strings.HasPrefix(captionLower, "/compress") {
		b.ClearUserState(userID)
		return b.ExecuteCompressFile(message.Chat.ID, doc.FileID, doc.FileName, 0)
	}

	fileNameLower := strings.ToLower(doc.FileName)
	mimeLower := strings.ToLower(doc.MimeType)
	isImage := strings.HasSuffix(fileNameLower, ".jpg") || strings.HasSuffix(fileNameLower, ".jpeg") || strings.HasSuffix(fileNameLower, ".png") || strings.HasPrefix(mimeLower, "image/")
	isPDF := strings.HasSuffix(fileNameLower, ".pdf") || strings.Contains(mimeLower, "pdf")

	if isImage || userState == "waiting_logbook_photo" {
		fileURL, err := b.API.GetFileDirectURL(doc.FileID)
		if err != nil {
			return fmt.Errorf("failed to get file download url: %w", err)
		}
		resp, err := http.Get(fileURL)
		if err != nil {
			return fmt.Errorf("failed to download photo document: %w", err)
		}
		defer resp.Body.Close()

		photoData, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read photo bytes: %w", err)
		}

		compressed := CompressJPEG(photoData, 480*1024)
		draft := b.GetLogbookDraft(message.From.ID)
		if draft == nil {
			kegiatan := strings.TrimSpace(message.Caption)
			if kegiatan == "" {
				kegiatan = "Mengerjakan kegiatan kerja praktek"
			}
			if b.MISClient != nil {
				draft = b.MISClient.GetConfig().DefaultLogbookEntry(kegiatan)
			} else {
				draft = &mis.LogbookEntry{
					Tanggal:      time.Now().Format("2006-01-02"),
					JamMulai:     "08.00",
					JamSelesai:   "16.00",
					Kegiatan:     kegiatan,
					SesuaiKuliah: 2,
					Minggu:       16,
				}
			}
		} else if message.Caption != "" && (draft.Kegiatan == "" || draft.Kegiatan == "Belum diisi") {
			draft.Kegiatan = message.Caption
		}

		draft.FotoBytes = compressed
		draft.FotoFilename = doc.FileName
		if !strings.HasSuffix(strings.ToLower(draft.FotoFilename), ".jpg") && !strings.HasSuffix(strings.ToLower(draft.FotoFilename), ".jpeg") {
			draft.FotoFilename = fmt.Sprintf("kegiatan_%s.jpg", time.Now().Format("20060102_150405"))
		}
		b.SetLogbookDraft(message.From.ID, draft)
		b.ClearUserState(message.From.ID)

		infoMsg := tgbotapi.NewMessage(message.Chat.ID, "📸 *Foto dokumentasi kegiatan berhasil dilampirkan!*")
		infoMsg.ParseMode = "Markdown"
		_, _ = b.API.Send(infoMsg)

		return b.ShowLogbookPreview(message.Chat.ID, message.From, draft)
	}

	if !isPDF {
		reply := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Format file tidak didukung. MIS PENS hanya menerima laporan progres dalam format *.PDF* atau foto kegiatan dalam format *.JPG/.PNG*.")
		_, err := b.API.Send(reply)
		return err
	}

	if doc.FileSize > 1024*1024 {
		sizeKB := doc.FileSize / 1024
		reply := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("⚠️ Ukuran file terlalu besar (%d KB). Batas maksimal server MIS PENS adalah *1 MB (1024 KB)*. Silakan kompres PDF Anda terlebih dahulu.", sizeKB))
		reply.ParseMode = "Markdown"
		_, err := b.API.Send(reply)
		return err
	}

	fileURL, err := b.API.GetFileDirectURL(doc.FileID)
	if err != nil {
		return fmt.Errorf("failed to get file download url: %w", err)
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("failed to download document: %w", err)
	}
	defer resp.Body.Close()

	pdfData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read document bytes: %w", err)
	}

	minggu := 16
	if b.MISClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if w, err := b.MISClient.FetchActiveWeek(ctx); err == nil && w > 0 {
			minggu = w
		}
		cancel()
	}

	draft := b.GetLogbookDraft(message.From.ID)
	if draft == nil {
		kegiatan := strings.TrimSpace(message.Caption)
		if kegiatan == "" {
			kegiatan = fmt.Sprintf("Mengerjakan kegiatan kerja praktek & progres laporan minggu ke-%d", minggu)
		}
		if b.MISClient != nil {
			draft = b.MISClient.GetConfig().DefaultLogbookEntry(kegiatan)
		} else {
			draft = &mis.LogbookEntry{
				Tanggal:      time.Now().Format("2006-01-02"),
				JamMulai:     "08.00",
				JamSelesai:   "16.00",
				Kegiatan:     kegiatan,
				SesuaiKuliah: 2,
				Minggu:       minggu,
			}
		}
	} else if message.Caption != "" && (draft.Kegiatan == "" || draft.Kegiatan == "Belum diisi") {
		draft.Kegiatan = message.Caption
	}

	draft.PDFBytes = pdfData
	draft.PDFFilename = doc.FileName
	draft.Minggu = minggu
	b.SetLogbookDraft(message.From.ID, draft)

	infoMsg := tgbotapi.NewMessage(message.Chat.ID, "📄 *Laporan progres PDF berhasil dilampirkan ke logbook!*")
	infoMsg.ParseMode = "Markdown"
	_, _ = b.API.Send(infoMsg)

	return b.ShowLogbookPreview(message.Chat.ID, message.From, draft)
}

func (b *Bot) PromptAttachPDF(chatID int64, from *tgbotapi.User) error {
	b.SetUserState(from.ID, "waiting_logbook_pdf")

	text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
		"  📄 *LAMPIRKAN LAPORAN PROGRES (PDF)*\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"Silakan kirimkan dokumen file laporan mingguan kamu (*.PDF*) ke chat ini.\n\n" +
		"📋 *Ketentuan MIS PENS:*\n" +
		"• Format file: *.PDF*\n" +
		"• Ukuran file maksimal: *1 MB*"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = GetCancelLogbookInlineKeyboard()
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) HandleRemovePDF(query *tgbotapi.CallbackQuery) error {
	draft := b.GetLogbookDraft(query.From.ID)
	if draft != nil {
		draft.PDFBytes = nil
		draft.PDFFilename = ""
		b.SetLogbookDraft(query.From.ID, draft)
	}

	ans := tgbotapi.NewCallback(query.ID, "Laporan PDF berhasil dihapus dari draft.")
	b.API.Request(ans)

	if draft != nil {
		return b.ShowLogbookPreview(query.Message.Chat.ID, query.From, draft)
	}
	return nil
}
