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
		LEDShape:      true,
		MaxWorkers:    4,
		EnableGlow:    true,
		GlowRange:     1.0,
		GlowStrength:  1.75,
		GlowGamma:     1.0,
		GlowExposure:  1.0,
		OffLightColor: color.RGBA{R: 50, G: 50, B: 50, A: 255},
	}
}

// LEDShapeCache: LED形状をプリレンダリングしてキャッシュ
type LEDShapeCache struct {
	mu     sync.RWMutex
	shapes map[ledShapeKey]*image.RGBA
}

type ledShapeKey struct {
	size   int
	circle bool
}

func NewLEDShapeCache() *LEDShapeCache {
	return &LEDShapeCache{
		shapes: make(map[ledShapeKey]*image.RGBA),
	}
}

func (lsc *LEDShapeCache) GetShape(ledSize int, isCircle bool) *image.RGBA {
	cacheKey := ledShapeKey{size: ledSize, circle: isCircle}

	lsc.mu.RLock()
	if shape, ok := lsc.shapes[cacheKey]; ok {
		lsc.mu.RUnlock()
		return shape
	}
	lsc.mu.RUnlock()

	shape := createLEDTemplate(ledSize, isCircle)
	lsc.mu.Lock()
	lsc.shapes[cacheKey] = shape
	lsc.mu.Unlock()

	return shape
}

func createLEDTemplate(ledSize int, isCircle bool) *image.RGBA {
	canvas := gg.NewContext(ledSize, ledSize)
	canvas.SetRGBA(0, 0, 0, 0)
	canvas.Clear()

	if isCircle {
		radius := float64(ledSize) / 2.0
		canvas.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
		canvas.DrawCircle(radius, radius, radius)
		canvas.Fill()
	} else {
		canvas.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
		canvas.DrawRectangle(0, 0, float64(ledSize), float64(ledSize))
		canvas.Fill()
	}

	return ForceRGBA(canvas.Image())
}

