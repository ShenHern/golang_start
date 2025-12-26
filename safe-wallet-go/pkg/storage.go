package pkg

import (
	"encoding/json"
	"errors"
	"os"
)

// SaveWallet encrypts and saves the wallet to a file
func SaveWallet(wallet *Wallet, filepath string, password string) error {
	// Marshal wallet to JSON
	jsonData, err := json.MarshalIndent(wallet, "", "  ")
	if err != nil {
		return err
	}

	// Encrypt the JSON data
	encrypted, err := EncryptData(jsonData, password)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filepath, encrypted, 0600) // 0600 = rw-------
}

// LoadWallet loads and decrypts a wallet from a file
func LoadWallet(filepath string, password string) (*Wallet, error) {
	// Read encrypted file
	encrypted, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("wallet file does not exist")
		}
		return nil, err
	}

	// Decrypt the data
	jsonData, err := DecryptData(encrypted, password)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON to wallet
	var wallet Wallet
	if err := json.Unmarshal(jsonData, &wallet); err != nil {
		return nil, err
	}

	return &wallet, nil
}

// CreateNewWallet creates a new empty wallet
func CreateNewWallet() *Wallet {
	return &Wallet{
		Version: 1,
		Groups:  []Group{},
	}
}

// WalletExists checks if a wallet file exists
func WalletExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}
