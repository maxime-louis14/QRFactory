package qr

import (
	"fmt"
	"sync"
)

var (
	// Variables globales pour les tables de Galois
	gfExp = make([]byte, 256)
	gfLog = make([]byte, 256)

	// Mutex pour protéger l'accès aux tables de Galois
	gfMutex sync.RWMutex

	// Variable pour suivre si l'initialisation a été effectuée
	gfInitialized bool
)

// InitGaloisField initialise les tables de multiplication du champ de Galois de manière thread-safe
func InitGaloisField() {
	gfMutex.Lock()
	defer gfMutex.Unlock()

	// Vérifier si l'initialisation a déjà été effectuée
	if gfInitialized {
		return
	}

	fmt.Println("Initialisation des tables de Galois...")

	// Initialisation des tables de multiplication du champ de Galois
	// IMPORTANT: Cette initialisation évite d'appeler GfMultiply de manière récursive
	x := byte(1)
	for i := 0; i < 256; i++ {
		gfExp[i] = x

		// Remplir la table de logarithmes
		if i < 255 {
			gfLog[x] = byte(i)
		}

		// Multiplication par 2 dans GF(2^8)
		// Cette implémentation est l'équivalent direct de GfMultiply(x, 2)
		// mais sans appeler la fonction qui pourrait créer une récursion infinie
		if (x & 0x80) != 0 {
			x = (x << 1) ^ 0x1D // Polynôme de réduction: x^8 + x^4 + x^3 + x^2 + 1 (0x1D)
		} else {
			x = x << 1
		}
	}

	gfInitialized = true
	fmt.Println("Initialisation des tables de Galois terminée.")
}

// GfMultiply effectue une multiplication dans le champ de Galois
func GfMultiply(x, y byte) byte {
	if x == 0 || y == 0 {
		return 0
	}

	if !gfInitialized {
		return 0
	}

	gfMutex.RLock()
	defer gfMutex.RUnlock()

	// Use int to avoid byte overflow when summing log values (max 254+254=508)
	return gfExp[(int(gfLog[x])+int(gfLog[y]))%255]
}

// GenerateReedSolomon génère les octets de correction d'erreur Reed-Solomon
func GenerateReedSolomon(data []byte, numECBytes int) []byte {
	generator := generateGenerator(numECBytes)
	remainder := make([]byte, numECBytes)

	for _, d := range data {
		factor := d ^ remainder[0]
		copy(remainder, remainder[1:])
		remainder[len(remainder)-1] = 0

		for i := 0; i < len(remainder); i++ {
			remainder[i] ^= GfMultiply(generator[i], factor)
		}
	}

	return remainder
}

// GenerateErrorCorrection génère les codes de correction d'erreur pour les données
func GenerateErrorCorrection(data []byte, level string) []byte {
	var numECBytes int
	switch level {
	case "L":
		numECBytes = 7
	case "M":
		numECBytes = 10
	case "Q":
		numECBytes = 13
	case "H":
		numECBytes = 17
	default:
		numECBytes = 10 // Par défaut niveau M
	}

	return GenerateReedSolomon(data, numECBytes)
}

// generateGenerator génère le polynôme générateur pour Reed-Solomon
// g(x) = (x + α^0)(x + α^1)...(x + α^(degree-1))
// Retourne les coefficients non-dominants [g_{degree-1}, ..., g_0]
func generateGenerator(degree int) []byte {
	gen := make([]byte, degree+1)
	gen[0] = 1

	for i := 0; i < degree; i++ {
		alpha := gfExp[i]
		// Multiply gen by (x + alpha): gen[j] ^= alpha * gen[j-1] for j from i+1 down to 1
		for j := i + 1; j > 0; j-- {
			gen[j] ^= GfMultiply(gen[j-1], alpha)
		}
	}

	return gen[1:]
}

// Polynômes générateurs pour différents niveaux de correction d'erreur
var GeneratorPolynomials = map[string][]int{
	"L": {1, 1},          // 7% de correction
	"M": {1, 1, 1},       // 15% de correction
	"Q": {1, 1, 1, 1},    // 25% de correction
	"H": {1, 1, 1, 1, 1}, // 30% de correction
}

