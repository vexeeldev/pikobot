# 🤖 Piko Assistant Telegram Bot

Bot Telegram berbasis Go yang siap dideploy ke **Vercel Serverless Function** menggunakan arsitektur Webhook, terintegrasi dengan **MongoDB Atlas**, serta mendukung mode **Long Polling** untuk kemudahan testing lokal.

---

## 📁 Struktur Folder

```text
pikobot/
├── api/
│   └── index.go            # Entrypoint Vercel Serverless Function (Webhook & Healthcheck)
├── cmd/
│   ├── bot/                # Entrypoint testing lokal (menggunakan Long Polling)
│   │   └── main.go
│   └── setwebhook/         # Script CLI pembantu untuk mendaftarkan URL Vercel ke Telegram
│       └── main.go
├── internal/
│   ├── bot/                # Arsitektur modular logika bot
│   │   ├── bot.go          # Core struct, lifecycle & user state manager
│   │   ├── banner.go       # Smart banner image & header dashboard
│   │   ├── callbacks.go    # Handler klik tombol inline keyboard
│   │   ├── commands.go     # Handler perintah (/start, /menu, /saldo, dll)
│   │   ├── keyboards.go    # Layout tombol menu, saldo & tools
│   │   ├── logbook.go      # Logika logbook MIS PENS & media attachments
│   │   ├── messages.go     # Parser teks pesan & input angka
│   │   ├── saldo.go        # Logika bisnis saldo & database transaksi
│   │   └── tools.go        # Handler UI & perintah Menu Tools
│   ├── config/             # Manajemen konfigurasi (.env & Environment Variables)
│   │   └── config.go
│   ├── database/           # Koneksi MongoDB & Repository Data
│   │   ├── mongodb.go
│   │   ├── transaction.go
│   │   └── user.go
│   ├── tools/              # Komponen independen utilitas tools
│   │   ├── compress.go     # Kompresi file PDF (optimasi pdfcpu) & foto gambar
│   │   ├── converter.go    # Konversi gambar (WEBP, PNG, JPG, BMP) & dokumen (JSON, CSV, HTML)
│   │   ├── imagemeta.go    # Ekstraksi resolusi, megapixels, rasio & data EXIF kamera
│   │   ├── qrcode.go       # Generator gambar QR Code
│   │   ├── translator.go   # Terjemahan multibahasa (gtx API / fallback)
│   │   └── urltools.go     # Shortlink, cek jejak redirect & URL metadata
│   └── utils/              # Helper format & parser mata uang Rupiah
│       ├── currency.go
│       └── currency_test.go
├── .env                    # Variabel lokal (rahasia, tidak di-commit)
├── .env.example            # Template variabel
├── .gitignore
├── go.mod
├── go.sum
├── vercel.json             # Konfigurasi routing Vercel
└── README.md
```

---

## ⚙️ Konfigurasi Environment Variable

