package qr

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"
	"sync"
)

// CalculateMinVersion calcule la version minimale nécessaire pour les données
// Retourne la version minimale ou une erreur si les données sont trop longues
func CalculateMinVersion(data string) (int, error) {
	// Calculer la taille totale nécessaire en bits
	totalBits := 4 + 8 + (len(data) * 8) // Mode (4) + Length (8) + Data (8 par caractère)

	// Table de capacité en bits pour le mode byte avec niveau de correction M
	capacityTable := []int{
		128, 224, 352, 512, 688, 864, 992, 1232, 1456, 1728,
		2032, 2320, 2672, 2920, 3320, 3624, 4056, 4504, 5016, 5352,
		5712, 6256, 6880, 7312, 8000, 8496, 9024, 9544, 10136, 10984,
		11640, 12328, 13048, 13800, 14496, 15312, 15936, 16816, 17728, 18672,
	}

	for version := 1; version <= 40; version++ {
		if capacityTable[version-1] >= totalBits {
			return version, nil
		}
	}

	return 0, fmt.Errorf("impossible de stocker les données même avec la version maximale")
}

// CalculateMinVersionForDataType calcule la version minimale nécessaire pour les données selon le type
func CalculateMinVersionForDataType(data string, dataType string) (int, error) {
	var totalBits int

	// Calcule les bits nécessaires selon le type de données
	switch dataType {
	case "numeric":
		// Mode (4) + Length (10 pour version 1-9) + Data (~3.33 bits par caractère)
		dataLength := len(data)
		dataBits := (dataLength / 3) * 10
		switch dataLength % 3 {
		case 1:
			dataBits += 4 // 1 chiffre = 4 bits
		case 2:
			dataBits += 7 // 2 chiffres = 7 bits
		}
		totalBits = 4 + 10 + dataBits
	case "alphanumeric":
		// Mode (4) + Length (9 pour version 1-9) + Data (~5.5 bits par caractère)
		dataLength := len(data)
		dataBits := (dataLength / 2) * 11
		if dataLength%2 == 1 {
			dataBits += 6 // 1 caractère seul = 6 bits
		}
		totalBits = 4 + 9 + dataBits
	case "kanji":
		// Mode (4) + Length (8 pour version 1-9) + Data (13 bits par caractère)
		totalBits = 4 + 8 + (len(data) * 13)
	default: // "byte" ou autre
		// Mode (4) + Length (8 pour version 1-9) + Data (8 bits par caractère)
		totalBits = 4 + 8 + (len(data) * 8)
	}

	// Table de capacité en bits pour différents modes avec niveau de correction M
	capacityTable := []int{
		128, 224, 352, 512, 688, 864, 992, 1232, 1456, 1728,
		2032, 2320, 2672, 2920, 3320, 3624, 4056, 4504, 5016, 5352,
		5712, 6256, 6880, 7312, 8000, 8496, 9024, 9544, 10136, 10984,
		11640, 12328, 13048, 13800, 14496, 15312, 15936, 16816, 17728, 18672,
	}

	for version := 1; version <= 40; version++ {
		if capacityTable[version-1] >= totalBits {
			return version, nil
		}
	}

	return 0, fmt.Errorf("impossible de stocker les données %s de type %s même avec la version maximale", data, dataType)
}

