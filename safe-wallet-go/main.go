package main

import (
	"fmt"
	"log"
)

// Example usage of the wallet service
func main() {
	// Example: Create a new wallet service
	filepath := "wallet.dat"
	password := "my-secure-password"

	service := NewWalletService(filepath, password)

	// Check if wallet exists, if not create a new one
	if !WalletExists(filepath) {
		fmt.Println("Creating new wallet...")
		if err := service.CreateNew(); err != nil {
			log.Fatal("Failed to create wallet:", err)
		}
	} else {
		fmt.Println("Loading existing wallet...")
		if err := service.Load(); err != nil {
			log.Fatal("Failed to load wallet:", err)
		}
	}

	// Example: Add a root group
	rootGroup := Group{
		ID:      "grp-root",
		Name:    "Root",
		Groups:  []Group{},
		Entries: []Entry{},
	}

	rootPath := Path{GroupIDs: []string{}}
	if err := service.AddGroup(rootPath, rootGroup); err != nil {
		log.Fatal("Failed to add root group:", err)
	}

	// Example: Add a work group under root
	workGroup := Group{
		ID:      "grp-work",
		Name:    "Work",
		Groups:  []Group{},
		Entries: []Entry{},
	}

	rootGroupPath := Path{GroupIDs: []string{"grp-root"}}
	if err := service.AddGroup(rootGroupPath, workGroup); err != nil {
		log.Fatal("Failed to add work group:", err)
	}

	// Example: Add an entry to the work group
	entry := Entry{
		ID:       "ent-1",
		Title:    "AWS Console",
		Username: "admin",
		Password: "encrypted-in-memory",
		URL:      "https://aws.amazon.com",
		Notes:    "MFA enabled",
	}

	workGroupPath := Path{GroupIDs: []string{"grp-root", "grp-work"}}
	if err := service.AddEntry(workGroupPath, entry); err != nil {
		log.Fatal("Failed to add entry:", err)
	}

	// Save the wallet
	if err := service.Save(); err != nil {
		log.Fatal("Failed to save wallet:", err)
	}

	fmt.Println("Wallet operations completed successfully!")

	// Example: Traverse forward
	fmt.Println("\nForward traversal:")
	service.TraverseForward(func(info PathInfo) bool {
		if info.IsEntry {
			fmt.Printf("  [Entry] Depth: %d, ID: %s, Title: %s\n", info.Depth, info.Entry.ID, info.Entry.Title)
		} else {
			fmt.Printf("  [Group] Depth: %d, ID: %s, Name: %s\n", info.Depth, info.Group.ID, info.Group.Name)
		}
		return true
	})

	// Example: Traverse backward
	fmt.Println("\nBackward traversal:")
	service.TraverseBackward(func(info PathInfo) bool {
		if info.IsEntry {
			fmt.Printf("  [Entry] Depth: %d, ID: %s, Title: %s\n", info.Depth, info.Entry.ID, info.Entry.Title)
		} else {
			fmt.Printf("  [Group] Depth: %d, ID: %s, Name: %s\n", info.Depth, info.Group.ID, info.Group.Name)
		}
		return true
	})

	// Example: Find entry by ID
	fmt.Println("\nFinding entry by ID:")
	var path Path
	var entryRes *Entry
	var err error
	path, entryRes, err = service.FindEntryByID("ent-1")
	if err != nil {
		log.Fatal("Failed to find entry:", err)
	}
	fmt.Printf("  Found entry: %s at path: %v\n", entryRes.Title, path.GroupIDs)
}
