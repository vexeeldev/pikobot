package tools

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestGenerateQRCode(t *testing.T) {
	pngBytes, err := GenerateQRCode("https://example.com", 256)
	if err != nil {
		t.Fatalf("GenerateQRCode failed: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatalf("GenerateQRCode returned empty bytes")
	}

	_, err = GenerateQRCode("   ", 256)
	if err == nil {
		t.Errorf("expected error for empty content, got nil")
	}
}

func TestParseConvertArgs(t *testing.T) {
	target, hint := ParseConvertArgs("png")
	if target != "png" || hint != "" {
		t.Errorf("unexpected parse result: target=%s, hint=%s", target, hint)
	}

	target, hint = ParseConvertArgs("gambar.webp png")
	if target != "png" || hint != "gambar.webp" {
		t.Errorf("unexpected parse result: target=%s, hint=%s", target, hint)
	}

	target, hint = ParseConvertArgs(".jpg")
	if target != "jpg" {
		t.Errorf("expected jpg, got %s", target)
	}
}

func TestConvertImage(t *testing.T) {

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatalf("png encode failed: %v", err)
	}

	jpgBytes, outName, mime, err := ConvertFile(pngBuf.Bytes(), "test.png", "jpg")
	if err != nil {
		t.Fatalf("convert PNG to JPG failed: %v", err)
	}
	if len(jpgBytes) == 0 || outName != "test.jpg" || mime != "image/jpeg" {
		t.Errorf("unexpected conversion output: name=%s, mime=%s", outName, mime)
	}

	bmpBytes, outName, mime, err := ConvertFile(pngBuf.Bytes(), "test.png", "bmp")
	if err != nil {
		t.Fatalf("convert PNG to BMP failed: %v", err)
	}
	if len(bmpBytes) == 0 || outName != "test.bmp" || mime != "image/bmp" {
		t.Errorf("unexpected conversion output: name=%s, mime=%s", outName, mime)
	}
}

func TestConvertDocuments(t *testing.T) {

	jsonData := []byte(`[{"name":"Budi","age":"25"},{"name":"Siti","age":"22"}]`)
	csvBytes, name, mime, err := ConvertFile(jsonData, "users.json", "csv")
	if err != nil {
		t.Fatalf("JSON to CSV failed: %v", err)
	}
	if !strings.Contains(string(csvBytes), "Budi") || name != "users.csv" || mime != "text/csv" {
		t.Errorf("unexpected CSV output: %s", string(csvBytes))
	}

	jsonResult, name, mime, err := ConvertFile(csvBytes, "users.csv", "json")
	if err != nil {
		t.Fatalf("CSV to JSON failed: %v", err)
	}
	if !strings.Contains(string(jsonResult), "Budi") || name != "users.json" || mime != "application/json" {
		t.Errorf("unexpected JSON output: %s", string(jsonResult))
	}

	mdData := []byte("# Judul Dokumen\n- Poin satu\n- Poin dua")
	htmlBytes, name, mime, err := ConvertFile(mdData, "doc.md", "html")
	if err != nil {
		t.Fatalf("MD to HTML failed: %v", err)
	}
	if !strings.Contains(string(htmlBytes), "<h1>Judul Dokumen</h1>") || name != "doc.html" || mime != "text/html" {
		t.Errorf("unexpected HTML output: %s", string(htmlBytes))
	}
}

func TestURLNormalize(t *testing.T) {
	if NormalizeURL("google.com") != "https://google.com" {
		t.Errorf("expected https://google.com")
	}
	if NormalizeURL("http://test.com") != "http://test.com" {
		t.Errorf("expected http://test.com")
	}
}

func TestTranslatorLanguageCodes(t *testing.T) {
	if NormalizeLanguageCode("en") != "en" {
		t.Errorf("expected en")
	}
	if NormalizeLanguageCode("jepang") != "ja" {
		t.Errorf("expected ja for jepang")
	}
	if NormalizeLanguageCode("indonesia") != "id" {
		t.Errorf("expected id for indonesia")
	}
	if NormalizeLanguageCode("mandarin") != "zh-CN" {
		t.Errorf("expected zh-CN for mandarin")
	}

	label := GetLanguageName("en")
	if !strings.Contains(label, "English") {
		t.Errorf("expected English in label, got %s", label)
	}
}

func TestImageMetadata(t *testing.T) {

	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode failed: %v", err)
	}

	meta, err := ExtractImageMetadata(buf.Bytes(), "photo.png")
	if err != nil {
		t.Fatalf("ExtractImageMetadata failed: %v", err)
	}
	if meta.Width != 800 || meta.Height != 600 {
		t.Errorf("expected 800x600, got %dx%d", meta.Width, meta.Height)
	}
	if meta.Format != "PNG" {
		t.Errorf("expected PNG, got %s", meta.Format)
	}
	if !strings.Contains(meta.AspectRatio, "4:3") {
		t.Errorf("expected 4:3 ratio, got %s", meta.AspectRatio)
	}
}

func TestCompressImage(t *testing.T) {

	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	for x := 0; x < 500; x++ {
		for y := 0; y < 500; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode failed: %v", err)
	}

	opts := CompressImageOptions{
		Quality:   60,
		MaxDim:    400,
		ForceJPEG: true,
	}

	res, err := CompressImage(buf.Bytes(), "photo.png", opts)
	if err != nil {
		t.Fatalf("CompressImage failed: %v", err)
	}

	if res.CompressedSize <= 0 {
		t.Errorf("expected compressed size > 0, got %d", res.CompressedSize)
	}
	if res.MimeType != "image/jpeg" {
		t.Errorf("expected image/jpeg, got %s", res.MimeType)
	}
}
