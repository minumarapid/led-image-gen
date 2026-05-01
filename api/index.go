package api

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"strconv"
	"strings"

	"led-image-gen/processor"
)

const maxFormMemory = 32 << 20

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(maxFormMemory); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "invalid image", http.StatusBadRequest)
		return
	}

	config := defaultConfig()
	applyConfigFromForm(&config, r)

	result, err := processor.GenerateLEDImage(src, &config)
	if err != nil {
		http.Error(w, "failed to process image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, result); err != nil {
		http.Error(w, "failed to encode image", http.StatusInternalServerError)
		return
	}
}

func defaultConfig() processor.Config {
	return processor.Config{
		Border:        10,
		LEDSize:       4,
		LEDGap:        2,
		LEDGamma:      1.0,
		LEDExposure:   1.0,
		LEDShape:      true,
		MaxWorkers:    1,
		EnableGlow:    true,
		GlowStrength:  1.0,
		GlowGamma:     1.0,
		GlowExposure:  1.0,
		OffLightColor: color.RGBA{40, 40, 40, 255}, // 透明黒
	}
}

func applyConfigFromForm(cfg *processor.Config, r *http.Request) {
	if v, ok := getFormValue(r, "Border"); ok {
		cfg.Border = parseInt(v, cfg.Border)
	}
	if v, ok := getFormValue(r, "LEDSize"); ok {
		cfg.LEDSize = parseInt(v, cfg.LEDSize)
	}
	if v, ok := getFormValue(r, "LEDGap"); ok {
		cfg.LEDGap = parseInt(v, cfg.LEDGap)
	}
	if v, ok := getFormValue(r, "LEDGamma"); ok {
		cfg.LEDGamma = parseFloat(v, cfg.LEDGamma)
	}
	if v, ok := getFormValue(r, "LEDExposure"); ok {
		cfg.LEDExposure = parseFloat(v, cfg.LEDExposure)
	}
	if v, ok := getFormValue(r, "LEDShape"); ok {
		cfg.LEDShape = parseBool(v, cfg.LEDShape)
	}
	if v, ok := getFormValue(r, "EnableGlow"); ok {
		cfg.EnableGlow = parseBool(v, cfg.EnableGlow)
	}
	if v, ok := getFormValue(r, "GlowStrength"); ok {
		cfg.GlowStrength = parseFloat(v, cfg.GlowStrength)
	}
	if v, ok := getFormValue(r, "GlowGamma"); ok {
		cfg.GlowGamma = parseFloat(v, cfg.GlowGamma)
	}
	if v, ok := getFormValue(r, "GlowExposure"); ok {
		cfg.GlowExposure = parseFloat(v, cfg.GlowExposure)
	}
	if v, ok := getFormValue(r, "OffLightColor"); ok {
		cfg.OffLightColor = parseRGBA(v, cfg.OffLightColor)
	}
}

func getFormValue(r *http.Request, key string) (string, bool) {
	vals := r.MultipartForm.Value[key]
	if len(vals) == 0 {
		return "", false
	}
	return vals[0], true
}

func parseInt(value string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return v
}

func parseFloat(value string, fallback float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return fallback
	}
	return v
}

func parseBool(value string, fallback bool) bool {
	v, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return v
}

func parseRGBA(value string, fallback color.RGBA) color.RGBA {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if strings.HasPrefix(value, "#") {
		value = strings.TrimPrefix(value, "#")
		if len(value) == 6 {
			value += "FF"
		}
		if len(value) == 8 {
			if hex, err := strconv.ParseUint(value, 16, 32); err == nil {
				return color.RGBA{
					R: uint8(hex >> 24),
					G: uint8(hex >> 16),
					B: uint8(hex >> 8),
					A: uint8(hex),
				}
			}
		}
	}
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return fallback
	}
	return color.RGBA{
		R: uint8(parseInt(parts[0], int(fallback.R))),
		G: uint8(parseInt(parts[1], int(fallback.G))),
		B: uint8(parseInt(parts[2], int(fallback.B))),
		A: uint8(parseInt(parts[3], int(fallback.A))),
	}
}
