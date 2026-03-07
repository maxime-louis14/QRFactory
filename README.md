# QRFactory

```
 ██████╗ ██████╗ ███████╗ █████╗  ██████╗████████╗ ██████╗ ██████╗ ██╗   ██╗
██╔═══██╗██╔══██╗██╔════╝██╔══██╗██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗╚██╗ ██╔╝
██║   ██║██████╔╝█████╗  ███████║██║        ██║   ██║   ██║██████╔╝ ╚████╔╝
██║▄▄ ██║██╔══██╗██╔══╝  ██╔══██║██║        ██║   ██║   ██║██╔══██╗  ╚██╔╝
╚██████╔╝██║  ██║██║     ██║  ██║╚██████╗   ██║   ╚██████╔╝██║  ██║   ██║
 ╚══▀▀═╝ ╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝ ╚═════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝   ╚═╝

  ██████╗ ██████╗      ██████╗ ███████╗███╗   ██╗███████╗██████╗  █████╗ ████████╗ ██████╗ ██████╗
 ██╔═══██╗██╔══██╗    ██╔════╝ ██╔════╝████╗  ██║██╔════╝██╔══██╗██╔══██╗╚══██╔══╝██╔═══██╗██╔══██╗
 ██║   ██║██████╔╝    ██║  ███╗█████╗  ██╔██╗ ██║█████╗  ██████╔╝███████║   ██║   ██║   ██║██████╔╝
 ██║▄▄ ██║██╔══██╗    ██║   ██║██╔══╝  ██║╚██╗██║██╔══╝  ██╔══██╗██╔══██║   ██║   ██║   ██║██╔══██╗
 ╚██████╔╝██║  ██║    ╚██████╔╝███████╗██║ ╚████║███████╗██║  ██║██║  ██║   ██║   ╚██████╔╝██║  ██║
  ╚══▀▀═╝ ╚═╝  ╚═╝     ╚═════╝ ╚══════╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝
```

QRFactory est un générateur de codes QR écrit from-scratch en Go, implémentant la norme ISO/IEC 18004. Le projet adopte une architecture modulaire et une approche TDD.

## Table des matières