// GenerateQRMatrix génère la matrice QR pour les données fournies
func GenerateQRMatrix(version int, data string, errorCorrectionLevel string) *image.RGBA {
	// Valider le niveau de correction d'erreur
	if errorCorrectionLevel != "L" && errorCorrectionLevel != "M" &&
		errorCorrectionLevel != "Q" && errorCorrectionLevel != "H" {
		// Par défaut, utiliser le niveau M
		errorCorrectionLevel = "M"
		fmt.Printf("Niveau de correction d'erreur invalide, utilisation du niveau M (15%%)\n")
	}

	// Détecter le type de données
	dataType := DetectDataType(data)
	fmt.Printf("Type de données détecté: %s\n", dataType)

	// Calculer la version minimale nécessaire
	minVersion, minVersionErr := CalculateMinVersionForDataType(data, dataType)
	if minVersionErr != nil {
		fmt.Printf("ERREUR: %v\n", minVersionErr)
		return nil
	}

	// Utiliser la version minimale si la version fournie est trop petite
	if version < minVersion {
		fmt.Printf("La version %d est trop petite pour les données. Utilisation de la version %d.\n", version, minVersion)
		version = minVersion
	}

	// Vérifier la validité de la version
	if version < 1 || version > 40 {
		return nil
	}

	// Préparer l'encodage des données
	var encodedData strings.Builder
	var modeIndicator, lengthBits string
	var encodedBits string
	var encodingErr error

	// Configurer selon le type de données
	switch dataType {
	case "numeric":
		modeIndicator = "0001" // Indicateur mode numérique
		// Taille du compteur selon la version
		if version >= 1 && version <= 9 {
			lengthBits = fmt.Sprintf("%010b", len(data)) // 10 bits pour versions 1-9
		} else {
			lengthBits = fmt.Sprintf("%012b", len(data)) // 12 bits pour versions 10-40
		}
		encodedBits, encodingErr = EncodeNumeric(data)
	case "alphanumeric":
		modeIndicator = "0010" // Indicateur mode alphanumérique
		// Taille du compteur selon la version
		if version >= 1 && version <= 9 {
			lengthBits = fmt.Sprintf("%09b", len(data)) // 9 bits pour versions 1-9
		} else {
			lengthBits = fmt.Sprintf("%011b", len(data)) // 11 bits pour versions 10-40
		}
		encodedBits, encodingErr = EncodeAlphanumeric(data)
	case "kanji":
		modeIndicator = "1000" // Indicateur mode kanji
		// Taille du compteur selon la version
		if version >= 1 && version <= 9 {
			lengthBits = fmt.Sprintf("%08b", len(data)) // 8 bits pour versions 1-9
		} else {
			lengthBits = fmt.Sprintf("%010b", len(data)) // 10 bits pour versions 10-40
		}
		encodedBits, encodingErr = EncodeKanji(data)
	default: // "byte" ou autre
		modeIndicator = "0100" // Indicateur mode byte
		// Taille du compteur selon la version
		if version >= 1 && version <= 9 {
			lengthBits = fmt.Sprintf("%08b", len(data)) // 8 bits pour versions 1-9
		} else {
			lengthBits = fmt.Sprintf("%016b", len(data)) // 16 bits pour versions 10-40
		}
		encodedBits, encodingErr = EncodeByte(data)
	}

	if encodingErr != nil {
		fmt.Printf("Erreur d'encodage: %v\n", encodingErr)
		return nil
	}

	// Ajouter l'indicateur de mode
	encodedData.WriteString(modeIndicator)

	// Ajouter la longueur des données
	encodedData.WriteString(lengthBits)

	// Ajouter les données encodées
	encodedData.WriteString(encodedBits)

	// Calculer la capacité disponible
	size := version*4 + 17
	matrix := image.NewRGBA(image.Rect(0, 0, size, size))

	// Initialiser la matrice en blanc
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			matrix.Set(x, y, color.White)
		}
	}

	// Ajouter les motifs de repérage
	AddFinderPatterns(matrix)
	AddSeparators(matrix)
	AddAlignmentPatterns(matrix, size)
	AddTimingPatterns(matrix)

	// Calculer la capacité disponible
	capacity := calculateAvailableCapacity(size, errorCorrectionLevel)

	// Vérifier si les données encodées dépassent la capacité
	if encodedData.Len() > capacity {
		// Si c'est le cas, essayer avec une version supérieure
		newVersion := version + 1
		fmt.Printf("ATTENTION: Les données encodées (%d bits) dépassent la capacité (%d bits)\n", encodedData.Len(), capacity)
		fmt.Printf("Tentative avec la version %d...\n", newVersion)
		return GenerateQRMatrix(newVersion, data, errorCorrectionLevel) // Appel récursif avec version supérieure
	}

	// 1. Ajouter le terminateur (max 4 zéros)
	termLen := 4
	if remaining := capacity - encodedData.Len(); remaining < 4 {
		termLen = remaining
	}
	if termLen < 0 {
		termLen = 0
	}
	for i := 0; i < termLen; i++ {
		encodedData.WriteString("0")
	}

	// 2. Aligner sur un octet
	for encodedData.Len()%8 != 0 {
		encodedData.WriteString("0")
	}

	// 3. Remplir avec les octets de bourrage (0xEC=11101100, 0x11=00010001)
	padByte := 0
	for encodedData.Len() < capacity {
		if padByte%2 == 0 {
			encodedData.WriteString("11101100")
		} else {
			encodedData.WriteString("00010001")
		}
		padByte++
	}

	// Ajouter la correction d'erreur
	finalData := AddErrorCorrection(encodedData.String(), errorCorrectionLevel, version)

	fmt.Printf("Niveau de correction: %s\n", errorCorrectionLevel)
	fmt.Printf("Capacité disponible: %d bits\n", capacity)
	fmt.Printf("Longueur des données encodées: %d bits\n", len(finalData))
	fmt.Printf("Données encodées: %s\n", finalData)

	// Placer les données
	PlaceData(matrix, finalData)

	// Appliquer le meilleur masque
	bestScore := math.MaxInt32
	var bestMatrix *image.RGBA
	bestMask := 0

	fmt.Println("Évaluation des masques:")
	for mask := 0; mask < 8; mask++ {
		maskedMatrix := ApplyMask(matrix, mask)
		score := EvaluateMask(maskedMatrix)
		fmt.Printf("  Masque %d: score %d\n", mask, score)

		if score < bestScore {
			bestScore = score
			bestMatrix = maskedMatrix
			bestMask = mask
		}
	}

	fmt.Printf("Meilleur masque sélectionné: %d (score: %d)\n", bestMask, bestScore)

	if bestMatrix != nil {
		matrix = bestMatrix
		// Ajouter les informations de format après avoir appliqué le masque
		AddFormatInfo(matrix, errorCorrectionLevel, bestMask)
	}

	return matrix
}

