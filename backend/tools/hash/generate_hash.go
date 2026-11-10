package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func GenerateHash() {
	password := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		return
	}
	fmt.Printf("Bcrypt hash for '%s':\n%s\n", password, string(hash))
}

func main() {
	GenerateHash()
}
