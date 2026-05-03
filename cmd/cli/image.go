package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"led-image-gen/processor"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var imageCmd = &cobra.Command{
	Use:   "image [input]",
	Short: "Convert an image to LED display style.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}
		var config processor.Config
		err = viper.Unmarshal(&config)
		if err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
		colorStr := viper.GetString("off-light-color")
		offColor, err := parseRGBA(colorStr)
		if err != nil {
			return err
		}
		config.OffLightColor = offColor
		inFile := args[0]
		if !cmd.Flags().Changed("workers") {
			fmt.Printf("Auto-detected max workers: using %d threads.\n", config.MaxWorkers)
		}
		outFile := viper.GetString("output")
		if !cmd.Flags().Changed("output") {
			ext := filepath.Ext(inFile)
			base := strings.TrimSuffix(filepath.Base(inFile), ext)
			outFile = base + "_led.png"
		}
		file, err := os.Open(inFile)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()

		src, _, err := image.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode input image: %w", err)
		}

		result, err := processor.GenerateLEDImage(src, &config)
		if err != nil {
			return fmt.Errorf("failed to process image: %w", err)
		}

		outFileHandle, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFileHandle.Close()

		err = png.Encode(outFileHandle, result)
		if err != nil {
			return fmt.Errorf("failed to encode output image: %w", err)
		}

		fmt.Printf("LED image saved to %s\n", outFile)

		return nil
	},
}

func init() {
	defaultMaxWorkers := runtime.NumCPU()
	defCfg := processor.DefaultConfig()
	rootCmd.AddCommand(imageCmd)
	imageCmd.Flags().StringP("output", "o", "<input_file_name>_led.png", "Output file name")
	imageCmd.Flags().IntP("border", "b", defCfg.Border, "Border size in pixels")
	imageCmd.Flags().IntP("led-size", "s", defCfg.LEDSize, "LED size in pixels")
	imageCmd.Flags().IntP("led-gap", "g", defCfg.LEDGap, "Gap between LEDs in pixels")
	imageCmd.Flags().Float64P("led-gamma", "", defCfg.LEDGamma, "Gamma correction for LED brightness")
	imageCmd.Flags().Float64P("led-exposure", "", defCfg.LEDExposure, "Exposure adjustment for LED brightness")
	imageCmd.Flags().IntP("max-workers", "w", defaultMaxWorkers, "Maximum number of worker threads")
	imageCmd.Flags().StringP("led-shape", "c", string(defCfg.LEDShape), "Shape of the LEDs (circle or square)")
	imageCmd.Flags().BoolP("enable-glow", "", defCfg.EnableGlow, "Enable glow effect around LEDs")
	imageCmd.Flags().Float64P("glow-range", "", defCfg.GlowRange, "Range of the glow effect in pixels")
	imageCmd.Flags().Float64P("glow-strength", "", defCfg.GlowStrength, "Strength of the glow effect")
	imageCmd.Flags().Float64P("glow-gamma", "", defCfg.GlowGamma, "Gamma correction for glow brightness")
	imageCmd.Flags().Float64P("glow-exposure", "", defCfg.GlowExposure, "Exposure adjustment for glow brightness")
	imageCmd.Flags().StringP("off-light-color", "", fmt.Sprintf("rgba(%d,%d,%d,%d)",
		defCfg.OffLightColor.R, defCfg.OffLightColor.G,
		defCfg.OffLightColor.B, defCfg.OffLightColor.A), "Color of the off LEDs in RGBA format (e.g., rgba(0,0,0,255))")
}

func parseRGBA(colorStr string) (color.RGBA, error) {
	var r, g, b, a uint8

	_, err := fmt.Sscanf(colorStr, "rgba(%d,%d,%d,%d)", &r, &g, &b, &a)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("the color format is invalid. Please specify it in the format 'rgba(0,0,0,255)': %w", err)
	}

	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}
