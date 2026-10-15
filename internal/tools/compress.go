package tools

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

type CompressionResult struct {
	OriginalSize   int64
	OriginalStr    string
	CompressedSize int64
	CompressedStr  string
	SavedBytes     int64
	SavedPercent   float64
	Data           []byte
	Filename       string
	MimeType       string
}

func CompressPDF(ctx context.Context, data []byte, filename string) (*CompressionResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data PDF kosong")
	}

	origSize := int64(len(data))
	rs := bytes.NewReader(data)
	var outBuf bytes.Buffer

	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	err := api.Optimize(ctx, rs, &outBuf, conf, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal mengompres PDF: %w", err)
	}

	compData := outBuf.Bytes()
	compSize := int64(len(compData))

	if compSize >= origSize {
		compData = data
		compSize = origSize
	}

	savedBytes := origSize - compSize
	savedPercent := 0.0
	if origSize > 0 {
		savedPercent = math.Max(0, (float64(savedBytes)/float64(origSize))*100.0)
	}

	outName := filename
	if !strings.HasSuffix(strings.ToLower(outName), ".pdf") {
		outName += ".pdf"
	}

	return &CompressionResult{
		OriginalSize:   origSize,
		OriginalStr:    FormatFileSize(origSize),
		CompressedSize: compSize,
		CompressedStr:  FormatFileSize(compSize),
		SavedBytes:     savedBytes,
		SavedPercent:   savedPercent,
		Data:           compData,
		Filename:       outName,
		MimeType:       "application/pdf",
	}, nil
}

type CompressImageOptions struct {
	Quality   int
	MaxDim    int
	TargetKB  int
	ForceJPEG bool
}

func CompressImage(data []byte, filename string, opts CompressImageOptions) (*CompressionResult, error) {
	origSize := int64(len(data))
	if origSize == 0 {
		return nil, fmt.Errorf("data gambar kosong")
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gagal membaca gambar: %w", err)
	}

	if opts.Quality <= 0 {
		opts.Quality = 72
	}
	if opts.MaxDim <= 0 {
		opts.MaxDim = 1920
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w > opts.MaxDim || h > opts.MaxDim {
		var newW, newH int
		if w > h {
			newW = opts.MaxDim
			newH = int(float64(h) * (float64(opts.MaxDim) / float64(w)))
		} else {
			newH = opts.MaxDim
			newW = int(float64(w) * (float64(opts.MaxDim) / float64(h)))
		}

		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
		img = dst
	}

	isPNG := strings.ToLower(format) == "png" && !opts.ForceJPEG

	var compData []byte
	var outExt string
	var mimeType string

	if isPNG && opts.TargetKB <= 0 {

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err == nil && buf.Len() < int(origSize) {
			compData = buf.Bytes()
			outExt = ".png"
			mimeType = "image/png"
		}
	}

	if len(compData) == 0 {
		outExt = ".jpg"
		mimeType = "image/jpeg"

		q := opts.Quality

		targetBytes := opts.TargetKB * 1024
		if targetBytes > 0 && targetBytes < int(origSize) {
			for tryQ := 85; tryQ >= 20; tryQ -= 15 {
				var buf bytes.Buffer
				if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: tryQ}); err == nil {
					compData = buf.Bytes()
					if buf.Len() <= targetBytes || tryQ == 20 {
						break
					}
				}
			}
		} else {
			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
				return nil, fmt.Errorf("gagal kompresi JPEG: %w", err)
			}
			compData = buf.Bytes()
		}
	}

	compSize := int64(len(compData))
	savedBytes := origSize - compSize
	savedPercent := 0.0
	if origSize > 0 {
		savedPercent = math.Max(0, (float64(savedBytes)/float64(origSize))*100.0)
	}

	baseName := strings.TrimSuffix(filename, strings.TrimPrefix(filename, "."))
	if baseName == "" {
		baseName = "compressed_image"
	}
	outName := baseName + outExt

	return &CompressionResult{
		OriginalSize:   origSize,
		OriginalStr:    FormatFileSize(origSize),
		CompressedSize: compSize,
		CompressedStr:  FormatFileSize(compSize),
		SavedBytes:     savedBytes,
		SavedPercent:   savedPercent,
		Data:           compData,
		Filename:       outName,
		MimeType:       mimeType,
	}, nil
}
