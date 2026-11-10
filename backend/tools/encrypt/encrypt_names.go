package encrypt

import (
	"fmt"
	"os"
	"strings"

	"github.com/kwizera/university-invoices/internal/encryption"
)

func GenerateEncrypt() {
	// Get encryption key from environment
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		key = "12345678901234567890123456789012"
		fmt.Println("Warning: Using default encryption key. Set ENCRYPTION_KEY environment variable for production.")
	}

	// Data to encrypt (names)
	data := map[string]string{
		"KWIZERISEZERANO": "KWIZERISEZERANO",
		"Germaine":        "Germaine",
		"Jean Remmy":      "Jean Remmy",
		"RUKUNDO":         "RUKUNDO",
		"RURANGWA":        "RURANGWA",
		"NIZEYIMANA":      "NIZEYIMANA",
		"UWIMBABAZI":      "UWIMBABAZI",
		"RUCAMUKIBATSI":   "RUCAMUKIBATSI",
		"Hope":            "Hope",
		"Jean Paul":       "Jean Paul",
		"Diallo Fiaro":    "Diallo Fiaro",
	}

	fmt.Println("Encrypted values:")
	fmt.Println(strings.Repeat("=", 80))

	for label, value := range data {
		encrypted, err := encryption.Encrypt(value, key)
		if err != nil {
			fmt.Printf("Error encrypting %s: %v\n", label, err)
			continue
		}
		fmt.Printf("%s: %s\n", label, encrypted)
	}

	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	GenerateEncrypt()
}
