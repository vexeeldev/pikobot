package mis

import (
	"time"
)

type LogbookEntry struct {
	Tanggal      string
	JamMulai     string
	JamSelesai   string
	Kegiatan     string
	SesuaiKuliah int
	MataKuliah   string
	Minggu       int
	FotoBytes    []byte
	FotoFilename string
	PDFBytes     []byte
	PDFFilename  string
}

type Config struct {
	Username   string
	Password   string
	NRP        string
	KPDaftar   string
	Mahasiswa  string
	Tahun      string
	Semester   string
	JamMulai   string
	JamSelesai string
}

func (c Config) DefaultLogbookEntry(kegiatan string) *LogbookEntry {
	jamMulai := c.JamMulai
	if jamMulai == "" {
		jamMulai = "08.00"
	}
	jamSelesai := c.JamSelesai
	if jamSelesai == "" {
		jamSelesai = "16.00"
	}

	return &LogbookEntry{
		Tanggal:      time.Now().Format("2006-01-02"),
		JamMulai:     jamMulai,
		JamSelesai:   jamSelesai,
		Kegiatan:     kegiatan,
		SesuaiKuliah: 2,
		Minggu:       16,
	}
}
