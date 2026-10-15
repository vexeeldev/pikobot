package bot

import (
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestToolsKeyboards(t *testing.T) {
	mainKb := GetMainMenuInlineKeyboard()
	foundTools := false
	for _, row := range mainKb.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData != nil && *btn.CallbackData == "menu_tools" {
				foundTools = true
				break
			}
		}
	}
	if !foundTools {
		t.Errorf("GetMainMenuInlineKeyboard does not contain menu_tools button")
	}

	toolsKb := GetToolsInlineKeyboard()
	expectedCallbacks := map[string]bool{
		"tools_qr":        false,
		"tools_converter": false,
		"tools_compress":  false,
		"tools_url":       false,
		"tools_translate": false,
		"tools_imagemeta": false,
		"menu_main":       false,
	}

	for _, row := range toolsKb.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData != nil {
				if _, ok := expectedCallbacks[*btn.CallbackData]; ok {
					expectedCallbacks[*btn.CallbackData] = true
				}
			}
		}
	}

	for cb, found := range expectedCallbacks {
		if !found {
			t.Errorf("GetToolsInlineKeyboard is missing callback: %s", cb)
		}
	}
}

func TestFormatToolsMenuText(t *testing.T) {
	user := &tgbotapi.User{
		ID:        12345,
		FirstName: "PikoUser",
	}

	text := formatToolsMenuText(user)
	if !strings.Contains(text, "PikoUser") {
		t.Errorf("formatToolsMenuText should contain user first name")
	}
	if !strings.Contains(text, "QR Code Generator") {
		t.Errorf("formatToolsMenuText should contain QR Code Generator")
	}
	if !strings.Contains(text, "File Converter") {
		t.Errorf("formatToolsMenuText should contain File Converter")
	}
	if !strings.Contains(text, "Compress") {
		t.Errorf("formatToolsMenuText should contain Compress")
	}
	if !strings.Contains(text, "URL Tools") {
		t.Errorf("formatToolsMenuText should contain URL Tools")
	}
	if !strings.Contains(text, "Translator") {
		t.Errorf("formatToolsMenuText should contain Translator")
	}
	if !strings.Contains(text, "Image Metadata") {
		t.Errorf("formatToolsMenuText should contain Image Metadata")
	}
}

func TestEditCallbackMessageNilSafety(t *testing.T) {
	b := &Bot{}
	if err := b.editCallbackMessage(nil, "test", nil); err != nil {
		t.Errorf("expected nil error for nil query, got %v", err)
	}
	if err := b.editCallbackMessage(&tgbotapi.CallbackQuery{}, "test", nil); err != nil {
		t.Errorf("expected nil error for nil message in query, got %v", err)
	}
}

func TestIsToolResultMessage(t *testing.T) {
	if isToolResultMessage(nil) {
		t.Errorf("expected false for nil message")
	}

	docMsg := &tgbotapi.Message{
		Document: &tgbotapi.Document{},
	}
	if !isToolResultMessage(docMsg) {
		t.Errorf("expected true for document message")
	}

	compMsg := &tgbotapi.Message{
		Caption: "✅ *KOMPRESI BERHASIL*\nFile siap diunduh",
	}
	if !isToolResultMessage(compMsg) {
		t.Errorf("expected true for compression result caption")
	}

	menuMsg := &tgbotapi.Message{
		Caption: "✨ *PIKO ASSISTANT* ✨\nDashboard",
	}
	if isToolResultMessage(menuMsg) {
		t.Errorf("expected false for dashboard menu")
	}
}

func TestBackToToolsKeyboard(t *testing.T) {
	kb := GetBackToToolsInlineKeyboard()
	foundTools := false
	foundMain := false
	for _, row := range kb.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData != nil && *btn.CallbackData == "tools_back_to_tools" {
				foundTools = true
			}
			if btn.CallbackData != nil && *btn.CallbackData == "tools_back_to_main" {
				foundMain = true
			}
		}
	}
	if !foundTools || !foundMain {
		t.Errorf("GetBackToToolsInlineKeyboard missing expected callbacks")
	}
}