// EvaluateMask évalue la qualité d'un masque selon les règles de pénalité du QR code
func EvaluateMask(matrix *image.RGBA) int {
	score := 0
	size := matrix.Bounds().Max.X

	// Règle 1: Pénalité pour 5+ modules de même couleur consécutifs
	score += evaluateRule1(matrix, size)

	// Règle 2: Pénalité pour les blocs de couleur 2x2
	score += evaluateRule2(matrix, size)

	// Règle 3: Pénalité pour motifs spécifiques ressemblant aux finder patterns
	score += evaluateRule3(matrix, size)

	// Règle 4: Équilibre entre modules noirs et blancs
	score += evaluateRule4(matrix, size)

	return score
}

// evaluateRule1 calcule la pénalité pour les séquences de modules de même couleur
func evaluateRule1(matrix *image.RGBA, size int) int {
	penalty := 0

	// Vérifier les lignes horizontales
	for y := 0; y < size; y++ {
		count := 1
		color := isBlack(matrix.At(0, y))

		for x := 1; x < size; x++ {
			if isBlack(matrix.At(x, y)) == color {
				count++
			} else {
				if count >= 5 {
					penalty += 3 + (count - 5)
				}
				count = 1
				color = isBlack(matrix.At(x, y))
			}
		}

		// Vérifier la séquence finale de la ligne
		if count >= 5 {
			penalty += 3 + (count - 5)
		}
	}

	// Vérifier les colonnes verticales
	for x := 0; x < size; x++ {
		count := 1
		color := isBlack(matrix.At(x, 0))

		for y := 1; y < size; y++ {
			if isBlack(matrix.At(x, y)) == color {
				count++
			} else {
				if count >= 5 {
					penalty += 3 + (count - 5)
				}
				count = 1
				color = isBlack(matrix.At(x, y))
			}
		}

		// Vérifier la séquence finale de la colonne
		if count >= 5 {
			penalty += 3 + (count - 5)
		}
	}

	return penalty
}