- [Installation](#installation)
- [Utilisation](#utilisation)
- [Interface en ligne de commande (CLI)](#interface-en-ligne-de-commande-cli)
- [Architecture](#architecture)
- [Tests](#tests)
- [Contribuer](#contribuer)
- [État du développement](#état-du-développement)
- [Licence](#licence)

## Installation

1. **Cloner le dépôt :**

   ```sh
   git clone https://github.com/le-veilleur/QRFactory.git
   cd QRFactory
   ```

2. **Initialiser les modules Go :**

   ```sh
   go mod tidy
   ```

3. **Compiler le binaire :**

   ```sh
   go build -o qrfactory cmd/qrfactory/main.go
   ```

4. **(Optionnel) Installer globalement :**

   ```sh
   mv qrfactory /usr/local/bin/
   ```

   Le binaire sera alors accessible depuis n'importe quel répertoire :

   ```sh
   qrfactory -d "https://example.com"
   ```

## Utilisation

```sh
go run cmd/qrfactory/main.go -d "https://example.com"
```

### Interface en ligne de commande (CLI)

```sh
go run cmd/qrfactory/main.go -d "DONNÉES" [options]
```

Options disponibles :

| Flag | Raccourci | Description | Défaut |
|------|-----------|-------------|--------|
| `--data` | `-d` | Données à encoder (obligatoire) | — |
| `--scale` | `-s` | Facteur d'échelle de l'image | `30` |
| `--version` | `-v` | Version du QR code (1–40) | auto |
| `--error-correction` | `-e` | Niveau de correction d'erreur (L, M, Q, H) | `H` |
| `--output` | `-o` | Fichier de sortie | `qrcode.png` |
| `--quiet-zone` | `-q` | Largeur de la zone calme (en modules) | `4` |
| `--bg-color` | | Couleur de fond | `white` |
| `--fg-color` | | Couleur des modules | `black` |

Exemples :

```sh
# QR code simple
go run cmd/qrfactory/main.go -d "https://example.com"

# Avec échelle et niveau de correction personnalisés
go run cmd/qrfactory/main.go -d "https://example.com" -s 20 -e M -o output.png

# Avec zone calme élargie
go run cmd/qrfactory/main.go -d "HELLO WORLD" -q 8

# Afficher la version du programme
go run cmd/qrfactory/main.go version
```

## Architecture

Pipeline de génération (orchestré dans `pkg/qr/generator.go`) :

1. Détection du type de données (`detect.go`) — numérique, alphanumérique, byte, kanji
2. Calcul de la version minimale requise
3. Encodage des bits de données (`encoding.go`) — indicateur de mode + longueur + données
4. Construction de la matrice `image.RGBA`, placement des motifs structurels (finder, séparateur, alignement, timing)
5. Correction d'erreur (`error_correction.go`) via le corps de Galois (`InitGaloisField`)
6. Placement des données en zigzag sur les modules valides
7. Évaluation des 8 masques, sélection du score de pénalité le plus faible (`mask.go`)
8. Ajout des bits d'information de format autour des finder patterns
9. Sauvegarde PNG avec zone calme (`SaveQRImageWithQuietZone`)

Structure du projet :

```
QRFactory/
├── cmd/
│   ├── main.go                  # Ancien point d'entrée CLI (sans quiet zone)
│   └── qrfactory/
│       └── main.go              # Point d'entrée principal CLI (Cobra)
│
├── internal/
│   └── model/
│       ├── qr_code.go
│       └── qr_code_test.go
│
├── pkg/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   └── qr/
│       ├── detect.go            # Détection du type de données
│       ├── encoding.go          # Encodage numérique, alphanumérique, byte, kanji
│       ├── error_correction.go  # Correction d'erreur Reed-Solomon
│       ├── generator.go         # Orchestration du pipeline de génération
│       ├── mask.go              # Application et évaluation des masques
│       ├── utils.go             # Utilitaires
│       ├── version.go           # Tables de capacité et de correction
│       ├── e2e_test.go
│       ├── generator_test.go
│       ├── unit_test.go
│       └── tests/
│           └── encoding_test.go
│
├── go.mod
└── go.sum
```

## Tests

```sh
# Tous les tests
go test ./...

# Package spécifique
go test ./pkg/qr/...

# Test unique
go test ./pkg/qr/ -run TestEncodeNumeric
```

Les tests couvrent : l'encodage (numérique, alphanumérique, byte, kanji), Reed-Solomon, le placement des données, les informations de format, et des tests E2E avec décodage réel via `gozxing`.

## Contribuer

1. Forkez le dépôt.
2. Créez une branche (`git checkout -b feature/ma-fonctionnalite`).
3. Commitez vos modifications (`git commit -am 'Ajoute une fonctionnalité'`).
4. Poussez la branche (`git push origin feature/ma-fonctionnalite`).
5. Ouvrez une Pull Request.

## État du développement

- [x] Détection du type de données (numérique, alphanumérique, byte, kanji)
- [x] Encodage numérique
- [x] Encodage alphanumérique
- [x] Encodage byte
- [x] Encodage kanji
- [x] Correction d'erreur Reed-Solomon (corps de Galois GF(256))
- [x] Placement des motifs structurels (finder, séparateur, timing, alignement)
- [x] Placement des données en zigzag
- [x] Application et sélection du masque optimal (8 masques, 4 règles de pénalité)
- [x] Informations de format (EC level + masque)
- [x] Génération d'image PNG (avec zone calme configurable)
- [x] Interface CLI complète (Cobra)
- [x] Tests unitaires et E2E

## Licence

Ce projet est sous licence MIT. Voir le fichier [LICENSE](LICENSE) pour plus de détails.