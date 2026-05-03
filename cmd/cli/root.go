package main

import "github.com/spf13/cobra"

// rootCmd は引数なしで "led-gen" と叩かれた時の挙動
var rootCmd = &cobra.Command{
	Use:   "led-gen",
	Short: "led-image-gen is a fast LED display style image converter.",
	Long:  `led-image-gen is a fast LED display style image converter.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