// evaluateRule2 calcule la pénalité pour les blocs 2x2 de même couleur
func evaluateRule2(matrix *image.RGBA, size int) int {
	penalty := 0

	for y := 0; y < size-1; y++ {
		for x := 0; x < size-1; x++ {
			color := isBlack(matrix.At(x, y))
			if isBlack(matrix.At(x+1, y)) == color &&
				isBlack(matrix.At(x, y+1)) == color &&
				isBlack(matrix.At(x+1, y+1)) == color {
				penalty += 3
			}
		}
	}

	return penalty
}

// evaluateRule3 calcule la pénalité pour les motifs finder-like (1:1:3:1:1)
func evaluateRule3(matrix *image.RGBA, size int) int {
	penalty := 0

	// Motif horizontal: noir-blanc-noir-noir-noir-blanc-noir
	pattern1 := []bool{true, false, true, true, true, false, true}
	// Motif vertical équivalent
	pattern2 := []bool{true, false, true, true, true, false, true}

	// Rechercher les motifs horizontaux
	for y := 0; y < size; y++ {
		for x := 0; x <= size-7; x++ {
			match := true
			for i := 0; i < 7; i++ {
				if isBlack(matrix.At(x+i, y)) != pattern1[i] {
					match = false
					break
				}
			}
			if match {
				penalty += 40
			}
		}
	}

	// Rechercher les motifs verticaux
	for x := 0; x < size; x++ {
		for y := 0; y <= size-7; y++ {
			match := true
			for i := 0; i < 7; i++ {
				if isBlack(matrix.At(x, y+i)) != pattern2[i] {
					match = false
					break
				}
			}
			if match {
				penalty += 40
			}
		}
	}

	return penalty
}

// evaluateRule4 calcule la pénalité pour le déséquilibre noir/blanc
func evaluateRule4(matrix *image.RGBA, size int) int {
	blackCount := 0
	totalCount := size * size

	// Compter les modules noirs
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if isBlack(matrix.At(x, y)) {
				blackCount++
			}
		}
	}

	// Calculer le pourcentage de modules noirs
	blackPercentage := (blackCount * 100) / totalCount

	// Calculer la différence avec 50% par tranches de 5%
	fivePercentDeviation := abs((blackPercentage - 50) / 5)

	return fivePercentDeviation * 10
}

// isBlack vérifie si une couleur est noire
func isBlack(c color.Color) bool {
	r, _, _, _ := c.RGBA()
	// Les valeurs RGBA sont normalisées sur 16 bits (0-65535)
	// On considère noir si la valeur est inférieure à la moitié (0x8000)
	return r < 0x8000
}

// AddFinderPatterns ajoute les motifs de positionnement à la matrice QR
func AddFinderPatterns(matrix *image.RGBA) {
	// Positions des motifs de positionnement (en haut à gauche, en haut à droite, en bas à gauche)
	positions := []struct{ x, y int }{
		{0, 0},                         // En haut à gauche
		{matrix.Bounds().Max.X - 7, 0}, // En haut à droite
		{0, matrix.Bounds().Max.Y - 7}, // En bas à gauche
	}

	for _, pos := range positions {
		// Dessiner le carré extérieur 7x7
		for i := 0; i < 7; i++ {
			for j := 0; j < 7; j++ {
				if i == 0 || i == 6 || j == 0 || j == 6 {
					matrix.Set(pos.x+i, pos.y+j, color.Black)
				}
			}
		}

		// Dessiner le carré intérieur 5x5
		for i := 1; i < 6; i++ {
			for j := 1; j < 6; j++ {
				matrix.Set(pos.x+i, pos.y+j, color.White)
			}
		}

		// Dessiner le carré central 3x3
		for i := 2; i < 5; i++ {
			for j := 2; j < 5; j++ {
				matrix.Set(pos.x+i, pos.y+j, color.Black)
			}
		}
	}
}

