package qr_test

import (
	"testing"

	"qrfactory/pkg/qr"
)

func TestEncodeNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Encode single digit",
			input:    "4",
			expected: "0100",
			wantErr:  false,
		},
		{
			name:     "Encode two digits",
			input:    "42",
			expected: "0101010",
			wantErr:  false,
		},
		{
			name:     "Encode three digits",
			input:    "123",
			expected: "0001111011",
			wantErr:  false,
		},
		{
			name:     "Invalid numeric input",
			input:    "12a",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := qr.EncodeNumeric(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("EncodeNumeric() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEncodeAlphanumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Encode single character",
			input:    "A",
			expected: "001010",
			wantErr:  false,
		},
		{
			name:     "Encode two characters",
			input:    "AB",
			expected: "00111001101", // A=10, B=11 → 10*45+11=461 → 11 bits
			wantErr:  false,
		},
		{
			name:     "Encode with special characters",
			input:    "A$",
			expected: "00111100111", // A=10, $=37 → 10*45+37=487 → 11 bits
			wantErr:  false,
		},
		{
			name:     "Invalid alphanumeric input",
			input:    "@",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := qr.EncodeAlphanumeric(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeAlphanumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("EncodeAlphanumeric() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEncodeByte(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Encode ASCII character",
			input:    "A",
			expected: "01000001",
			wantErr:  false,
		},
		{
			name:     "Encode UTF-8 character",
			input:    "é",
			expected: "11000011" + "10101001",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := qr.EncodeByte(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("EncodeByte() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEncodeKanji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Encode Hiragana",
			input:    "あ",
			expected: "0000100100000", // SJIS 0x82A0 → (0x01*0xC0)+0x60=288 → 13 bits
			wantErr:  false,
		},
		{
			name:     "Invalid Kanji input",
			input:    "A",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := qr.EncodeKanji(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeKanji() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("EncodeKanji() = %v, want %v", got, tt.expected)
			}
		})
	}
}
