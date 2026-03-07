package main

import (
	"fmt"
	"os"
	"qrfactory/internal/model"
	"qrfactory/pkg/config"
	"qrfactory/pkg/qr"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfg       *config.QRConfig
	scale     int
	quietZone int
)

var rootCmd = &cobra.Command{
	Use:   "qrfactory",
	Short: "QRFactory - Advanced QR Code Generator",
	Long: `QRFactory is a powerful command-line tool for generating QR codes.
It supports various data types, error correction levels, and customization options.
See https://github.com/le-veilleur for more information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`
  ██████╗ ██████╗ ███████╗ █████╗  ██████╗████████╗ ██████╗ ██████╗ ██╗   ██╗
 ██╔═══██╗██╔══██╗██╔════╝██╔══██╗██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗╚██╗ ██╔╝
 ██║   ██║██████╔╝█████╗  ███████║██║        ██║   ██║   ██║██████╔╝ ╚████╔╝
 ██║▄▄ ██║██╔══██╗██╔══╝  ██╔══██║██║        ██║   ██║   ██║██╔══██╗  ╚██╔╝
 ╚██████╔╝██║  ██║██║     ██║  ██║╚██████╗   ██║   ╚██████╔╝██║  ██║   ██║
  ╚══▀▀═╝ ╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝ ╚═════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝   ╚═╝
                         QR Code Generator — ISO/IEC 18004
`)
		start := time.Now()

		// Validate configuration
		fmt.Println("Validating configuration...")
		if err := config.ValidateConfig(cfg); err != nil {
			fmt.Printf("Configuration error: %v\n", err)
			os.Exit(1)
		}

		// Initialize Galois Field
		fmt.Println("Initializing Galois Field...")
		qr.InitGaloisField()
		fmt.Printf("Initialization completed in %v\n", time.Since(start))

		// Create QRCode model
		fmt.Println("Creating QRCode model...")
		qrCode := model.NewQRCode(cfg.Data, cfg.Version, cfg.ErrorCorrectionLevel)

		// Detect data type
		fmt.Println("Detecting data type...")
		dataType := qr.DetectDataType(cfg.Data)
		fmt.Printf("Detected data type: %s\n", dataType)

		// Calculate minimum required version
		fmt.Println("Calculating minimum required version...")
		minVersion, err := qr.CalculateMinVersionForDataType(cfg.Data, dataType)
		if err != nil {
			fmt.Printf("Error calculating version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Calculated minimum version: %d\n", minVersion)

		// Use minimum version if necessary
		if cfg.Version < minVersion {
			fmt.Printf("Version %d is too small for the data. Using version %d.\n", cfg.Version, minVersion)
			cfg.Version = minVersion
			qrCode.Version = minVersion
			qrCode.Size = minVersion*4 + 17
		}

		// Generate QR code matrix
		fmt.Println("Generating QR matrix...")
		genStart := time.Now()
		matrix := qr.GenerateQRMatrix(cfg.Version, cfg.Data, cfg.ErrorCorrectionLevel)
		fmt.Printf("Matrix generation completed in %v\n", time.Since(genStart))

		if matrix == nil {
			fmt.Println("Error generating QR code")
			os.Exit(1)
		}

		// Associate matrix with the model
		fmt.Println("Associating matrix with the model...")
		qrCode.SetMatrix(matrix)

		// Save image
		fmt.Println("Saving image...")
		saveStart := time.Now()
		if err := qr.SaveQRImageWithQuietZone(matrix, cfg.OutputFile, scale, quietZone); err != nil {
			fmt.Printf("Error saving image: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Save completed in %v\n", time.Since(saveStart))

		fmt.Printf("QR code successfully generated: %s\n", cfg.OutputFile)
		fmt.Printf("Total execution time: %v\n", time.Since(start))
	},
}

// Version command to display the application version
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("QRFactory 1.0.0")
	},
}

func init() {
	// Create default configuration
	cfg = config.NewDefaultConfig()

	// Add version command
	rootCmd.AddCommand(versionCmd)

	// Configure command flags
	rootCmd.Flags().StringVarP(&cfg.Data, "data", "d", "", "Data to encode in the QR code")
	rootCmd.Flags().IntVarP(&cfg.Version, "version", "v", cfg.Version, "QR code version (1-40)")
	rootCmd.Flags().StringVarP(&cfg.ErrorCorrectionLevel, "error-correction", "e", "H", "Error correction level (L, M, Q, H)")
	rootCmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", "qrcode.png", "Output file path")
	rootCmd.Flags().StringVar(&cfg.BackgroundColor, "bg-color", cfg.BackgroundColor, "Background color")
	rootCmd.Flags().StringVar(&cfg.ForegroundColor, "fg-color", cfg.ForegroundColor, "Module color")
	rootCmd.Flags().IntVarP(&scale, "scale", "s", 30, "Image scale (default: 30)")
	rootCmd.Flags().IntVarP(&quietZone, "quiet-zone", "q", 4, "Quiet zone width in modules (default: 4)")

	// Mark required flags
	rootCmd.MarkFlagRequired("data")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