// AddSeparators ajoute les séparateurs à la matrice QR
func AddSeparators(matrix *image.RGBA) {
	size := matrix.Bounds().Max.X

	// Séparateurs horizontaux
	for x := 0; x < 8; x++ {
		matrix.Set(x, 7, color.White)        // En haut à gauche
		matrix.Set(size-8+x, 7, color.White) // En haut à droite
		matrix.Set(x, size-8, color.White)   // En bas à gauche
	}

	// Séparateurs verticaux
	for y := 0; y < 8; y++ {
		matrix.Set(7, y, color.White)        // En haut à gauche
		matrix.Set(size-8, y, color.White)   // En haut à droite
		matrix.Set(7, size-8+y, color.White) // En bas à gauche
	}
}

// PlaceData place les données dans la matrice QR selon le motif en zigzag standard
func PlaceData(matrix *image.RGBA, data string) {
	size := matrix.Bounds().Max.X

	// Convert data string to bit slice
	bits := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] == '1' {
			bits = append(bits, 1)
		} else {
			bits = append(bits, 0)
		}
	}

	bitIndex := 0
	goingUp := true

	// Process column pairs from right to left, skipping timing column 6
	col := size - 1
	for col > 0 {
		if col == 6 {
			col--
			continue
		}

		// For each row in the current direction
		if goingUp {
			for row := size - 1; row >= 0; row-- {
				for dx := 0; dx < 2; dx++ {
					c := col - dx
					if c < 0 {
						continue
					}
					if isValidDataPosition(c, row, size) {
						if bitIndex < len(bits) {
							if bits[bitIndex] == 1 {
								matrix.Set(c, row, color.Black)
							} else {
								matrix.Set(c, row, color.White)
							}
							bitIndex++
						}
					}
				}
			}
		} else {
			for row := 0; row < size; row++ {
				for dx := 0; dx < 2; dx++ {
					c := col - dx
					if c < 0 {
						continue
					}
					if isValidDataPosition(c, row, size) {
						if bitIndex < len(bits) {
							if bits[bitIndex] == 1 {
								matrix.Set(c, row, color.Black)
							} else {
								matrix.Set(c, row, color.White)
							}
							bitIndex++
						}
					}
				}
			}
		}

		goingUp = !goingUp
		col -= 2
	}
}

// isValidDataPosition vérifie si une position peut contenir des données
func isValidDataPosition(x, y, size int) bool {
	// Vérifier les motifs de positionnement
	if (x < 9 && y < 9) || // En haut à gauche
		(x > size-9 && y < 9) || // En haut à droite
		(x < 9 && y > size-9) { // En bas à gauche
		return false
	}

	// Vérifier la colonne de timing
	if x == 6 {
		return false
	}

	// Vérifier la ligne de timing
	if y == 6 {
		return false
	}

	// Vérifier les motifs d'alignement (exclure tout le carré 5×5)
	alignmentPositions := getAlignmentPositions(size)
	for _, pos := range alignmentPositions {
		ax, ay := pos[0], pos[1]
		if x >= ax-2 && x <= ax+2 && y >= ay-2 && y <= ay+2 {
			return false
		}
	}

	return true
}

// getAlignmentPositions calcule les positions des motifs d'alignement
func getAlignmentPositions(size int) [][2]int {
	version := (size - 17) / 4
	if version < 2 {
		return [][2]int{}
	}

	// Table officielle des positions d'alignement pour les versions 2 à 7
	alignmentTable := map[int][]int{
		2: {6, 18},
		3: {6, 22},
		4: {6, 26},
		5: {6, 30},
		6: {6, 34},
		7: {6, 22, 38},
	}

	var positions []int
	if pos, ok := alignmentTable[version]; ok {
		positions = pos
	} else {
		// Pour les versions supérieures, calcule dynamiquement
		numAlign := version/7 + 2
		step := 0
		if numAlign > 2 {
			step = (size - 13) / (numAlign - 1)
		}
		positions = make([]int, numAlign)
		positions[0] = 6
		for i := 1; i < numAlign-1; i++ {
			positions[i] = positions[i-1] + step
		}
		positions[numAlign-1] = size - 7
	}

	// Génère les couples (x, y) pour chaque motif d'alignement
	var result [][2]int
	for _, x := range positions {
		for _, y := range positions {
			// On évite les coins où il y a déjà un finder pattern
			if !((x == 6 && y == 6) ||
				(x == 6 && y == size-7) ||
				(x == size-7 && y == 6)) {
				result = append(result, [2]int{x, y})
			}
		}
	}
	return result
}

