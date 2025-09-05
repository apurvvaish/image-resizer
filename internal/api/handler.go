package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"image-resizer/internal/service"

	"github.com/disintegration/imaging"
)

// HealthHandler simple health endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request size
	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)

	if err := r.ParseMultipartForm(12 << 20); err != nil {
		http.Error(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, fh, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing form file 'file': "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read full file into memory
	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed reading uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(buf) == 0 {
		http.Error(w, "empty file uploaded", http.StatusBadRequest)
		return
	}

	// detect input mime/type
	inputMime := fh.Header.Get("Content-Type")
	if inputMime == "" {
		inputMime = "image/png"
	}

	// Decode original
	img, err := imaging.Decode(bytes.NewReader(buf))
	if err != nil || img == nil {
		http.Error(w, "failed decoding image: "+err.Error(), http.StatusBadRequest)
		return
	}

	// parse preset sizes
	var presets []string
	if s := r.FormValue("sizes"); s != "" {
		if err := json.Unmarshal([]byte(s), &presets); err != nil {
			presets = strings.Split(s, ",")
		}
	}

	// parse multiple custom sizes
	var customSizes []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if c := r.FormValue("customSizes"); c != "" {
		if err := json.Unmarshal([]byte(c), &customSizes); err != nil {
			http.Error(w, "failed to parse customSizes: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// parse optional single custom size (for backward compatibility)
	var custom struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if c := r.FormValue("custom"); c != "" {
		if err := json.Unmarshal([]byte(c), &custom); err == nil {
			if custom.Width > 0 && custom.Height > 0 {
				customSizes = append(customSizes, custom)
			}
		}
	}

	// parse format override
	format := strings.ToLower(r.FormValue("format"))
	var outputMime, ext string
	switch format {
	case "jpg", "jpeg", "image/jpg", "image/jpeg":
		outputMime = "image/jpeg"
		ext = ".jpg"
	case "png", "image/png":
		outputMime = "image/png"
		ext = ".png"
	default:
		// fall back
		outputMime = inputMime
		if outputMime != "image/jpeg" && outputMime != "image/png" {
			outputMime = "image/png"
			ext = ".png"
		} else {
			ext = filepath.Ext(fh.Filename)
			if ext == "" {
				ext = ".png"
			}
		}
	}

	// requested sizes
	type sizeSpec struct {
		name   string
		width  int
		height int
	}
	var requested []sizeSpec
	presetMap := map[string]int{
		"thumbnail": 100,
		"medium":    500,
		"large":     1000,
	}
	for _, p := range presets {
		if w, ok := presetMap[p]; ok {
			requested = append(requested, sizeSpec{name: p, width: w, height: 0})
		} else if n, err := strconv.Atoi(p); err == nil && n > 0 {
			requested = append(requested, sizeSpec{name: strconv.Itoa(n), width: n, height: 0})
		}
	}
	for _, cs := range customSizes {
		if cs.Width > 0 && cs.Height > 0 {
			name := fmt.Sprintf("%dx%d", cs.Width, cs.Height)
			requested = append(requested, sizeSpec{name: name, width: cs.Width, height: cs.Height})
		}
	}
	if len(requested) == 0 {
		requested = []sizeSpec{
			{name: "thumbnail", width: 100, height: 0},
			{name: "medium", width: 500, height: 0},
			{name: "large", width: 1000, height: 0},
		}
	}

	// Prepare response
	type respT struct {
		Original string            `json:"original"`
		Resized  map[string]string `json:"resized"`
		Filename string            `json:"filename"`
		Filesize int64             `json:"filesize"`
		MimeType string            `json:"mimeType"`
	}
	resp := respT{
		Resized:  map[string]string{},
		Filesize: fh.Size,
		MimeType: outputMime,
	}

	// build new filename with chosen ext
	base := strings.TrimSuffix(fh.Filename, filepath.Ext(fh.Filename))
	resp.Filename = base + ext

	// encode original
	origDataURI, err := service.EncodeImageToDataURI(img, outputMime)
	if err != nil {
		http.Error(w, "failed encoding original: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Original = origDataURI

	// resized
	for _, s := range requested {
		dstImg := service.ResizeWithPreserve(img, s.width, s.height)
		dataURI, err := service.EncodeImageToDataURI(dstImg, outputMime)
		if err != nil {
			http.Error(w, "failed encoding resized: "+err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Resized[s.name+ext] = dataURI
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
