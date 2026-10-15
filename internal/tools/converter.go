package tools

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

func SupportedFormats() []string {
	return []string{
		"PNG", "JPG/JPEG", "BMP", "GIF", "JSON", "CSV", "HTML", "TXT",
	}
}

func ParseConvertArgs(args string) (targetFormat string, sourceHint string) {
	parts := strings.Fields(args)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return strings.ToLower(strings.TrimPrefix(parts[0], ".")), ""
	}

	p0 := strings.ToLower(strings.TrimPrefix(parts[0], "."))
	p1 := strings.ToLower(strings.TrimPrefix(parts[1], "."))

	if strings.Contains(parts[0], ".") && !strings.Contains(parts[1], ".") {
		return p1, parts[0]
	}
	if strings.Contains(parts[1], ".") && !strings.Contains(parts[0], ".") {
		return p0, parts[1]
	}

	return p1, parts[0]
}

func ConvertFile(data []byte, filename string, targetExt string) ([]byte, string, string, error) {
	targetExt = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(targetExt), "."))
	if targetExt == "" {
		return nil, "", "", fmt.Errorf("format tujuan konversi belum ditentukan (contoh: png, jpg, bmp, json)")
	}

	srcExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	if baseName == "" {
		baseName = "converted_file"
	}

	imageFormats := map[string]bool{
		"png": true, "jpg": true, "jpeg": true, "bmp": true, "gif": true, "webp": true,
	}

	if imageFormats[targetExt] {
		return convertImage(data, baseName, srcExt, targetExt)
	}

	switch targetExt {
	case "json":
		return convertToJSON(data, baseName, srcExt)
	case "csv":
		return convertToCSV(data, baseName, srcExt)
	case "html":
		return convertToHTML(data, baseName, srcExt)
	case "txt":
		return convertToTXT(data, baseName, srcExt)
	default:
		return nil, "", "", fmt.Errorf("format tujuan '.%s' belum didukung. Format didukung: png, jpg, bmp, gif, json, csv, html, txt", targetExt)
	}
}

func convertImage(data []byte, baseName, srcExt, targetExt string) ([]byte, string, string, error) {
	var img image.Image
	var err error

	img, _, err = image.Decode(bytes.NewReader(data))
	if err != nil {

		if webpImg, webpErr := webp.Decode(bytes.NewReader(data)); webpErr == nil {
			img = webpImg
			err = nil
		} else {

			if bmpImg, bmpErr := bmp.Decode(bytes.NewReader(data)); bmpErr == nil {
				img = bmpImg
				err = nil
			}
		}
	}

	if err != nil || img == nil {
		return nil, "", "", fmt.Errorf("gagal membaca format gambar sumber. Pastikan file adalah gambar valid (WEBP, PNG, JPG, GIF, BMP): %v", err)
	}

	var buf bytes.Buffer
	var mimeType string
	var outFilename string

	switch targetExt {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", "", fmt.Errorf("gagal encode ke PNG: %w", err)
		}
		mimeType = "image/png"
		outFilename = baseName + ".png"

	case "jpg", "jpeg":
		opts := &jpeg.Options{Quality: 92}
		if err := jpeg.Encode(&buf, img, opts); err != nil {
			return nil, "", "", fmt.Errorf("gagal encode ke JPEG: %w", err)
		}
		mimeType = "image/jpeg"
		outFilename = baseName + ".jpg"

	case "bmp":
		if err := bmp.Encode(&buf, img); err != nil {
			return nil, "", "", fmt.Errorf("gagal encode ke BMP: %w", err)
		}
		mimeType = "image/bmp"
		outFilename = baseName + ".bmp"

	case "gif":
		opts := &gif.Options{NumColors: 256}
		if err := gif.Encode(&buf, img, opts); err != nil {
			return nil, "", "", fmt.Errorf("gagal encode ke GIF: %w", err)
		}
		mimeType = "image/gif"
		outFilename = baseName + ".gif"

	default:
		return nil, "", "", fmt.Errorf("format output gambar '%s' tidak didukung", targetExt)
	}

	return buf.Bytes(), outFilename, mimeType, nil
}

