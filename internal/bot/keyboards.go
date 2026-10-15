package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetMainMenuInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Logbook KP", "menu_logbook"),
			tgbotapi.NewInlineKeyboardButtonData("💵 Keuangan", "menu_keuangan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛠️ Menu Tools", "menu_tools"),
			tgbotapi.NewInlineKeyboardButtonData("🔔 Pengingat", "menu_pengingat"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Catatan / To-Do", "menu_catatan"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Pengaturan", "menu_pengaturan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Panduan & Bantuan", "menu_help"),
		),
	)
}

func GetLogbookConfirmInlineKeyboard(hasPhoto, hasPDF bool, jamMulai, jamSelesai string) tgbotapi.InlineKeyboardMarkup {
	if jamMulai == "" {
		jamMulai = "08.00"
	}
	if jamSelesai == "" {
		jamSelesai = "16.00"
	}

	submitText := "🚀 Simpan ke MIS PENS"
	if hasPhoto && hasPDF {
		submitText = "🚀 Simpan Logbook, Foto & PDF"
	} else if hasPhoto {
		submitText = "🚀 Simpan Logbook & Foto"
	} else if hasPDF {
		submitText = "🚀 Simpan Logbook & PDF"
	}

	var attachmentButtons []tgbotapi.InlineKeyboardButton
	if hasPhoto {
		attachmentButtons = append(attachmentButtons, tgbotapi.NewInlineKeyboardButtonData("🗑️ Hapus Foto", "logbook_remove_photo"))
	} else {
		attachmentButtons = append(attachmentButtons, tgbotapi.NewInlineKeyboardButtonData("📸 +Foto", "logbook_prompt_photo"))
	}

	if hasPDF {
		attachmentButtons = append(attachmentButtons, tgbotapi.NewInlineKeyboardButtonData("🗑️ Hapus PDF", "logbook_remove_pdf"))
	} else {
		attachmentButtons = append(attachmentButtons, tgbotapi.NewInlineKeyboardButtonData("📄 +PDF Progres", "logbook_prompt_pdf"))
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(submitText, "logbook_submit"),
		),
		attachmentButtons,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏰ Ubah Jam ("+jamMulai+" - "+jamSelesai+")", "logbook_change_time"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Tulis Ulang", "menu_logbook"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "logbook_cancel"),
		),
	)
}

func GetCancelLogbookInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "logbook_cancel"),
		),
	)
}

func GetKeuanganInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 Catat Transaksi", "menu_catat"),
			tgbotapi.NewInlineKeyboardButtonData("👛 Saldo", "menu_saldo"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📜 Riwayat Transaksi", "menu_riwayat"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Laporan Keuangan", "menu_laporan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗂️ Kategori", "menu_kategori"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Anggaran", "menu_anggaran"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Kembali ke Menu Utama", "menu_main"),
		),
	)
}

func GetSaldoInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Atur Ulang Saldo", "saldo_set"),
			tgbotapi.NewInlineKeyboardButtonData("💵 Tambah Saldo", "saldo_add"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Perbarui", "menu_saldo"),
			tgbotapi.NewInlineKeyboardButtonData("🔙 Menu Keuangan", "menu_keuangan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Kembali ke Menu Utama", "menu_main"),
		),
	)
}

func GetCancelSaldoInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "saldo_cancel"),
		),
	)
}

func GetBackToKeuanganInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Menu Keuangan", "menu_keuangan"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "menu_main"),
		),
	)
}

func GetBackToMenuInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Kembali ke Menu Utama", "menu_main"),
		),
	)
}

func GetToolsInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔲 QR Code Generator", "tools_qr"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 File Converter", "tools_converter"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗜️ Compress PDF & Foto", "tools_compress"),
			tgbotapi.NewInlineKeyboardButtonData("🔗 URL Tools", "tools_url"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌐 Translator", "tools_translate"),
			tgbotapi.NewInlineKeyboardButtonData("🖼️ Image Metadata", "tools_imagemeta"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Kembali ke Menu Utama", "menu_main"),
		),
	)
}

func GetURLToolsInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✂️ Buat Shortlink", "tools_url_short"),
			tgbotapi.NewInlineKeyboardButtonData("🔀 Cek Redirect", "tools_url_redirect"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Cek Metadata URL", "tools_url_meta"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Menu Tools", "menu_tools"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "menu_main"),
		),
	)
}

func GetTranslatorInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇬🇧 English", "tools_tr_en"),
			tgbotapi.NewInlineKeyboardButtonData("🇮🇩 Indonesia", "tools_tr_id"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇯🇵 Japanese", "tools_tr_ja"),
			tgbotapi.NewInlineKeyboardButtonData("🇸🇦 Arabic", "tools_tr_ar"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Menu Tools", "menu_tools"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "menu_main"),
		),
	)
}

func GetCancelToolsInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "tools_cancel"),
		),
	)
}

func GetBackToToolsInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛠️ Menu Tools", "tools_back_to_tools"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "tools_back_to_main"),
		),
	)
}

func GetBackToURLToolsInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Menu URL Tools", "tools_back_to_url"),
			tgbotapi.NewInlineKeyboardButtonData("🛠️ Menu Tools", "tools_back_to_tools"),
		),
	)
}
