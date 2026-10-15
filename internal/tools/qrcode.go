package tools

import (
	"fmt"
	"strings"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCode(content string, size int) ([]byte, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, fmt.Errorf("konten teks atau URL tidak boleh kosong")
	}

	if size <= 0 {
		size = 512
	} else if size > 2048 {
		size = 2048
	}

	pngBytes, err := qrcode.Encode(trimmed, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat QR Code: %w", err)
	}

	return pngBytes, nil
}