// AddErrorCorrectionEC ajoute les codes de correction d'erreur aux données
func AddErrorCorrectionEC(data string, ecLevel string, version int) string {
	InitGaloisField()

	ecWords := calculateECWords(version, ecLevel)

	// Convert binary string to bytes
	dataBytes := make([]byte, 0, len(data)/8)
	for i := 0; i+8 <= len(data); i += 8 {
		dataBytes = append(dataBytes, binaryStringToByte(data[i:i+8]))
	}

	// Generate EC bytes using proper Reed-Solomon
	ecBytes := GenerateReedSolomon(dataBytes, ecWords)

	result := data
	for _, b := range ecBytes {
		result += fmt.Sprintf("%08b", b)
	}

	return result
}

// Convertit une chaîne binaire en byte
func binaryStringToByte(binary string) byte {
	var result byte
	for i := 0; i < len(binary); i++ {
		if binary[i] == '1' {
			result |= 1 << uint(7-i)
		}
	}
	return result
}

// totalCodewordsTable contient le nombre total de codewords par version (ISO/IEC 18004 Table 9)
var totalCodewordsTable = []int{
	26, 44, 70, 100, 134, 172, 196, 242, 292, 346,
	404, 466, 532, 581, 655, 733, 815, 901, 991, 1085,
	1156, 1258, 1364, 1474, 1588, 1706, 1828, 1921, 2051, 2185,
	2323, 2465, 2611, 2761, 2876, 3034, 3196, 3362, 3532, 3706,
}

// DataCodewords retourne le nombre de codewords de données pour une version et un niveau EC
func DataCodewords(version int, ecLevel string) int {
	if version < 1 || version > 40 {
		return 0
	}
	return totalCodewordsTable[version-1] - calculateECWords(version, ecLevel)
}

// calculateECWords retourne le nombre total de mots de code de correction d'erreur
// pour une version et un niveau donnés (source: ISO/IEC 18004)
func calculateECWords(version int, ecLevel string) int {
	ecWordsTable := map[string][]int{
		// Index 0 = version 1, index 39 = version 40
		"L": {7, 10, 15, 20, 26, 36, 40, 48, 60, 72, 80, 96, 104, 120, 132, 144, 168, 180, 196, 224, 224, 252, 270, 300, 312, 336, 360, 390, 420, 450, 480, 510, 540, 570, 570, 600, 630, 660, 720, 750},
		"M": {10, 16, 26, 36, 48, 64, 72, 88, 110, 130, 150, 176, 198, 216, 240, 280, 308, 338, 364, 416, 442, 476, 504, 560, 588, 644, 700, 728, 784, 812, 868, 924, 980, 1036, 1064, 1120, 1204, 1260, 1316, 1372},
		"Q": {13, 22, 36, 52, 72, 96, 108, 132, 160, 192, 224, 260, 288, 320, 360, 408, 448, 504, 546, 600, 644, 690, 750, 810, 870, 952, 1020, 1050, 1140, 1200, 1290, 1350, 1440, 1530, 1590, 1680, 1770, 1860, 1950, 2040},
		"H": {17, 28, 44, 64, 88, 112, 130, 156, 192, 224, 264, 308, 352, 384, 432, 480, 532, 588, 650, 700, 750, 816, 900, 960, 1050, 1110, 1200, 1260, 1350, 1440, 1530, 1620, 1710, 1800, 1890, 1980, 2100, 2220, 2310, 2430},
	}

	if table, ok := ecWordsTable[ecLevel]; ok {
		if version >= 1 && version <= 40 {
			return table[version-1]
		}
	}
	return 10
}

// Génère les bytes de correction d'erreur
func generateECBytes(data []byte, numECBytes int) []byte {
	// Initialiser le tableau de correction d'erreur
	ecBytes := make([]byte, numECBytes)

	// Copier les données dans le tableau de correction d'erreur
	for i := 0; i < len(data); i++ {
		feedback := data[i]
		for j := 0; j < len(ecBytes); j++ {
			temp := ecBytes[j]
			ecBytes[j] = byte(int(feedback) ^ int(temp))
			feedback = temp
		}
	}

	return ecBytes
}
