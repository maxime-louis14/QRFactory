package qr

import (
	"image"
	"image/color"
	"testing"
)

// --- Reed-Solomon ---

// TestGenerateReedSolomon vérifie les EC bytes pour l'exemple "HELLO WORLD" V1-M
// (valeurs de référence issues de thonky.com / ISO 18004)
func TestGenerateReedSolomon_HelloWorld(t *testing.T) {
	InitGaloisField()

	// Data codewords pour "HELLO WORLD" en mode alphanumérique, V1-M
	data := []byte{32, 91, 11, 120, 209, 114, 220, 77, 67, 64, 236, 17, 236, 17, 236, 17}
	expected := []byte{196, 35, 39, 119, 235, 215, 231, 226, 93, 23}

	got := GenerateReedSolomon(data, 10)

	if len(got) != len(expected) {
		t.Fatalf("length: got %d, want %d", len(got), len(expected))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("EC[%d]: got %d, want %d", i, got[i], expected[i])
		}
	}
}

func TestGenerateReedSolomon_ZeroData(t *testing.T) {
	InitGaloisField()
	ec := GenerateReedSolomon([]byte{0, 0, 0, 0}, 4)
	for i, b := range ec {
		if b != 0 {
			t.Errorf("EC[%d] = %d, want 0 for all-zero data", i, b)
		}
	}
}

// --- Format Information ---

func isBlackPixel(matrix *image.RGBA, x, y int) bool {
	r, _, _, _ := matrix.At(x, y).RGBA()
	return r < 0x8000
}

// TestAddFormatInfo_FirstCopy vérifie que les bits de la première copie
// sont placés aux bonnes positions pour M/mask-0 (string: "101010000010010")
func TestAddFormatInfo_FirstCopy(t *testing.T) {
	size := 21
	matrix := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			matrix.Set(x, y, color.White)
		}
	}

	AddFormatInfo(matrix, "M", 0)

	// Format string "101010000010010" → bit[i]='1' = noir, '0' = blanc
	formatStr := "101010000010010"
	wantBlack := func(i int) bool { return formatStr[i] == '1' }

	// Première copie – partie horizontale (y=8)
	horizontalPos := [][2]int{{0, 8}, {1, 8}, {2, 8}, {3, 8}, {4, 8}, {5, 8}, {7, 8}, {8, 8}}
	for i, pos := range horizontalPos {
		x, y := pos[0], pos[1]
		got := isBlackPixel(matrix, x, y)
		if got != wantBlack(i) {
			t.Errorf("first copy bit[%d] at (%d,%d): got black=%v, want %v", i, x, y, got, wantBlack(i))
		}
	}

	// Première copie – partie verticale (x=8)
	verticalRows := []int{7, 5, 4, 3, 2, 1, 0} // bits 8-14
	for j, row := range verticalRows {
		bit := 8 + j
		got := isBlackPixel(matrix, 8, row)
		if got != wantBlack(bit) {
			t.Errorf("first copy bit[%d] at (8,%d): got black=%v, want %v", bit, row, got, wantBlack(bit))
		}
	}
}

// TestAddFormatInfo_SecondCopy vérifie la deuxième copie (haut-droit et bas-gauche)
func TestAddFormatInfo_SecondCopy(t *testing.T) {
	size := 21
	matrix := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			matrix.Set(x, y, color.White)
		}
	}

	AddFormatInfo(matrix, "M", 0)

	formatStr := "101010000010010"
	wantBlack := func(i int) bool { return formatStr[i] == '1' }

	// Deuxième copie – 8 bits horizontaux (y=8, x de size-1 vers size-8)
	for i := 0; i < 8; i++ {
		x := size - 1 - i
		got := isBlackPixel(matrix, x, 8)
		if got != wantBlack(i) {
			t.Errorf("second copy bit[%d] at (%d,8): got black=%v, want %v", i, x, got, wantBlack(i))
		}
	}

	// Deuxième copie – 7 bits verticaux (x=8, y de size-7 à size-1)
	for i := 8; i < 15; i++ {
		y := size - 7 + (i - 8)
		got := isBlackPixel(matrix, 8, y)
		if got != wantBlack(i) {
			t.Errorf("second copy bit[%d] at (8,%d): got black=%v, want %v", i, y, got, wantBlack(i))
		}
	}
}

// --- isValidDataPosition ---