func convertToJSON(data []byte, baseName, srcExt string) ([]byte, string, string, error) {
	if srcExt == "csv" {
		reader := csv.NewReader(bytes.NewReader(data))
		records, err := reader.ReadAll()
		if err != nil {
			return nil, "", "", fmt.Errorf("gagal membaca CSV: %w", err)
		}
		if len(records) == 0 {
			return nil, "", "", fmt.Errorf("file CSV kosong")
		}

		headers := records[0]
		var result []map[string]string
		for i := 1; i < len(records); i++ {
			row := records[i]
			item := make(map[string]string)
			for j, val := range row {
				colName := fmt.Sprintf("col_%d", j+1)
				if j < len(headers) && headers[j] != "" {
					colName = headers[j]
				}
				item[colName] = val
			}
			result = append(result, item)
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, "", "", fmt.Errorf("gagal membuat JSON: %w", err)
		}
		return jsonBytes, baseName + ".json", "application/json", nil
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, "", "", fmt.Errorf("data bukan format JSON atau CSV yang valid")
	}
	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return nil, "", "", err
	}
	return pretty, baseName + ".json", "application/json", nil
}

func convertToCSV(data []byte, baseName, srcExt string) ([]byte, string, string, error) {
	var list []map[string]interface{}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, "", "", fmt.Errorf("file harus berupa format JSON array of objects untuk diubah ke CSV: %v", err)
	}

	if len(list) == 0 {
		return nil, "", "", fmt.Errorf("JSON kosong")
	}

	var headers []string
	first := list[0]
	for k := range first {
		headers = append(headers, k)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write(headers); err != nil {
		return nil, "", "", err
	}

	for _, item := range list {
		var row []string
		for _, h := range headers {
			row = append(row, fmt.Sprintf("%v", item[h]))
		}
		if err := writer.Write(row); err != nil {
			return nil, "", "", err
		}
	}
	writer.Flush()

	return buf.Bytes(), baseName + ".csv", "text/csv", nil
}

func convertToHTML(data []byte, baseName, srcExt string) ([]byte, string, string, error) {
	content := string(data)
	lines := strings.Split(content, "\n")
	var htmlLines []string

	htmlLines = append(htmlLines, "<!DOCTYPE html>", "<html>", "<head>", "<meta charset=\"UTF-8\">", fmt.Sprintf("<title>%s</title>", baseName), "<style>body{font-family:system-ui,sans-serif;line-height:1.6;padding:2rem;max-width:800px;margin:auto;color:#222;background:#fcfcfc;}</style>", "</head>", "<body>")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			htmlLines = append(htmlLines, "<br/>")
		} else if strings.HasPrefix(trimmed, "### ") {
			htmlLines = append(htmlLines, fmt.Sprintf("<h3>%s</h3>", escapeHTML(trimmed[4:])))
		} else if strings.HasPrefix(trimmed, "## ") {
			htmlLines = append(htmlLines, fmt.Sprintf("<h2>%s</h2>", escapeHTML(trimmed[3:])))
		} else if strings.HasPrefix(trimmed, "# ") {
			htmlLines = append(htmlLines, fmt.Sprintf("<h1>%s</h1>", escapeHTML(trimmed[2:])))
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			htmlLines = append(htmlLines, fmt.Sprintf("<li>%s</li>", escapeHTML(trimmed[2:])))
		} else {
			htmlLines = append(htmlLines, fmt.Sprintf("<p>%s</p>", escapeHTML(trimmed)))
		}
	}

	htmlLines = append(htmlLines, "</body>", "</html>")
	return []byte(strings.Join(htmlLines, "\n")), baseName + ".html", "text/html", nil
}

func convertToTXT(data []byte, baseName, srcExt string) ([]byte, string, string, error) {
	return data, baseName + ".txt", "text/plain", nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