// AddTimingPatterns ajoute les motifs de timing à la matrice QR
func AddTimingPatterns(matrix *image.RGBA) {
	size := matrix.Bounds().Max.X
	for i := 8; i < size-8; i++ {
		col := color.RGBA{255, 255, 255, 255} // Blanc par défaut
		if i%2 == 0 {
			col = color.RGBA{0, 0, 0, 255} // Noir sur les cases paires
		}
		matrix.Set(i, 6, col)
		matrix.Set(6, i, col)
	}
}

// AddAlignmentPatterns ajoute les motifs d'alignement à la matrice QR
func AddAlignmentPatterns(matrix *image.RGBA, version int) {
	positions := GetAlignmentPatternPositions(version)
	for _, pos := range positions {
		AddAlignmentPattern(matrix, pos.x, pos.y)
	}
}

// Ajoute un motif d'alignement à une position donnée
func AddAlignmentPattern(matrix *image.RGBA, x, y int) {
	// Dessiner le carré extérieur 5x5
	for i := -2; i <= 2; i++ {
		for j := -2; j <= 2; j++ {
			if i == -2 || i == 2 || j == -2 || j == 2 {
				matrix.Set(x+i, y+j, color.Black)
			} else {
				matrix.Set(x+i, y+j, color.White)
			}
		}
	}

	// Dessiner le carré intérieur 3x3
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == -1 || i == 1 || j == -1 || j == 1 {
				matrix.Set(x+i, y+j, color.White)
			} else {
				matrix.Set(x+i, y+j, color.Black)
			}
		}
	}
}

// Structure pour les tâches de traitement
type DataProcessingTask struct {
	data   string
	result chan string
}

// Pool de workers pour le traitement des données
func ProcessDataWithWorkers(data string, numWorkers int) string {
	tasks := make(chan DataProcessingTask)

	// Démarrage des workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			for task := range tasks {
				var result string
				switch DetectDataType(task.data) {
				case "numeric":
					result, _ = EncodeNumeric(task.data)
				case "alphanumeric":
					result, _ = EncodeAlphanumeric(task.data)
				case "byte":
					result, _ = EncodeByte(task.data)
				case "kanji":
					result, _ = EncodeKanji(task.data)
				}
				task.result <- result
			}
		}()
	}

	// Diviser les données en chunks et les envoyer aux workers
	chunkSize := len(data) / numWorkers
	if chunkSize == 0 {
		chunkSize = 1
	}

	var finalResult strings.Builder
	var wg sync.WaitGroup

	for i := 0; i < len(data); i += chunkSize {
		wg.Add(1)
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		resultChan := make(chan string)
		tasks <- DataProcessingTask{
			data:   data[i:end],
			result: resultChan,
		}

		go func() {
			defer wg.Done()
			result := <-resultChan
			finalResult.WriteString(result)
		}()
	}

	wg.Wait()
	close(tasks)

	return finalResult.String()
}

// SaveQRImage sauvegarde la matrice QR en image PNG
func SaveQRImage(matrix *image.RGBA, outputFile string, scale int) error {
	bounds := matrix.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Création d'une nouvelle image avec la taille mise à l'échelle
	scaledImage := image.NewRGBA(image.Rect(0, 0, width*scale, height*scale))

	// Mise à l'échelle de l'image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			color := matrix.At(x, y)
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					scaledImage.Set(x*scale+sx, y*scale+sy, color)
				}
			}
		}
	}

	// Création du fichier de sortie
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier : %v", err)
	}
	defer file.Close()

	// Encodage en PNG
	if err := png.Encode(file, scaledImage); err != nil {
		return fmt.Errorf("erreur lors de l'encodage PNG : %v", err)
	}

	return nil
}

