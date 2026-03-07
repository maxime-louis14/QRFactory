package qr

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strings"
	"testing"
)

// TestEncodeAlphanumeric teste la fonction EncodeAlphanumeric pour s'assurer qu'elle génère correctement le code binaire attendu pour une chaîne alphanumérique donnée.
func TestEncodeAlphanumeric(t *testing.T) {
	// Données d'entrée pour le test
	data := "TEST MAXIME"
	// Résultat attendu pour les données d'entrée
	expected := "1010010011110100001001110011010100011110001101101000000001110"

	// Appel de la fonction à tester
	fmt.Println("Testing EncodeAlphanumeric with input:", data)
	result, err := EncodeAlphanumeric(data)

	// Vérification des erreurs potentielles
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Comparaison du résultat obtenu avec le résultat attendu (en supprimant les espaces)
	if strings.ReplaceAll(result, " ", "") != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	} else {
		fmt.Println("Test passed! Result:", result)
	}
}

// TestEncodeByte teste la fonction EncodeByte pour s'assurer qu'elle génère correctement le code binaire attendu pour une chaîne de caractères donnée.
func TestEncodeByte(t *testing.T) {
	// Données d'entrée pour le test
	data := "HELLO"
	// Résultat attendu pour les données d'entrée
	expected := "0100100001000101010011000100110001001111"

	// Appel de la fonction à tester
	result, err := EncodeByte(data)

	// Vérification des erreurs potentielles
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Comparaison du résultat obtenu avec le résultat attendu
	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}

// TestEncodeNumeric teste la fonction EncodeNumeric pour s'assurer qu'elle génère correctement le code binaire attendu pour une chaîne numérique donnée.
func TestEncodeNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Nombres simples",
			input:    "12345",
			expected: "00011110110101101", // "123"→10 bits + "45"→7 bits
			wantErr:  false,
		},
		{
			name:     "Nombres avec zéros",
			input:    "00123",
			expected: "00000000010010111", // "001"→10 bits + "23"→7 bits
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeNumeric(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("EncodeNumeric() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestEncodeKanji teste la fonction EncodeKanji pour s'assurer qu'elle génère correctement le code binaire attendu pour une chaîne Kanji donnée.
func TestEncodeKanji(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "Kanji simple",
			input:       "世界",
			expected:    "01011101000100011011000101",
			shouldError: false,
		},
		{
			name:        "Konnichiwa en hiragana",
			input:       "こんにちは",
			expected:    "00001001100010000101110001000010100100100001001111110000101001101",
			shouldError: false,
		},
		{
			name:        "Mélange de Kanji et caractères non-Kanji",
			input:       "世界ABC",
			shouldError: true,
			expected:    "",
		},
		{
			name:        "Chaîne vide",
			input:       "",
			shouldError: true,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeKanji(tt.input)

			// Vérification de la gestion des erreurs
			if tt.shouldError {
				if err == nil {
					t.Errorf("EncodeKanji(%q) devrait retourner une erreur", tt.input)
				}
				return
			}

			// Vérification des cas valides
			if err != nil {
				t.Fatalf("EncodeKanji(%q) a retourné une erreur inattendue: %v", tt.input, err)
			}

			// Vérification du résultat
			if result != tt.expected {
				t.Errorf("EncodeKanji(%q)\nAttendu:  %s\nObtenu:   %s", tt.input, tt.expected, result)
			}
		})
	}
}

// TestToInt teste la fonction ToInt pour s'assurer qu'elle convertit correctement une chaîne en entier.
func TestToInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		name     string
		isError  bool // Indique si on s'attend à une erreur
	}{
		{"123", 123, "Nombre positif", false},
		{"-456", -456, "Nombre négatif", false},
		{"0", 0, "Zéro", false},
		{"999999999", 999999999, "Nombre très grand", false},
		{"-999999999", -999999999, "Nombre très petit", false},
		{"abc", 0, "Chaîne non numérique (doit retourner 0)", true}, // Attend une erreur
		{"2147483647", math.MaxInt32, "Plus grand nombre entier 32 bits", false},
		{"-2147483648", math.MinInt32, "Plus petit nombre entier 32 bits", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToInt(tt.input)

			// Vérification des erreurs
			if tt.isError {
				if err == nil {
					t.Error("Expected error for non-numeric input, but got nil")
				}
				return // Sortir du test si on attend une erreur
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("For input %s, expected %d but got %d", tt.input, tt.expected, result)
			}
		})
	}
}

func TestGenerateQRMatrix(t *testing.T) {
	tests := []struct {
		name    string
		version int
		data    string
		ecLevel string
		wantNil bool
	}{
		{
			name:    "Version 1 valide",
			version: 1,
			data:    "HELLO",
			ecLevel: "M",
			wantNil: false,
		},
		{
			name:    "Version invalide",
			version: 50,
			data:    "TEST",
			ecLevel: "M",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitGaloisField() // Initialisation nécessaire
			result := GenerateQRMatrix(tt.version, tt.data, tt.ecLevel)
			if (result == nil) != tt.wantNil {
				t.Errorf("GenerateQRMatrix() returned nil: %v, want nil: %v", result == nil, tt.wantNil)
			}
		})
	}
}

func TestSaveQRImage(t *testing.T) {
	// Créer une petite image test
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if (x+y)%2 == 0 {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}

	tempFile := "test_qr.png"
	defer os.Remove(tempFile) // Nettoyage après le test

	// Test de sauvegarde
	err := SaveQRImage(img, tempFile, 2)
	if err != nil {
		t.Errorf("SaveQRImage() error = %v", err)
	}

	// Vérifier que le fichier existe
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Errorf("SaveQRImage() failed to create file")
	}
}

func TestDetectDataType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Données numériques",
			input:    "12345",
			expected: "numeric",
		},
		{
			name:     "Données alphanumériques",
			input:    "HELLO123",
			expected: "alphanumeric",
		},
		{
			name:     "Données bytes",
			input:    "Hello, World!",
			expected: "byte",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectDataType(tt.input)
			if result != tt.expected {
				t.Errorf("DetectDataType(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsKanji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Kanji valide",
			input:    "世界",
			expected: true,
		},
		{
			name:     "Texte non-Kanji",
			input:    "Hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKanji(tt.input)
			if result != tt.expected {
				t.Errorf("IsKanji(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateQRCodeWithURL(t *testing.T) {
	url := "https://example.com"
	version := 1
	matrix := GenerateQRMatrix(version, url, "M")

	if matrix == nil {
		t.Error("GenerateQRMatrix returned nil")
	}

	// Test de sauvegarde de l'image
	tempFile := "test_qr_url.png"
	defer os.Remove(tempFile)

	err := SaveQRImage(matrix, tempFile, 2)
	if err != nil {
		t.Errorf("SaveQRImage() error = %v", err)
	}

	// Vérifier que le fichier existe
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Errorf("SaveQRImage() failed to create file")
	}
}

func TestURLDetection(t *testing.T) {
	url := "https://github.com/le-veilleur"

	// Tester la détection du type
	dataType := DetectDataType(url)
	if dataType != "alphanumeric" {
		t.Errorf("URL '%s' devrait être détectée comme alphanumérique, mais a été détectée comme %s", url, dataType)
	}

	// Vérifier que le QR code peut être généré correctement
	InitGaloisField() // Nécessaire pour la génération
	// Commencer avec la version 1 et laisser l'algorithme ajuster
	qrMatrix := GenerateQRMatrix(1, url, "M")
	if qrMatrix == nil {
		t.Errorf("La génération de QR code a échoué")
	}
}
