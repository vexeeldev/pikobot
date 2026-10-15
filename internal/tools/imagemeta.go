package tools

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	_ "golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

type ImageMetadata struct {
	Filename      string
	FileSize      int64
	FileSizeStr   string
	Format        string
	Width         int
	Height        int
	Megapixels    float64
	AspectRatio   string
	ColorModel    string
	HasEXIF       bool
	CameraMake    string
	CameraModel   string
	DateTime      string
	LensModel     string
	ExposureTime  string
	FNumber       string
	ISO           string
	FocalLength   string
	Flash         string
	Software      string
	HasGPS        bool
	Latitude      float64
	Longitude     float64
	GoogleMapsURL string
}

func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func calculateAspectRatio(w, h int) string {
	if w <= 0 || h <= 0 {
		return "-"
	}

	g := gcd(w, h)
	sw := w / g
	sh := h / g

	val := float64(w) / float64(h)
	orientation := "Landscape"
	if w < h {
		orientation = "Portrait"
	} else if w == h {
		orientation = "Square"
	}

	if math.Abs(val-1.0) < 0.02 {
		return fmt.Sprintf("1:1 (%s)", orientation)
	} else if math.Abs(val-1.777) < 0.05 || math.Abs(float64(h)/float64(w)-1.777) < 0.05 {
		if w > h {
			return fmt.Sprintf("16:9 (%s)", orientation)
		}
		return fmt.Sprintf("9:16 (%s)", orientation)
	} else if math.Abs(val-1.333) < 0.05 || math.Abs(float64(h)/float64(w)-1.333) < 0.05 {
		if w > h {
			return fmt.Sprintf("4:3 (%s)", orientation)
		}
		return fmt.Sprintf("3:4 (%s)", orientation)
	} else if math.Abs(val-1.5) < 0.05 || math.Abs(float64(h)/float64(w)-1.5) < 0.05 {
		if w > h {
			return fmt.Sprintf("3:2 (%s)", orientation)
		}
		return fmt.Sprintf("2:3 (%s)", orientation)
	}

	if sw <= 32 && sh <= 32 {
		return fmt.Sprintf("%d:%d (%s)", sw, sh, orientation)
	}

	return fmt.Sprintf("%.2f:1 (%s)", val, orientation)
}

func ExtractImageMetadata(data []byte, filename string) (*ImageMetadata, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data gambar kosong")
	}

	if filename == "" {
		filename = "gambar.jpg"
	}

	meta := &ImageMetadata{
		Filename:    filename,
		FileSize:    int64(len(data)),
		FileSizeStr: FormatFileSize(int64(len(data))),
		Format:      strings.ToUpper(strings.TrimPrefix(filepath.Ext(filename), ".")),
	}

	cfg, formatName, err := image.DecodeConfig(bytes.NewReader(data))
	if err == nil {
		meta.Width = cfg.Width
		meta.Height = cfg.Height
		meta.Format = strings.ToUpper(formatName)
		if cfg.ColorModel != nil {
			meta.ColorModel = fmt.Sprintf("%T", cfg.ColorModel)
			meta.ColorModel = strings.TrimPrefix(meta.ColorModel, "color.")
		}
	} else {

		if webpImg, webpErr := webp.Decode(bytes.NewReader(data)); webpErr == nil {
			bounds := webpImg.Bounds()
			meta.Width = bounds.Dx()
			meta.Height = bounds.Dy()
			meta.Format = "WEBP"
			meta.ColorModel = fmt.Sprintf("%T", webpImg.ColorModel())
			meta.ColorModel = strings.TrimPrefix(meta.ColorModel, "color.")
		}
	}

	if meta.Width > 0 && meta.Height > 0 {
		meta.Megapixels = float64(meta.Width*meta.Height) / 1000000.0
		meta.AspectRatio = calculateAspectRatio(meta.Width, meta.Height)
	}

	x, err := exif.Decode(bytes.NewReader(data))
	if err == nil && x != nil {
		meta.HasEXIF = true

		if tag, err := x.Get(exif.Make); err == nil {
			if s, err := tag.StringVal(); err == nil {
				meta.CameraMake = strings.Trim(s, "\" \x00")
			}
		}

		if tag, err := x.Get(exif.Model); err == nil {
			if s, err := tag.StringVal(); err == nil {
				meta.CameraModel = strings.Trim(s, "\" \x00")
			}
		}

		if dt, err := x.DateTime(); err == nil {
			meta.DateTime = dt.Format("02-01-2006 15:04:05")
		}

		if tag, err := x.Get(exif.LensModel); err == nil {
			if s, err := tag.StringVal(); err == nil {
				meta.LensModel = strings.Trim(s, "\" \x00")
			}
		}

		if tag, err := x.Get(exif.ExposureTime); err == nil {
			num, den, err := tag.Rat2(0)
			if err == nil {
				if num == 1 || num < den {
					meta.ExposureTime = fmt.Sprintf("%d/%d detik", num, den)
				} else {
					meta.ExposureTime = fmt.Sprintf("%.2f detik", float64(num)/float64(den))
				}
			}
		}

		if tag, err := x.Get(exif.FNumber); err == nil {
			num, den, err := tag.Rat2(0)
			if err == nil && den > 0 {
				meta.FNumber = fmt.Sprintf("f/%.1f", float64(num)/float64(den))
			}
		}

		if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
			if val, err := tag.Int(0); err == nil {
				meta.ISO = fmt.Sprintf("ISO %d", val)
			}
		}

		if tag, err := x.Get(exif.FocalLength); err == nil {
			num, den, err := tag.Rat2(0)
			if err == nil && den > 0 {
				meta.FocalLength = fmt.Sprintf("%.1f mm", float64(num)/float64(den))
			}
		}

		if tag, err := x.Get(exif.Software); err == nil {
			if s, err := tag.StringVal(); err == nil {
				meta.Software = strings.Trim(s, "\" \x00")
			}
		}

		lat, lon, err := x.LatLong()
		if err == nil && !math.IsNaN(lat) && !math.IsNaN(lon) {
			meta.HasGPS = true
			meta.Latitude = lat
			meta.Longitude = lon
			meta.GoogleMapsURL = fmt.Sprintf("https://maps.google.com/?q=%.6f,%.6f", lat, lon)
		}
	}

	return meta, nil
}
