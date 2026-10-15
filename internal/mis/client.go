package mis

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

const (
	misBaseURL       = "https://online.mis.pens.ac.id"
	misLoginStartURL = "https://online.mis.pens.ac.id/index.php?Login=1&halAwal=1"
	misLogbookForm   = "https://online.mis.pens.ac.id/mEntry_Logbook_KP1.php"
	misLogbookSubmit = "https://online.mis.pens.ac.id/entry_logbook_kp1.php"
	casHost          = "https://login.pens.ac.id"
	userAgent        = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

var (
	reLT         = regexp.MustCompile(`name=["']lt["']\s+value=["']([^"']+)["']`)
	reFormAction = regexp.MustCompile(`<form[^>]*action=["']([^"']+)["']`)
	reValMinggu  = regexp.MustCompile(`name=["']valMinggu["'][^>]*value=["'](\d+)["']`)
)

type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	isLoggedIn bool
	lastLogin  time.Time
}

func NewClient(cfg Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 20 * time.Second,
	}

	return &Client{
		cfg:        cfg,
		httpClient: client,
	}, nil
}

func (c *Client) SetConfig(cfg Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg = cfg
	c.isLoggedIn = false
}

func (c *Client) GetConfig() Config {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cfg
}

func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) Login(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cfg.Username == "" || c.cfg.Password == "" {
		_ = godotenv.Load()
		if u := os.Getenv("MIS_USERNAME"); u != "" {
			c.cfg.Username = u
		}
		if p := os.Getenv("MIS_PASSWORD"); p != "" {
			c.cfg.Password = p
		}
	}

	if c.cfg.Username == "" || c.cfg.Password == "" {
		return errors.New("MIS_USERNAME atau MIS_PASSWORD belum dikonfigurasi di .env")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, misLoginStartURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gagal menghubungi portal MIS PENS: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gagal membaca halaman login: %w", err)
	}
	bodyStr := string(bodyBytes)

	ltMatches := reLT.FindStringSubmatch(bodyStr)
	if len(ltMatches) < 2 {
		return errors.New("tidak dapat menemukan token login 'lt' dari CAS PENS")
	}
	ltToken := ltMatches[1]

	var postURL string
	actionMatches := reFormAction.FindStringSubmatch(bodyStr)
	if len(actionMatches) >= 2 {
		action := actionMatches[1]
		if strings.HasPrefix(action, "http") {
			postURL = action
		} else {
			postURL = casHost + action
		}
	} else {
		postURL = resp.Request.URL.String()
	}

	formData := url.Values{
		"username": {c.cfg.Username},
		"password": {c.cfg.Password},
		"lt":       {ltToken},
		"_eventId": {"submit"},
		"submit":   {"LOGIN"},
	}

	loginReq, err := http.NewRequestWithContext(ctx, http.MethodPost, postURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("User-Agent", userAgent)
	loginReq.Header.Set("Referer", resp.Request.URL.String())

	loginResp, err := c.httpClient.Do(loginReq)
	if err != nil {
		return fmt.Errorf("gagal submit login ke CAS: %w", err)
	}
	defer loginResp.Body.Close()

	verifyReq, err := http.NewRequestWithContext(ctx, http.MethodGet, misLogbookForm, nil)
	if err != nil {
		return err
	}
	verifyReq.Header.Set("User-Agent", userAgent)

	verifyResp, err := c.httpClient.Do(verifyReq)
	if err != nil {
		return fmt.Errorf("gagal verifikasi sesi login: %w", err)
	}
	defer verifyResp.Body.Close()

	verifyBody, _ := io.ReadAll(verifyResp.Body)
	verifyStr := string(verifyBody)

	if strings.Contains(verifyStr, "Anda harus Login terlebih") {
		c.isLoggedIn = false
		return errors.New("autentikasi gagal: periksa kembali MIS_USERNAME dan MIS_PASSWORD kamu")
	}

	c.isLoggedIn = true
	c.lastLogin = time.Now()
	return nil
}

func (c *Client) FetchLogbookPage(ctx context.Context) (string, error) {
	if !c.isLoggedIn || time.Since(c.lastLogin) > 30*time.Minute {
		if err := c.Login(ctx); err != nil {
			return "", err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, misLogbookForm, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func (c *Client) FetchActiveWeek(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, misLogbookForm, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	if strings.Contains(bodyStr, "Anda harus Login terlebih") {
		return 0, errors.New("sesi kedaluwarsa")
	}

	matches := reValMinggu.FindStringSubmatch(bodyStr)
	if len(matches) >= 2 {
		if week, err := strconv.Atoi(matches[1]); err == nil && week > 0 {
			return week, nil
		}
	}

	return 0, errors.New("minggu tidak ditemukan di halaman")
}

func (c *Client) SubmitLogbook(ctx context.Context, entry *LogbookEntry) error {

	if !c.isLoggedIn || time.Since(c.lastLogin) > 30*time.Minute {
		if err := c.Login(ctx); err != nil {
			return err
		}
	}

	if entry.Minggu <= 0 {
		if week, err := c.FetchActiveWeek(ctx); err == nil {
			entry.Minggu = week
		} else {
			entry.Minggu = 16
		}
	}

	doSubmit := func() (string, error) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		sid := fmt.Sprintf("%f", r.Float64())

		sesuai := entry.SesuaiKuliah
		if sesuai == 0 {
			sesuai = 2
		}

		jamMulai := entry.JamMulai
		if jamMulai == "" {
			jamMulai = c.cfg.JamMulai
			if jamMulai == "" {
				jamMulai = "08.00"
			}
		}

		jamSelesai := entry.JamSelesai
		if jamSelesai == "" {
			jamSelesai = c.cfg.JamSelesai
			if jamSelesai == "" {
				jamSelesai = "16.00"
			}
		}

		tanggal := entry.Tanggal
		if tanggal == "" {
			tanggal = time.Now().Format("2006-01-02")
		}

		formData := url.Values{
			"valnrpMahasiswa": {c.cfg.NRP},
			"valTahun":        {c.cfg.Tahun},
			"valSemester":     {c.cfg.Semester},
			"Simpan":          {"1"},
			"valMinggu":       {fmt.Sprintf("%d", entry.Minggu)},
			"tanggal":         {tanggal},
			"jam_mulai":       {jamMulai},
			"jam_selesai":     {jamSelesai},
			"kegiatan":        {entry.Kegiatan},
			"sesuai_kuliah":   {fmt.Sprintf("%d", sesuai)},
			"matakuliah":      {entry.MataKuliah},
			"kp_daftar":       {c.cfg.KPDaftar},
			"mahasiswa":       {c.cfg.Mahasiswa},
			"Setuju":          {"1"},
			"sid":             {sid},
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, misLogbookSubmit, strings.NewReader(formData.Encode()))
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Referer", misLogbookForm)
		req.Header.Set("Origin", misBaseURL)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("gagal mengirim request logbook: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		return string(respBody), nil
	}

	respStr, err := doSubmit()
	if err != nil {
		return err
	}

	if strings.Contains(respStr, "Anda harus Login terlebih") {
		c.isLoggedIn = false
		if err := c.Login(ctx); err != nil {
			return fmt.Errorf("sesi kedaluwarsa dan gagal login ulang: %w", err)
		}
		_, err = doSubmit()
		if err != nil {
			return fmt.Errorf("gagal submit ulang setelah login: %w", err)
		}
	}

	return nil
}

func (c *Client) UploadFotoKegiatan(ctx context.Context, entry *LogbookEntry) error {
	if len(entry.FotoBytes) == 0 {
		return nil
	}

	if !c.isLoggedIn || time.Since(c.lastLogin) > 30*time.Minute {
		if err := c.Login(ctx); err != nil {
			return err
		}
	}

	doUpload := func() (string, error) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("tahun", c.cfg.Tahun)
		_ = writer.WriteField("cbSemester", c.cfg.Semester)
		_ = writer.WriteField("tanggal", entry.Tanggal)
		_ = writer.WriteField("tanggal_dp", "1")
		_ = writer.WriteField("tanggal_year_start", "2020")
		_ = writer.WriteField("tanggal_year_end", "2026")
		_ = writer.WriteField("tanggal_mon", "")
		_ = writer.WriteField("tanggal_da1", "1603645200")
		_ = writer.WriteField("tanggal_da2", "1791997200")
		_ = writer.WriteField("tanggal_sna", "1")
		_ = writer.WriteField("tanggal_aut", "")
		_ = writer.WriteField("tanggal_frm", "")
		_ = writer.WriteField("tanggal_tar", "")
		_ = writer.WriteField("tanggal_inp", "")
		_ = writer.WriteField("tanggal_fmt", "d-m-Y")
		_ = writer.WriteField("tanggal_dis", "")
		_ = writer.WriteField("tanggal_pr1", "")
		_ = writer.WriteField("tanggal_pr2", "")
		_ = writer.WriteField("tanggal_prv", "")
		_ = writer.WriteField("tanggal_pth", "calendar/")
		_ = writer.WriteField("tanggal_spd", "[[],[],[]]")
		_ = writer.WriteField("tanggal_spt", "0")
		_ = writer.WriteField("tanggal_och", "")
		_ = writer.WriteField("jam_mulai", entry.JamMulai)
		_ = writer.WriteField("jam_selesai", entry.JamSelesai)
		_ = writer.WriteField("kegiatan", entry.Kegiatan)
		_ = writer.WriteField("sesuai_kuliah", fmt.Sprintf("%d", entry.SesuaiKuliah))
		_ = writer.WriteField("matakuliah", entry.MataKuliah)
		_ = writer.WriteField("Setuju", "1")
		_ = writer.WriteField("kp_daftar", c.cfg.KPDaftar)
		_ = writer.WriteField("mahasiswa", c.cfg.Mahasiswa)

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="fileupload"; filename=""`)
		h.Set("Content-Type", "application/octet-stream")
		_, _ = writer.CreatePart(h)

		_ = writer.WriteField("kirimfile1", "")

		filename := entry.FotoFilename
		if filename == "" {
			filename = "kegiatan.jpg"
		}
		h2 := make(textproto.MIMEHeader)
		h2.Set("Content-Disposition", fmt.Sprintf(`form-data; name="fileupload2"; filename="%s"`, filename))
		h2.Set("Content-Type", "image/jpeg")
		part2, err := writer.CreatePart(h2)
		if err != nil {
			return "", fmt.Errorf("gagal membuat form file foto: %w", err)
		}
		if _, err := part2.Write(entry.FotoBytes); err != nil {
			return "", fmt.Errorf("gagal menulis data foto: %w", err)
		}

		_ = writer.WriteField("kirimfile2", "1")
		_ = writer.WriteField("minggu", fmt.Sprintf("%d", entry.Minggu))

		if err := writer.Close(); err != nil {
			return "", fmt.Errorf("gagal menutup multipart writer: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, misLogbookForm, body)
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Referer", misLogbookForm)
		req.Header.Set("Origin", misBaseURL)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("gagal mengirim request upload foto: %w", err)
		}
		defer resp.Body.Close()

		respBytes, _ := io.ReadAll(resp.Body)
		return string(respBytes), nil
	}

	respStr, err := doUpload()
	if err != nil {
		return err
	}

	if strings.Contains(respStr, "Anda harus Login terlebih") {
		c.isLoggedIn = false
		if err := c.Login(ctx); err != nil {
			return fmt.Errorf("sesi kedaluwarsa saat upload foto: %w", err)
		}
		respStr, err = doUpload()
		if err != nil {
			return fmt.Errorf("gagal upload foto setelah login ulang: %w", err)
		}
	}

	if !strings.Contains(respStr, "File berhasil diunggah") {
		reAlert := regexp.MustCompile(`alert\(['"]([^'"]+)['"]\)`)
		if m := reAlert.FindStringSubmatch(respStr); len(m) > 1 {
			return fmt.Errorf("server MIS menolak: %s", m[1])
		}
		return errors.New("server MIS tidak memberikan konfirmasi sukses upload foto")
	}

	return nil
}

func (c *Client) UploadLaporanProgres(ctx context.Context, entry *LogbookEntry) error {
	if len(entry.PDFBytes) == 0 {
		return errors.New("file PDF kosong")
	}

	if !c.isLoggedIn || time.Since(c.lastLogin) > 30*time.Minute {
		if err := c.Login(ctx); err != nil {
			return err
		}
	}

	minggu := entry.Minggu
	if minggu <= 0 {
		if week, err := c.FetchActiveWeek(ctx); err == nil {
			minggu = week
		} else {
			minggu = 16
		}
	}

	doUpload := func() (string, error) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		tanggal := entry.Tanggal
		if tanggal == "" {
			tanggal = time.Now().Format("2006-01-02")
		}
		jamMulai := entry.JamMulai
		if jamMulai == "" {
			jamMulai = c.cfg.JamMulai
			if jamMulai == "" {
				jamMulai = "08.00"
			}
		}
		jamSelesai := entry.JamSelesai
		if jamSelesai == "" {
			jamSelesai = c.cfg.JamSelesai
			if jamSelesai == "" {
				jamSelesai = "16.00"
			}
		}
		kegiatan := entry.Kegiatan
		if kegiatan == "" {
			kegiatan = fmt.Sprintf("Progres Laporan KP Minggu ke-%d", minggu)
		}
		sesuai := entry.SesuaiKuliah
		if sesuai == 0 {
			sesuai = 2
		}

		_ = writer.WriteField("tahun", c.cfg.Tahun)
		_ = writer.WriteField("cbSemester", c.cfg.Semester)
		_ = writer.WriteField("tanggal", tanggal)
		_ = writer.WriteField("tanggal_dp", "1")
		_ = writer.WriteField("tanggal_year_start", "2020")
		_ = writer.WriteField("tanggal_year_end", "2026")
		_ = writer.WriteField("tanggal_mon", "")
		_ = writer.WriteField("tanggal_da1", "1603645200")
		_ = writer.WriteField("tanggal_da2", "1791997200")
		_ = writer.WriteField("tanggal_sna", "1")
		_ = writer.WriteField("tanggal_aut", "")
		_ = writer.WriteField("tanggal_frm", "")
		_ = writer.WriteField("tanggal_tar", "")
		_ = writer.WriteField("tanggal_inp", "")
		_ = writer.WriteField("tanggal_fmt", "d-m-Y")
		_ = writer.WriteField("tanggal_dis", "")
		_ = writer.WriteField("tanggal_pr1", "")
		_ = writer.WriteField("tanggal_pr2", "")
		_ = writer.WriteField("tanggal_prv", "")
		_ = writer.WriteField("tanggal_pth", "calendar/")
		_ = writer.WriteField("tanggal_spd", "[[],[],[]]")
		_ = writer.WriteField("tanggal_spt", "0")
		_ = writer.WriteField("tanggal_och", "")
		_ = writer.WriteField("jam_mulai", jamMulai)
		_ = writer.WriteField("jam_selesai", jamSelesai)
		_ = writer.WriteField("kegiatan", kegiatan)
		_ = writer.WriteField("sesuai_kuliah", fmt.Sprintf("%d", sesuai))
		_ = writer.WriteField("matakuliah", entry.MataKuliah)
		_ = writer.WriteField("Setuju", "1")
		_ = writer.WriteField("kp_daftar", c.cfg.KPDaftar)
		_ = writer.WriteField("mahasiswa", c.cfg.Mahasiswa)

		filename := entry.PDFFilename
		if filename == "" {
			filename = fmt.Sprintf("Laporan_KP_Minggu_%d.pdf", minggu)
		}
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="fileupload"; filename="%s"`, filename))
		h.Set("Content-Type", "application/pdf")
		part1, err := writer.CreatePart(h)
		if err != nil {
			return "", fmt.Errorf("gagal membuat form file PDF: %w", err)
		}
		if _, err := part1.Write(entry.PDFBytes); err != nil {
			return "", fmt.Errorf("gagal menulis data PDF: %w", err)
		}

		_ = writer.WriteField("kirimfile1", "1")

		h2 := make(textproto.MIMEHeader)
		h2.Set("Content-Disposition", `form-data; name="fileupload2"; filename=""`)
		h2.Set("Content-Type", "application/octet-stream")
		_, _ = writer.CreatePart(h2)

		_ = writer.WriteField("kirimfile2", "")
		_ = writer.WriteField("minggu", fmt.Sprintf("%d", minggu))

		if err := writer.Close(); err != nil {
			return "", fmt.Errorf("gagal menutup multipart writer: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, misLogbookForm, body)
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Referer", misLogbookForm)
		req.Header.Set("Origin", misBaseURL)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("gagal mengirim request upload PDF: %w", err)
		}
		defer resp.Body.Close()

		respBytes, _ := io.ReadAll(resp.Body)
		return string(respBytes), nil
	}

	respStr, err := doUpload()
	if err != nil {
		return err
	}

	if strings.Contains(respStr, "Anda harus Login terlebih") {
		c.isLoggedIn = false
		if err := c.Login(ctx); err != nil {
			return fmt.Errorf("sesi kedaluwarsa saat upload PDF: %w", err)
		}
		respStr, err = doUpload()
		if err != nil {
			return fmt.Errorf("gagal upload PDF setelah login ulang: %w", err)
		}
	}

	if !strings.Contains(respStr, "File berhasil diunggah") {
		reAlert := regexp.MustCompile(`alert\(['"]([^'"]+)['"]\)`)
		if m := reAlert.FindStringSubmatch(respStr); len(m) > 1 {
			return fmt.Errorf("server MIS menolak: %s", m[1])
		}
		return errors.New("server MIS tidak memberikan konfirmasi sukses upload PDF")
	}

	return nil
}

func (c *Client) UploadLaporanProgresByData(ctx context.Context, minggu int, filename string, pdfData []byte) error {
	return c.UploadLaporanProgres(ctx, &LogbookEntry{
		Minggu:      minggu,
		PDFFilename: filename,
		PDFBytes:    pdfData,
	})
}