func GenerateLEDImage(srcImage image.Image, config *Config) (image.Image, error) {
	baseLUT := createGammaLUT(config.LEDExposure, config.LEDGamma)
	glowLUT := createGammaLUT(config.GlowExposure, config.GlowGamma)
	ledShapeCache := NewLEDShapeCache()

	src := ForceRGBA(srcImage)
	bounds := src.Bounds()
	srcWidth, srcHeight := bounds.Max.X, bounds.Max.Y
	finalWidth := (config.Border * 2) + (config.LEDSize+config.LEDGap)*srcWidth - config.LEDGap
	finalHeight := (config.Border * 2) + (config.LEDSize+config.LEDGap)*srcHeight - config.LEDGap
	displayWidth := (config.LEDSize + config.LEDGap) * srcWidth
	displayHeight := (config.LEDSize + config.LEDGap) * srcHeight
	baseLayer := image.NewRGBA(image.Rect(0, 0, finalWidth, finalHeight))
	fillRGBA(baseLayer, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	var glowLayer *image.RGBA
	if config.EnableGlow {
		glowLayer = image.NewRGBA(image.Rect(0, 0, finalWidth, finalHeight))
		fillRGBA(glowLayer, color.RGBA{R: 0, G: 0, B: 0, A: 0})
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
		var baseCanvas *image.RGBA
		var glowCanvas *image.RGBA
		stripePixels := longSideC / (config.MaxWorkers - i)
		if stripePixels < 1 {
			stripePixels = 1
		}
		stripeSize := cellSize * stripePixels
		if splitX {
			baseCanvas = image.NewRGBA(image.Rect(0, 0, stripeSize, displayHeight))
			if config.EnableGlow {
				glowCanvas = image.NewRGBA(image.Rect(0, 0, stripeSize, displayHeight))
			}

		} else {
			baseCanvas = image.NewRGBA(image.Rect(0, 0, displayWidth, stripeSize))
			if config.EnableGlow {
				glowCanvas = image.NewRGBA(image.Rect(0, 0, displayWidth, stripeSize))
			}
		}
		wg.Add(1)
		go func(startX int, startY int, stripePixels int, splitX bool, baseCanvas *image.RGBA, glowCanvas *image.RGBA) {
			defer wg.Done()
			var w, h int
			if splitX {
				w = stripePixels
				h = srcHeight
			} else {
				w = srcWidth
				h = stripePixels
			}

			// ローカルキャッシュでアクセスを削減
			cellSize := config.LEDSize + config.LEDGap
			offLightColor := config.OffLightColor
			isGlowEnabled := glowCanvas != nil

			// プリレンダリング形状を取得
			ledTemplate := ledShapeCache.GetShape(config.LEDSize, config.LEDShape)
			baseImage := baseCanvas
			var glowImage *image.RGBA
			if isGlowEnabled {
				glowImage = glowCanvas
			}

			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					srcX := startX + x
					srcY := startY + y
					srcColor := src.At(srcX, srcY)
					r, g, b, a := srcColor.RGBA()
					isBlack := r == 0 && g == 0 && b == 0

					baseColor := applyLUTToRGBA(srcColor, baseLUT)
					var glowColor color.RGBA
					if isGlowEnabled {
						glowColor = applyLUTToRGBA(srcColor, glowLUT)
					}

					if a == 0 {
						if offLightColor.A == 0 {
							continue
						}
						baseColor = offLightColor
						if isGlowEnabled {
							glowColor = offLightColor
						}
					}
					if isBlack {
						baseColor = offLightColor
					}

					drawGlow := isGlowEnabled && !isBlack
					cellX := x * cellSize
					cellY := y * cellSize

					// プリレンダリング形状を色付けして配置
					drawColoredLED(baseImage, ledTemplate, cellX, cellY, baseColor)
					if drawGlow && glowImage != nil {
						drawColoredLED(glowImage, ledTemplate, cellX, cellY, glowColor)
					}
				}
			}

			pasteX := config.Border + (startX * cellSize)
			pasteY := config.Border + (startY * cellSize)
			results <- StripeResult{PasteX: pasteX, PasteY: pasteY, BaseImg: baseImage, GlowImg: glowImage}
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
		drawImageAt(baseLayer, res.BaseImg, res.PasteX, res.PasteY)
		if config.EnableGlow && res.GlowImg != nil {
			drawImageAt(glowLayer, res.GlowImg, res.PasteX, res.PasteY)
		}
	}
	baseImage := baseLayer
	if config.EnableGlow {
		//blurSigma := config.GlowRange + float64(config.LEDSize)
		blurSigma := config.GlowRange + float64(config.LEDSize)
		if blurSigma < 0.1 {
			blurSigma = 0.1
		}
		blurredGlow := imaging.Blur(glowLayer, blurSigma)
		compositeAdditiveRGBA(baseImage, ForceRGBA(blurredGlow), config)
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

func fillRGBA(dst *image.RGBA, col color.RGBA) {
	for y := 0; y < dst.Bounds().Dy(); y++ {
		off := y * dst.Stride
		for x := 0; x < dst.Bounds().Dx(); x++ {
			idx := off + x*4
			dst.Pix[idx+0] = col.R
			dst.Pix[idx+1] = col.G
			dst.Pix[idx+2] = col.B
			dst.Pix[idx+3] = col.A
		}
	}
}

func drawImageAt(dst *image.RGBA, src image.Image, x, y int) {
	rect := image.Rect(x, y, x+src.Bounds().Dx(), y+src.Bounds().Dy())
	draw.Draw(dst, rect, src, src.Bounds().Min, draw.Over)
}

// drawColoredLED: プリレンダリング形状を指定の色で描画
func drawColoredLED(dst *image.RGBA, template *image.RGBA, x, y int, col color.RGBA) {
	bounds := template.Bounds()
	dstBounds := dst.Bounds()

	tPix := template.Pix
	tStride := template.Stride
	dPix := dst.Pix
	dStride := dst.Stride

	for ty := 0; ty < bounds.Dy(); ty++ {
		tOff := (ty+bounds.Min.Y)*tStride + bounds.Min.X*4
		dy := y + ty
		if dy < dstBounds.Min.Y || dy >= dstBounds.Max.Y {
			continue
		}
		dOff := dy * dStride
		for tx := 0; tx < bounds.Dx(); tx++ {
			alpha := tPix[tOff+3] // テンプレートのアルファ値（マスクとして機能する）
			if alpha != 0 {
				dx := x + tx
				if dx >= dstBounds.Min.X && dx < dstBounds.Max.X {
					pixOff := dOff + dx*4
					if alpha == 255 {
						dPix[pixOff+0] = col.R
						dPix[pixOff+1] = col.G
						dPix[pixOff+2] = col.B
						dPix[pixOff+3] = col.A
					} else {
						// ★ここを修正：RGBもマスクのアルファ値(alpha)の割合で乗算する
						dPix[pixOff+0] = uint8(int(col.R) * int(alpha) / 255)
						dPix[pixOff+1] = uint8(int(col.G) * int(alpha) / 255)
						dPix[pixOff+2] = uint8(int(col.B) * int(alpha) / 255)
						dPix[pixOff+3] = uint8(int(col.A) * int(alpha) / 255)
					}
				}
			}
			tOff += 4
		}
	}
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

func applyLUTToRGBA(srcColor color.Color, lut *GammaLUT) color.RGBA {
	result := lut.Apply(srcColor)
	return color.RGBA{
		R: uint8(result.R >> 8),
		G: uint8(result.G >> 8),
		B: uint8(result.B >> 8),
		A: uint8(result.A >> 8),
	}
}

func compositeAdditiveRGBA(dst *image.RGBA, src *image.RGBA, config *Config) {
	glowAdditiveFactor := config.GlowStrength
	bounds := dst.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// ポインタの事前計算でキャッシュ効率を向上
	dstPix := dst.Pix
	srcPix := src.Pix
	dstStride := dst.Stride

	for y := 0; y < h; y++ {
		dstOff := y * dstStride
		for x := 0; x < w; x++ {
			pixOff := dstOff + x*4
			dstPix[pixOff+0] = clampByte(int(float64(dstPix[pixOff+0]) + float64(srcPix[pixOff+0])*glowAdditiveFactor))
			dstPix[pixOff+1] = clampByte(int(float64(dstPix[pixOff+1]) + float64(srcPix[pixOff+1])*glowAdditiveFactor))
			dstPix[pixOff+2] = clampByte(int(float64(dstPix[pixOff+2]) + float64(srcPix[pixOff+2])*glowAdditiveFactor))
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
