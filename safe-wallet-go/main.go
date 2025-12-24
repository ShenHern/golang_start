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

	// Example: Add a root group (ID will be auto-generated)
	rootGroup := &Group{
		Name:    "Root",
		Groups:  []Group{},
		Entries: []Entry{},
	}

	rootPath := Path{GroupIDs: []string{}}
	if err := service.AddGroup(rootPath, rootGroup); err != nil {
		log.Fatal("Failed to add root group:", err)
	}
	fmt.Printf("Created root group with ID: %s\n", rootGroup.ID)

	// Example: Add a work group under root (ID will be auto-generated)
	workGroup := &Group{
		Name:    "Work",
		Groups:  []Group{},
		Entries: []Entry{},
	}

	rootGroupPath := Path{GroupIDs: []string{rootGroup.ID}}
	if err := service.AddGroup(rootGroupPath, workGroup); err != nil {
		log.Fatal("Failed to add work group:", err)
	}
	fmt.Printf("Created work group with ID: %s\n", workGroup.ID)

	// Example: Add an entry to the work group (ID will be auto-generated)
	entry := &Entry{
		Title:    "AWS Console",
		Username: "admin",
		Password: "encrypted-in-memory",
		URL:      "https://aws.amazon.com",
		Notes:    "MFA enabled",
	}

	workGroupPath := Path{GroupIDs: []string{rootGroup.ID, workGroup.ID}}
	if err := service.AddEntry(workGroupPath, entry); err != nil {
		log.Fatal("Failed to add entry:", err)
	}
	fmt.Printf("Created entry with ID: %s\n", entry.ID)

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

	// Example: Find entry by ID (using the auto-generated ID)
	fmt.Println("\nFinding entry by ID:")
	var path Path
	var entryRes *Entry
	var err error
	path, entryRes, err = service.FindEntryByID(entry.ID)
	if err != nil {
		log.Fatal("Failed to find entry:", err)
	}
	fmt.Printf("  Found entry: %s at path: %v\n", entryRes.Title, path.GroupIDs)

	// Example: Update an entry
	fmt.Println("\nUpdating entry:")
	updatedEntry := Entry{
		Title:    "AWS Console (Updated)",
		Username: "admin",
		Password: "new-encrypted-password",
		URL:      "https://aws.amazon.com",
		Notes:    "MFA enabled, 2FA required",
	}

	// Create path with entry ID for update
	updatePath := Path{
		GroupIDs: path.GroupIDs,
		EntryID:  entry.ID,
	}

	if err := service.UpdateEntry(updatePath, updatedEntry); err != nil {
		log.Fatal("Failed to update entry:", err)
	}
	fmt.Printf("  Updated entry: %s\n", updatedEntry.Title)

	// Verify the update by finding the entry again
	_, updatedEntryRes, err := service.FindEntryByID(entry.ID)
	if err != nil {
		log.Fatal("Failed to find updated entry:", err)
	}
	fmt.Printf("  Verified - Entry title: %s, Notes: %s\n", updatedEntryRes.Title, updatedEntryRes.Notes)

	// Save the wallet after update
	if err := service.Save(); err != nil {
		log.Fatal("Failed to save wallet after update:", err)
	}
	fmt.Println("Wallet saved successfully after update!")
}