func TestIsValidDataPosition_FinderPatterns(t *testing.T) {
	size := 21
	// Coins des finder patterns → invalides
	for _, pos := range [][2]int{{0, 0}, {0, 8}, {8, 0}, {size - 1, 0}, {0, size - 1}} {
		if isValidDataPosition(pos[0], pos[1], size) {
			t.Errorf("(%d,%d) should not be valid (finder pattern)", pos[0], pos[1])
		}
	}
}

func TestIsValidDataPosition_TimingPatterns(t *testing.T) {
	size := 21
	for x := 8; x < size-8; x++ {
		if isValidDataPosition(x, 6, size) {
			t.Errorf("(%d,6) should not be valid (timing pattern)", x)
		}
		if isValidDataPosition(6, x, size) {
			t.Errorf("(6,%d) should not be valid (timing pattern)", x)
		}
	}
}

func TestIsValidDataPosition_DataArea(t *testing.T) {
	size := 21
	// Centre de la zone de données → valide
	validPositions := [][2]int{{10, 10}, {15, 15}, {18, 10}, {10, 18}}
	for _, pos := range validPositions {
		if !isValidDataPosition(pos[0], pos[1], size) {
			t.Errorf("(%d,%d) should be valid data position", pos[0], pos[1])
		}
	}
}

func TestIsValidDataPosition_AlignmentPattern_V2(t *testing.T) {
	size := 25 // version 2
	// Centre du motif d'alignement à (18,18) → invalide
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			x, y := 18+dx, 18+dy
			if isValidDataPosition(x, y, size) {
				t.Errorf("(%d,%d) inside alignment pattern should not be valid", x, y)
			}
		}
	}
	// Juste à l'extérieur → valide (si pas une autre zone réservée)
	if !isValidDataPosition(15, 15, size) {
		t.Errorf("(15,15) should be valid data position for V2")
	}
}

// --- Motifs d'alignement ---

func TestAlignmentPattern_V1_None(t *testing.T) {
	// V1 n'a pas de motif d'alignement
	matrix := image.NewRGBA(image.Rect(0, 0, 21, 21))
	for y := 0; y < 21; y++ {
		for x := 0; x < 21; x++ {
			matrix.Set(x, y, color.White)
		}
	}
	AddAlignmentPatterns(matrix, 21)
	// Vérifier que la zone (18,18)±2 est toujours blanche pour V1
	// (pas de motif d'alignement en V1)
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			r, _, _, _ := matrix.At(18+dx, 18+dy).RGBA()
			if r < 0x8000 {
				t.Errorf("V1 should have no alignment at (%d,%d), but pixel is black", 18+dx, 18+dy)
			}
		}
	}
}

func TestAlignmentPattern_V2_Present(t *testing.T) {
	// V2 doit avoir un motif d'alignement centré en (18,18)
	matrix := image.NewRGBA(image.Rect(0, 0, 25, 25))
	for y := 0; y < 25; y++ {
		for x := 0; x < 25; x++ {
			matrix.Set(x, y, color.White)
		}
	}
	AddAlignmentPatterns(matrix, 25)

	// Le bord (±2) doit être noir
	center := [2]int{18, 18}
	for _, dx := range []int{-2, 2} {
		for dy := -2; dy <= 2; dy++ {
			if !isBlackPixel(matrix, center[0]+dx, center[1]+dy) {
				t.Errorf("alignment border at (%d,%d) should be black", center[0]+dx, center[1]+dy)
			}
		}
	}
	// Le centre doit être noir
	if !isBlackPixel(matrix, 18, 18) {
		t.Error("alignment center (18,18) should be black")
	}
}

// --- Pipeline complet ---

func TestGenerateQRMatrix_ReturnsNilOnInvalidVersion(t *testing.T) {
	InitGaloisField()
	if GenerateQRMatrix(50, "HELLO", "M") != nil {
		t.Error("version 50 should return nil")
	}
	// version 0 auto-upgrades to the minimum required version, so it returns a valid matrix
	if GenerateQRMatrix(0, "HELLO", "M") == nil {
		t.Error("version 0 should auto-upgrade and return a valid matrix")
	}
}

func TestGenerateQRMatrix_AutoUpgradesVersion(t *testing.T) {
	InitGaloisField()
	// "HELLO WORLD" ne rentre pas en V1-H → doit auto-upgrader
	matrix := GenerateQRMatrix(1, "HELLO WORLD", "H")
	if matrix == nil {
		t.Fatal("should auto-upgrade and return a valid matrix")
	}
	// V2 = 25×25
	bounds := matrix.Bounds()
	if bounds.Max.X != 25 || bounds.Max.Y != 25 {
		t.Errorf("expected 25×25 matrix for V2, got %dx%d", bounds.Max.X, bounds.Max.Y)
	}
}
