package utils

import (
	"math/rand"
	"strings"
	"time"
)

var (
	glitchChars = []rune("█▓▒░#*?!@%&")
	glitchWords = []string{"[DATA_EXPUNGED]", "[ERR_CORRUPT]", "[SECTOR_LOST]", "[SIGNAL_DEGRADED]"}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// CorruptText takes an input string and randomly corrupts it based on a severity percentage (0-100).
func CorruptText(input string, severity int) string {
	if severity <= 0 {
		return input
	}
	if severity > 100 {
		severity = 100
	}

	// Calculate a realistic probability based on severity (e.g., severity 10 means 10% chance per character)
	probCharCorrupt := float64(severity) / 100.0
	probWordCorrupt := float64(severity) / 500.0 // Word corruption is rarer

	words := strings.Fields(input)
	for i, word := range words {
		// Chance to corrupt the entire word
		if rand.Float64() < probWordCorrupt {
			words[i] = glitchWords[rand.Intn(len(glitchWords))]
			continue
		}

		// Otherwise, chance to corrupt individual characters within the word
		runes := []rune(word)
		for j, char := range runes {
			// Don't corrupt spaces or common sentence punctuation as often to maintain 'readability' of the corruption
			if char == ' ' || char == '.' || char == ',' {
				continue
			}
			if rand.Float64() < probCharCorrupt {
				runes[j] = glitchChars[rand.Intn(len(glitchChars))]
			}
		}
		words[i] = string(runes)
	}

	return strings.Join(words, " ")
}
