package main

import (
	"fmt"
	"log"

	"github.com/guyghost/constantine/internal/exchanges/dydx"
)

func main() {
	fmt.Println("=== dYdX Mnemonic Authentication Test ===")

	// Example 1: Generate a new mnemonic
	fmt.Println("1. Generating new mnemonic...")
	mnemonic, err := dydx.GenerateMnemonic()
	if err != nil {
		log.Fatal("Failed to generate mnemonic:", err)
	}
	fmt.Printf("   Generated mnemonic: %s\n", mnemonic[:50]+"...")
	fmt.Println("   ⚠️  Save this mnemonic securely!")

	// Example 2: Create wallet from mnemonic
	fmt.Println("2. Creating wallet from mnemonic...")
	wallet, err := dydx.NewWalletFromMnemonic(mnemonic, 0)
	if err != nil {
		log.Fatal("Failed to create wallet:", err)
	}
	fmt.Printf("   Address: %s\n", wallet.Address)
	fmt.Printf("   SubAccount: %d\n", wallet.SubAccountNumber)
	fmt.Printf("   Full SubAccount Address: %s\n", wallet.SubAccountAddress())

	// Example 3: Validate a mnemonic
	fmt.Println("3. Validating mnemonic...")
	if err := dydx.ValidateMnemonic(mnemonic); err != nil {
		log.Fatal("Mnemonic validation failed:", err)
	}
	fmt.Println("   ✓ Mnemonic is valid")

	// Example 4: Create dYdX client with mnemonic
	fmt.Println("4. Creating dYdX client...")
	client, err := dydx.NewClientWithMnemonic(mnemonic, 0)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	fmt.Printf("   ✓ Client created successfully\n")
	fmt.Printf("   Wallet Address: %s\n", client.GetWalletAddress())
	fmt.Printf("   SubAccount Address: %s\n", client.GetSubAccountAddress())
	fmt.Printf("   Is Authenticated: %v\n\n", client.IsAuthenticated())

	fmt.Println("=== Test Complete ===")
	fmt.Println("\nTo use this in your bot:")
	fmt.Println("1. Save your mnemonic to .env file:")
	fmt.Println("   EXCHANGE_API_SECRET=\"your mnemonic phrase here\"")
	fmt.Println("2. Set EXCHANGE=dydx")
	fmt.Println("3. Run: ./bin/constantine --headless")
}
