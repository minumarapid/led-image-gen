package processor

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"math"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

type Config struct {
	Border        int
	LEDSize       int
	LEDGap        int
	LEDGamma      float64
	LEDExposure   float64
	LEDShape      bool // true: circle, false: square
	MaxWorkers    int
	EnableGlow    bool
	GlowRange     float64
	GlowStrength  float64
	GlowGamma     float64
	GlowExposure  float64
	OffLightColor color.RGBA
}

func DefaultConfig() *Config {
	return &Config{
		Border:        10,
		LEDSize:       4,
		LEDGap:        2,
		LEDGamma:      1.0,
		LEDExposure:   1.0,
		LEDShape:      false,
		MaxWorkers:    4,
		EnableGlow:    true,
		GlowRange:     1.0,
		GlowStrength:  1.75,
		GlowGamma:     1.0,
		GlowExposure:  1.0,
		OffLightColor: color.RGBA{40, 40, 40, 255}, // 透明黒
	}
}

func GenerateLEDImage(srcImage image.Image, config *Config) (image.Image, error) {
	baseLUT := createGammaLUT(config.LEDExposure, config.LEDGamma)
	glowLUT := createGammaLUT(config.GlowExposure, config.GlowGamma)
	src := ForceRGBA(srcImage)
	bounds := src.Bounds()
	srcWidth, srcHeight := bounds.Max.X, bounds.Max.Y
	finalWidth := (config.Border * 2) + (config.LEDSize+config.LEDGap)*srcWidth - config.LEDGap
	finalHeight := (config.Border * 2) + (config.LEDSize+config.LEDGap)*srcHeight - config.LEDGap
	displayWidth := (config.LEDSize + config.LEDGap) * srcWidth
	displayHeight := (config.LEDSize + config.LEDGap) * srcHeight
	baseLayer := gg.NewContext(finalWidth, finalHeight)
	var glowLayer *gg.Context
	if config.EnableGlow {
		glowLayer = gg.NewContext(finalWidth, finalHeight)
	}
	baseLayer.SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	baseLayer.Clear()
	if glowLayer != nil {
		glowLayer.SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 0})
		glowLayer.Clear()
	}
	type StripeResult struct {
		PasteX  int // 最終キャンバスのどこに貼るか（X座標）
		PasteY  int // 最終キャンバスのどこに貼るか（Y座標）
		BaseImg image.Image
		GlowImg image.Image
	}
	if config.MaxWorkers < 1 {
		config.MaxWorkers = 1
	}
	var splitX bool
	if srcWidth > srcHeight {
		splitX = true
	} else {
		splitX = false
	}
	var srcLongSide int
	if splitX {
		srcLongSide = srcWidth
	} else {
		srcLongSide = srcHeight
	}
	if config.MaxWorkers > srcLongSide {
		config.MaxWorkers = srcLongSide
	}
	results := make(chan StripeResult, config.MaxWorkers)
	var wg sync.WaitGroup
	longSideC := srcLongSide
	startX := 0
	startY := 0
	cellSize := config.LEDSize + config.LEDGap
	for i := 0; i < config.MaxWorkers; i++ {
		var baseCanvas *gg.Context
		var glowCanvas *gg.Context
		stripePixels := longSideC / (config.MaxWorkers - i)
		if stripePixels < 1 {
			stripePixels = 1
		}
		stripeSize := cellSize * stripePixels
		if splitX {
			baseCanvas = gg.NewContext(stripeSize, displayHeight)
			if config.EnableGlow {
				glowCanvas = gg.NewContext(stripeSize, displayHeight)
			}

		} else {
			baseCanvas = gg.NewContext(displayWidth, stripeSize)
			if config.EnableGlow {
				glowCanvas = gg.NewContext(displayWidth, stripeSize)
			}
		}
		wg.Add(1)
		go func(startX int, startY int, stripePixels int, splitX bool, baseCanvas *gg.Context, glowCanvas *gg.Context) {
			defer wg.Done()
			baseCanvas.SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 0})
			baseCanvas.Clear()
			if glowCanvas != nil {
				glowCanvas.SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 0})
				glowCanvas.Clear()
			}
			var w, h int
			if splitX {
				w = stripePixels
				h = srcHeight
			} else {
				w = srcWidth
				h = stripePixels
			}
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					srcX := startX + x
					srcY := startY + y
					srcColor := src.At(srcX, srcY)
					r, g, b, a := srcColor.RGBA()
					isBlack := r == 0 && g == 0 && b == 0
					baseColor := color.Color(baseLUT.Apply(srcColor))
					glowColor := color.Color(glowLUT.Apply(srcColor))
					if a == 0 {
						if config.OffLightColor.A == 0 {
							continue
						}
						baseColor = config.OffLightColor
						glowColor = config.OffLightColor
					}
					if isBlack {
						baseColor = config.OffLightColor
					}
					drawGlow := glowCanvas != nil && !isBlack
					cellX := float64(x * cellSize)
					cellY := float64(y * cellSize)
					if config.LEDShape {
						radius := float64(config.LEDSize) / 2.0
						cx := cellX + radius
						cy := cellY + radius
						baseCanvas.SetColor(baseColor)
						baseCanvas.DrawCircle(cx, cy, radius)
						baseCanvas.Fill()
						if drawGlow {
							glowCanvas.SetColor(glowColor)
							glowCanvas.DrawCircle(cx, cy, radius)
							glowCanvas.Fill()
						}
					} else {
						baseCanvas.SetColor(baseColor)
						baseCanvas.DrawRectangle(cellX, cellY, float64(config.LEDSize), float64(config.LEDSize))
						baseCanvas.Fill()
						if drawGlow {
							glowCanvas.SetColor(glowColor)
							glowCanvas.DrawRectangle(cellX, cellY, float64(config.LEDSize), float64(config.LEDSize))
							glowCanvas.Fill()
						}
					}
				}
			}
			pasteX := config.Border + (startX * cellSize)
			pasteY := config.Border + (startY * cellSize)
			results <- StripeResult{PasteX: pasteX, PasteY: pasteY, BaseImg: baseCanvas.Image(), GlowImg: func() image.Image {
				if glowCanvas == nil {
					return nil
				}
				return glowCanvas.Image()
			}()}
		}(startX, startY, stripePixels, splitX, baseCanvas, glowCanvas)
		longSideC -= stripePixels
		if splitX {
			startX += stripePixels
		} else {
			startY += stripePixels
		}
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	for res := range results {
		baseLayer.DrawImage(res.BaseImg, res.PasteX, res.PasteY)
		if config.EnableGlow && res.GlowImg != nil {
			glowLayer.DrawImage(res.GlowImg, res.PasteX, res.PasteY)
		}
	}
	baseImage := ForceRGBA(baseLayer.Image())
	if config.EnableGlow {
		blurSigma := config.GlowStrength + float64(config.LEDSize)
		if blurSigma < 0.1 {
			blurSigma = 0.1
		}
		blurredGlow := imaging.Blur(glowLayer.Image(), blurSigma)
		compositeAdditiveRGBA(baseImage, ForceRGBA(blurredGlow))
	}
	return baseImage, nil
}

func ForceRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	return dst
}

type GammaLUT struct {
	table [65536]uint16
}

func createGammaLUT(exposure float64, gamma float64) *GammaLUT {
	lut := &GammaLUT{}

	for i := 0; i <= 65535; i++ {
		val := float64(i) / 65535.0
		val = val * exposure
		if val > 1.0 {
			val = 1.0
		}
		corrected := math.Pow(val, gamma)

		lut.table[i] = uint16(corrected * 65535.0)
	}

	return lut
}

func (g *GammaLUT) Apply(c color.Color) color.RGBA64 {
	r, gVal, b, a := c.RGBA()
	return color.RGBA64{
		R: g.table[r],
		G: g.table[gVal],
		B: g.table[b],
		A: uint16(a),
	}
}

func compositeAdditiveRGBA(dst *image.RGBA, src *image.RGBA) {
	const glowAdditiveFactor = 1.5
	bounds := dst.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := dst.PixOffset(x, y)
			dst.Pix[off+0] = clampByte(int(float64(dst.Pix[off+0]) + float64(src.Pix[off+0])*glowAdditiveFactor))
			dst.Pix[off+1] = clampByte(int(float64(dst.Pix[off+1]) + float64(src.Pix[off+1])*glowAdditiveFactor))
			dst.Pix[off+2] = clampByte(int(float64(dst.Pix[off+2]) + float64(src.Pix[off+2])*glowAdditiveFactor))
		}
	}
}

func clampByte(val int) uint8 {
	if val < 0 {
		return 0
	}
	if val > 255 {
		return 255
	}
	return uint8(val)
}
