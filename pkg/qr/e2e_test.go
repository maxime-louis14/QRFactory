package qr_test

import (
	"image"
	"testing"

	"github.com/makiuchi-d/gozxing"
	gozxingqr "github.com/makiuchi-d/gozxing/qrcode"
	"qrfactory/pkg/qr"
)

// decodeMatrix décode une matrice QR image.RGBA en texte via gozxing (ZXing Go port).
func decodeMatrix(t *testing.T, matrix *image.RGBA) string {
	t.Helper()
	bmp, err := gozxing.NewBinaryBitmapFromImage(matrix)
	if err != nil {
		t.Fatalf("NewBinaryBitmapFromImage: %v", err)
	}
	result, err := gozxingqr.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		t.Fatalf("QR decode failed: %v", err)
	}
	return result.GetText()
}

func generateAndDecode(t *testing.T, data, ecLevel string) string {
	t.Helper()
	qr.InitGaloisField()
	matrix := qr.GenerateQRMatrix(1, data, ecLevel)
	if matrix == nil {
		t.Fatal("GenerateQRMatrix returned nil")
	}
	return decodeMatrix(t, matrix)
}

// --- E2E : Mode numérique ---

func TestE2E_Numeric_Short(t *testing.T) {
	got := generateAndDecode(t, "12345", "M")
	if got != "12345" {
		t.Errorf("got %q, want %q", got, "12345")
	}
}

func TestE2E_Numeric_AllDigits(t *testing.T) {
	got := generateAndDecode(t, "0123456789", "M")
	if got != "0123456789" {
		t.Errorf("got %q, want %q", got, "0123456789")
	}
}

// --- E2E : Mode alphanumérique ---

func TestE2E_Alphanumeric_Hello(t *testing.T) {
	got := generateAndDecode(t, "HELLO", "M")
	if got != "HELLO" {
		t.Errorf("got %q, want %q", got, "HELLO")
	}
}

func TestE2E_Alphanumeric_HelloWorld(t *testing.T) {
	got := generateAndDecode(t, "HELLO WORLD", "M")
	if got != "HELLO WORLD" {
		t.Errorf("got %q, want %q", got, "HELLO WORLD")
	}
}

func TestE2E_Alphanumeric_WithSpecialChars(t *testing.T) {
	input := "HTTPS://EXAMPLE.COM"
	qr.InitGaloisField()
	matrix := qr.GenerateQRMatrix(1, input, "H")
	if matrix == nil {
		t.Fatal("GenerateQRMatrix returned nil")
	}
	got := decodeMatrix(t, matrix)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

// --- E2E : Mode byte ---

func TestE2E_Byte_LowercaseText(t *testing.T) {
	input := "hello@world.com" // '@' hors table alphanumeric → byte mode
	qr.InitGaloisField()
	matrix := qr.GenerateQRMatrix(1, input, "M")
	if matrix == nil {
		t.Fatal("GenerateQRMatrix returned nil")
	}
	got := decodeMatrix(t, matrix)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

// --- E2E : Niveaux de correction ---

func TestE2E_ErrorLevels(t *testing.T) {
	for _, level := range []string{"L", "M", "Q", "H"} {
		t.Run("Level_"+level, func(t *testing.T) {
			got := generateAndDecode(t, "HELLO", level)
			if got != "HELLO" {
				t.Errorf("level %s: got %q, want %q", level, got, "HELLO")
			}
		})
	}
}

// --- E2E : Version 2 (motif d'alignement requis) ---

func TestE2E_Version2_URL(t *testing.T) {
	input := "HTTPS://GITHUB.COM/LE-VEILLEUR"
	qr.InitGaloisField()
	matrix := qr.GenerateQRMatrix(1, input, "M") // auto-upgrade vers V2
	if matrix == nil {
		t.Fatal("GenerateQRMatrix returned nil")
	}
	got := decodeMatrix(t, matrix)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestE2E_Version2_LongAlphanumeric(t *testing.T) {
	input := "THE QUICK BROWN FOX"
	qr.InitGaloisField()
	matrix := qr.GenerateQRMatrix(1, input, "M")
	if matrix == nil {
		t.Fatal("GenerateQRMatrix returned nil")
	}
	got := decodeMatrix(t, matrix)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}