Isi kredensial di file [`.env`](file:///home/malakul-tech/project/pikobot/.env) (untuk lokal) dan di dashboard **Vercel Settings -> Environment Variables** (untuk produksi):

| Variabel | Deskripsi |
|---|---|
| `TELEGRAM_BOT_TOKEN` | Token Bot API dari BotFather |
| `TELEGRAM_BOT_USERNAME` | Username bot (misal: `PikoAssistantBot`) |
| `MONGODB_URI` | Connection URI MongoDB Atlas (ganti `<db_password>` dengan password user database) |
| `MONGODB_DATABASE` | Nama database MongoDB (default: `pikobot`) |
| `APP_ENV` | `development` atau `production` |
| `MIS_USERNAME` | Akun NetID / Email PENS untuk login otomatis ke MIS |
| `MIS_PASSWORD` | Password akun MIS PENS |
| `MIS_NRP` | NRP Mahasiswa (default: `3124521031`) |
| `MIS_KP_DAFTAR` | ID Pendaftaran KP MIS (default: `7733`) |
| `MIS_MAHASISWA` | ID Mahasiswa MIS (default: `25533`) |
| `MIS_JAM_MULAI` | Jam mulai KP default (default: `08.00`) |
| `MIS_JAM_SELESAI` | Jam selesai KP default (default: `16.00`) |

---

## 📝 Fitur Otomatisasi Logbook KP & Progres (MIS PENS)

Pikobot terintegrasi langsung dengan portal **MIS PENS** (`online.mis.pens.ac.id`) via **CAS Auto-Login**:
- **Tombol Menu**: Buka menu dashboard, klik **📝 Logbook KP**.
- **Perintah Cepat**: `/logbook mengerjakan testing alert whatsapp`
- **Jam Otomatis**: Default jam kerja otomatis diset ke `08.00 - 16.00` (dapat diubah fleksibel lewat tombol / chat).
- **Alur All-in-One (Teks + Foto + PDF Progres)**:
  - Kirim foto kegiatan (JPG) langsung ke bot (otomatis dikompres jika > 480 KB).
  - Kirim dokumen laporan mingguan (PDF) langsung ke bot (maks 1 MB).
  - Tinjau semua lampiran dalam 1 kartu konfirmasi, lalu klik **🚀 Simpan ke MIS PENS**!
- **Auto-Login CAS**: Jika sesi kedaluwarsa, bot otomatis melakukan login ulang di balik layar.


## 🚀 1. Menjalankan di Lokal (Development Mode)

Gunakan mode **polling** di lokal agar tidak perlu setup HTTPS / ngrok:

```bash
go run cmd/bot/main.go
```

Bot akan otomatis menghubungkan MongoDB, mendengarkan pesan dari Telegram, dan menyimpan info pengguna yang berinteraksi.

---

## 💰 Fitur Kelola Saldo (Wallet)

Fitur saldo kini aktif dan tersimpan permanen di **MongoDB**:
- **Buka Menu Saldo**: Ketuk tombol `👛 Saldo` di dashboard atau ketik `/saldo`.
- **Atur Saldo Langsung**:
  - `/setsaldo 500000` (atau `/setsaldo 500k`, `/setsaldo 1jt`)
  - Atau klik tombol **`[ ✏️ Atur Ulang Saldo ]`** lalu ketik angka nominalnya di chat.
- **Tambah Saldo**:
  - `/tambahsaldo 50000` (atau `/tambahsaldo 50k`)
  - Atau klik tombol **`[ 💵 Tambah Saldo ]`** lalu ketik nominalnya di chat.
- **Format Input Fleksibel**:
  Mendukung penulisan angka bebas: `50000`, `50.000`, `Rp 50.000`, `50k`, `100rb`, `1.5jt`, dll.

---

## 🛠️ Fitur Menu Tools & Utilitas Cerdas

Akses dashboard tools dengan mengetuk tombol **`🛠️ Menu Tools`** pada menu utama atau menggunakan perintah `/tools`.

### 1. 🔲 QR Code Generator
Buat gambar QR Code resolusi tinggi langsung untuk disimpan atau dibagikan:
- **Perintah:** `/qr <teks atau url>`
- **Contoh:** `/qr https://example.com`
- Atau klik tombol `🔲 QR Code Generator` pada menu tools, lalu ketik teks/link di chat.

### 2. 🔄 File Converter
Konversi format gambar, dokumen, atau file teks:
- **Format didukung:**
  - Gambar: `WEBP`, `PNG`, `JPG/JPEG`, `BMP`, `GIF`
  - Dokumen/Data: `JSON`, `CSV`, `HTML`, `TXT`
- **Cara Penggunaan:**
  - Balas (reply) foto atau dokumen dengan: `/convert png`
  - Atau kirim file baru dengan caption: `/convert png` (contoh: `/convert gambar.webp png`)

### 3. 🗜️ Compress PDF & Foto
Perkecil ukuran file PDF atau foto agar hemat memori & mudah diunggah:
- **Dokumen PDF:** Optimasi stream dan objek menggunakan engine `pdfcpu`.
- **Foto/Gambar:** Kompresi kualitas JPEG adaptif & penyesuaian resolusi.
- **Cara Penggunaan:**
  - Balas (reply) dokumen PDF atau foto dengan: `/compress`
  - Kirim dokumen/foto baru dengan caption: `/compress`
  - Atau klik tombol `🗜️ Compress PDF & Foto` pada menu tools, lalu kirim filenya.

### 4. 🔗 URL Tools
Peralatan jaringan lengkap untuk tautan web:
- **Buat Shortlink:** `/short <url>` (contoh: `/short https://example.com/long/path`)
- **Cek Jejak Redirect:** `/redirect <url>` (menelusuri status HTTP 301, 302 hingga URL akhir)
- **Cek Metadata URL:** `/urlmeta <url>` (mengambil judul, deskripsi, OpenGraph, dan info web server)
- Atau jalankan `/url` untuk membuka submenu interaktif.

### 5. 🌐 Translator
Terjemahkan pesan secara cepat dan akurat:
- **Perintah:** `/translate <bahasa> <teks>` (alias: `/tr`)
- **Contoh:** `/translate en Halo dunia` atau `/tr id Good morning`
- **Balas Pesan:** Reply pesan teks seseorang dengan `/tr en` untuk menerjemahkan langsung.
- Mendukung berbagai bahasa: English (`en`), Indonesia (`id`), Japanese (`ja`), Arabic (`ar`), Korean (`ko`), Mandarin (`zh`), dll.

### 6. 🖼️ Image Metadata Inspector
Tampilkan detail teknis foto dan metadata kamera:
- **Perintah:** `/metadata` atau `/imgmeta`
- **Informasi:** Resolusi (px), Megapixels, Aspek Rasio, Ukuran File, Format & Color Model, serta data EXIF (Model HP/Kamera, ISO, Aperture, Shutter Speed, Software, dan Titik Koordinat GPS Google Maps).
- **Cara Penggunaan:** Balas foto dengan `/metadata` atau klik tombol `🖼️ Image Metadata` lalu kirim foto.

---

## 🌐 2. Deploy ke Vercel

### Langkah A: Push ke GitHub & Import ke Vercel
1. Push repository ini ke GitHub / GitLab / Bitbucket.
2. Buka dashboard [Vercel](https://vercel.com/) -> **Add New Project** -> Import repository ini.
3. Tambahkan **Environment Variable** di pengaturan project Vercel:
   - `TELEGRAM_BOT_TOKEN`: `8915982113:AAEnJznyAnaobcBrJSJ3SXiYZ_RxpguOnZA`
   - `TELEGRAM_BOT_USERNAME`: `PikoAssistantBot`
   - `MONGODB_URI`: `mongodb+srv://malakultech_db_user:<password>@pikobot.sqixdyf.mongodb.net/?retryWrites=true&w=majority`
   - `MONGODB_DATABASE`: `pikobot`
   - `APP_ENV`: `production`
4. Klik **Deploy**.

### Langkah B: Daftarkan Webhook ke Telegram
Setelah Vercel selesai deploy (misalnya Anda mendapatkan URL `https://pikobot-xyz.vercel.app`):

Jalankan perintah ini dari komputer Anda:
```bash
go run cmd/setwebhook/main.go https://pikobot-xyz.vercel.app
```

Atau bisa juga langsung melalui browser:
```
https://api.telegram.org/bot<TOKEN_ANDA>/setWebhook?url=https://pikobot-xyz.vercel.app/
```

Kini setiap pesan yang dikirim ke bot Telegram Anda akan diproses secara serverless oleh Vercel dan data pengguna akan tersimpan di MongoDB!
