package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TranslationResult struct {
	OriginalText   string
	TranslatedText string
	SourceLang     string
	SourceLangName string
	TargetLang     string
	TargetLangName string
}

var langMap = map[string]string{
	"en":         "en",
	"eng":        "en",
	"english":    "en",
	"inggris":    "en",
	"id":         "id",
	"ind":        "id",
	"indo":       "id",
	"indonesia":  "id",
	"ja":         "ja",
	"jp":         "ja",
	"jpn":        "ja",
	"jepang":     "ja",
	"japanese":   "ja",
	"ar":         "ar",
	"ara":        "ar",
	"arab":       "ar",
	"arabic":     "ar",
	"ko":         "ko",
	"kor":        "ko",
	"korea":      "ko",
	"korean":     "ko",
	"zh":         "zh-CN",
	"cn":         "zh-CN",
	"chinese":    "zh-CN",
	"mandarin":   "zh-CN",
	"cina":       "zh-CN",
	"es":         "es",
	"spanish":    "es",
	"spanyol":    "es",
	"fr":         "fr",
	"french":     "fr",
	"prancis":    "fr",
	"de":         "de",
	"german":     "de",
	"jerman":     "de",
	"ru":         "ru",
	"russian":    "ru",
	"rusia":      "ru",
	"it":         "it",
	"italian":    "it",
	"italia":     "it",
	"pt":         "pt",
	"portuguese": "pt",
	"portugis":   "pt",
	"nl":         "nl",
	"dutch":      "nl",
	"belanda":    "nl",
	"th":         "th",
	"thai":       "th",
	"thailand":   "th",
	"vi":         "vi",
	"vietnam":    "vi",
	"vietnamese": "vi",
	"ms":         "ms",
	"melayu":     "ms",
	"malay":      "ms",
	"jv":         "jv",
	"jw":         "jv",
	"jawa":       "jv",
	"su":         "su",
	"sunda":      "su",
	"tr":         "tr",
	"turki":      "tr",
	"turkish":    "tr",
}

var langNames = map[string]string{
	"en":    "🇬🇧 English (Inggris)",
	"id":    "🇮🇩 Bahasa Indonesia",
	"ja":    "🇯🇵 日本語 (Jepang)",
	"ar":    "🇸🇦 العربية (Arab)",
	"ko":    "🇰🇷 한국어 (Korea)",
	"zh-CN": "🇨🇳 中文 (Mandarin)",
	"zh":    "🇨🇳 中文 (Mandarin)",
	"es":    "🇪🇸 Español (Spanyol)",
	"fr":    "🇫🇷 Français (Prancis)",
	"de":    "🇩🇪 Deutsch (Jerman)",
	"ru":    "🇷🇺 Русский (Rusia)",
	"it":    "🇮🇹 Italiano (Italia)",
	"pt":    "🇵🇹 Português (Portugis)",
	"nl":    "🇳🇱 Nederlands (Belanda)",
	"th":    "🇹🇭 ภาษาไทย (Thailand)",
	"vi":    "🇻🇳 Tiếng Việt (Vietnam)",
	"ms":    "🇲🇾 Bahasa Melayu",
	"jv":    "🇮🇩 Basa Jawa",
	"su":    "🇮🇩 Basa Sunda",
	"tr":    "🇹🇷 Türkçe (Turki)",
}

func NormalizeLanguageCode(input string) string {
	clean := strings.ToLower(strings.TrimSpace(input))
	if code, ok := langMap[clean]; ok {
		return code
	}
	if len(clean) == 2 {
		return clean
	}
	return "en"
}

func GetLanguageName(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if name, ok := langNames[code]; ok {
		return name
	}

	for k, v := range langNames {
		if strings.HasPrefix(code, k) {
			return v
		}
	}
	return strings.ToUpper(code)
}

func TranslateText(ctx context.Context, text, targetLang, sourceLang string) (*TranslationResult, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, fmt.Errorf("teks yang akan diterjemahkan tidak boleh kosong")
	}

	tl := NormalizeLanguageCode(targetLang)
	sl := "auto"
	if sourceLang != "" && sourceLang != "auto" {
		sl = NormalizeLanguageCode(sourceLang)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	apiURL := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=%s&dt=t&q=%s",
		sl, tl, url.QueryEscape(trimmed),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request translasi: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				res, parseErr := parseGoogleTranslateResponse(body, trimmed, tl)
				if parseErr == nil {
					return res, nil
				}
			}
		}
	}

	srcPair := sl
	if srcPair == "auto" {
		srcPair = "autodetect"
	}
	mmURL := fmt.Sprintf(
		"https://api.mymemory.translated.net/get?q=%s&langpair=%s|%s",
		url.QueryEscape(trimmed), srcPair, tl,
	)

	reqMM, err := http.NewRequestWithContext(ctx, http.MethodGet, mmURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungkan ke layanan translasi cadangan: %w", err)
	}
	reqMM.Header.Set("User-Agent", "PikoBot/1.0")

	respMM, err := client.Do(reqMM)
	if err != nil {
		return nil, fmt.Errorf("gagal menerjemahkan teks: %w", err)
	}
	defer respMM.Body.Close()

	bodyMM, err := io.ReadAll(respMM.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons translasi: %w", err)
	}

	var mmResp struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
		ResponseStatus int `json:"responseStatus"`
	}

	if err := json.Unmarshal(bodyMM, &mmResp); err != nil || mmResp.ResponseData.TranslatedText == "" {
		return nil, fmt.Errorf("layanan translasi sedang sibuk, silakan coba lagi beberapa saat")
	}

	return &TranslationResult{
		OriginalText:   trimmed,
		TranslatedText: mmResp.ResponseData.TranslatedText,
		SourceLang:     sl,
		SourceLangName: GetLanguageName(sl),
		TargetLang:     tl,
		TargetLangName: GetLanguageName(tl),
	}, nil
}

func parseGoogleTranslateResponse(body []byte, originalText, targetLang string) (*TranslationResult, error) {
	var raw []interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return nil, fmt.Errorf("respon translasi kosong")
	}

	sentences, ok := raw[0].([]interface{})
	if !ok || len(sentences) == 0 {
		return nil, fmt.Errorf("struktur kalimat translasi tidak sesuai")
	}

	var sb strings.Builder
	for _, item := range sentences {
		if piece, ok := item.([]interface{}); ok && len(piece) > 0 {
			if str, ok := piece[0].(string); ok {
				sb.WriteString(str)
			}
		}
	}

	translated := strings.TrimSpace(sb.String())
	if translated == "" {
		return nil, fmt.Errorf("hasil translasi kosong")
	}

	detectedLang := "auto"
	if len(raw) > 2 {
		if det, ok := raw[2].(string); ok && det != "" {
			detectedLang = det
		}
	}

	return &TranslationResult{
		OriginalText:   originalText,
		TranslatedText: translated,
		SourceLang:     detectedLang,
		SourceLangName: GetLanguageName(detectedLang),
		TargetLang:     targetLang,
		TargetLangName: GetLanguageName(targetLang),
	}, nil
}