// SaveQRImageWithQuietZone sauvegarde la matrice QR en image PNG avec une zone calme (quiet zone)
func SaveQRImageWithQuietZone(matrix *image.RGBA, outputFile string, scale int, quietZone int) error {
	bounds := matrix.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Calculer les dimensions avec la zone calme
	totalWidth := width + (quietZone * 2)
	totalHeight := height + (quietZone * 2)

	// Création d'une nouvelle image avec la taille mise à l'échelle incluant la zone calme
	scaledImage := image.NewRGBA(image.Rect(0, 0, totalWidth*scale, totalHeight*scale))

	// Remplir l'image entière en blanc pour la zone calme
	for y := 0; y < totalHeight*scale; y++ {
		for x := 0; x < totalWidth*scale; x++ {
			scaledImage.Set(x, y, color.White)
		}
	}

	// Mise à l'échelle de l'image QR au centre
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := matrix.At(x, y)
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					scaledImage.Set((x+quietZone)*scale+sx, (y+quietZone)*scale+sy, c)
				}
			}
		}
	}

	// Création du fichier de sortie
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier : %v", err)
	}
	defer file.Close()

	// Encodage en PNG
	if err := png.Encode(file, scaledImage); err != nil {
		return fmt.Errorf("erreur lors de l'encodage PNG : %v", err)
	}

	return nil
}

// setFormatBit place un bit de format info (noir=true, blanc=false)
func setFormatBit(matrix *image.RGBA, x, y int, black bool) {
	if black {
		matrix.Set(x, y, color.Black)
	} else {
		matrix.Set(x, y, color.White)
	}
}

// AddFormatInfo ajoute l'information de format au QR code selon ISO/IEC 18004
// Les positions sont vérifiées d'après la table 25 du standard.
func AddFormatInfo(matrix *image.RGBA, ecLevel string, maskPattern int) {
	formatInfoBits := FormatInfo[ecLevel][maskPattern]
	size := matrix.Bounds().Max.X

	for i := 0; i < 15; i++ {
		bit := formatInfoBits[i] == '1'

		// --- Première copie (autour du finder haut-gauche) ---
		// Bits 0-7 : horizontaux sur la ligne y=8 (cols 0→8, on saute col 6=timing)
		// Bits 8-14 : verticaux sur la col x=8 (rows 7→0, on saute row 6=timing)
		switch {
		case i < 6:
			setFormatBit(matrix, i, 8, bit)
		case i == 6:
			setFormatBit(matrix, 7, 8, bit) // saute x=6 (timing)
		case i == 7:
			setFormatBit(matrix, 8, 8, bit)
		case i == 8:
			setFormatBit(matrix, 8, 7, bit)
		default:
			// i=9→y=5, i=10→y=4, ..., i=14→y=0  (on saute y=6 entre i=8 et i=9)
			setFormatBit(matrix, 8, 14-i, bit)
		}

		// --- Deuxième copie (finder haut-droit + finder bas-gauche) ---
		// Bits 0-7 : sur la ligne y=8, cols size-1 → size-8 (de l'extérieur vers l'intérieur)
		// Bits 8-14 : sur la col x=8, rows size-7 → size-1
		if i < 8 {
			setFormatBit(matrix, size-1-i, 8, bit)
		} else {
			setFormatBit(matrix, 8, size-7+(i-8), bit)
		}
	}

	// Module sombre (toujours noir) : position (8, 4*version+9)
	version := (size - 17) / 4
	matrix.Set(8, 4*version+9, color.Black)
}

// calculateAvailableCapacity retourne la capacité en bits pour les données
// selon la taille de la matrice et le niveau de correction d'erreur.
func calculateAvailableCapacity(size int, ecLevel string) int {
	version := (size - 17) / 4
	if version >= 1 && version <= 40 {
		return DataCodewords(version, ecLevel) * 8
	}

	// Fallback : compter les positions valides
	capacity := 0
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			if isValidDataPosition(x, y, size) {
				capacity++
			}
		}
	}
	return capacity
}

// Fonction existante utilisée pour la correction d'erreur
func AddErrorCorrection(data string, level string, version int) string {
	// Utilise la fonction existante dans error_correction.go
	return AddErrorCorrectionEC(data, level, version)
}
