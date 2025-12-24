package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	filepath := "wallet.dat"

	// Step 1: Handle password
	var password string
	if !WalletExists(filepath) {
		fmt.Println("=== Safe Wallet - New Wallet ===")
		fmt.Print("Create a password for your new wallet: ")
		password = readPassword()
		if password == "" {
			log.Fatal("Password cannot be empty")
		}
		fmt.Print("Confirm password: ")
		confirmPassword := readPassword()
		if password != confirmPassword {
			log.Fatal("Passwords do not match")
		}
	} else {
		fmt.Println("=== Safe Wallet ===")
		fmt.Print("Enter your wallet password: ")
		password = readPassword()
	}

	// Initialize service
	service := NewWalletService(filepath, password)

	// Load or create wallet
	if !WalletExists(filepath) {
		fmt.Println("Creating new wallet...")
		if err := service.CreateNew(); err != nil {
			log.Fatal("Failed to create wallet:", err)
		}
		fmt.Println("Wallet created successfully!")
	} else {
		if err := service.Load(); err != nil {
			log.Fatal("Failed to load wallet: ", err)
		}
		fmt.Println("Wallet loaded successfully!")
	}

	// Start at root (empty path)
	currentPath := Path{GroupIDs: []string{}}

	// Show menu on startup
	fmt.Println("\nWelcome to Safe Wallet!")
	displayMenu()

	// Main CLI loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		displayCurrentLocation(service, currentPath)

		fmt.Print("\nEnter command (type 'help' for menu): ")
		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}

		switch strings.ToLower(command) {
		case "help", "h", "?":
			displayMenu()
		case "1", "cg", "create-group":
			handleCreateGroup(service, currentPath, scanner)
		case "2", "ce", "create-entry":
			handleCreateEntry(service, currentPath, scanner)
		case "3", "l", "list":
			handleList(service, currentPath)
		case "4", "ug", "update-group":
			handleUpdateGroup(service, currentPath, scanner)
		case "6", "ue", "update-entry":
			handleUpdateEntry(service, currentPath, scanner)
		case "7", "dg", "delete-group":
			handleDeleteGroup(service, currentPath, scanner)
		case "8", "de", "delete-entry":
			handleDeleteEntry(service, currentPath, scanner)
		case "9", "f", "forward":
			currentPath = handleTraverseForward(service, currentPath, scanner)
		case "10", "b", "back":
			currentPath = handleTraverseBackward(service, currentPath)
		case "11", "s", "search":
			handleSearchEntry(service, scanner)
		case "12", "t", "tree":
			handleDisplayTree(service, currentPath)
		case "13", "n", "navigate":
			currentPath = handleNavigateIntoGroup(service, currentPath, scanner)
		case "14", "r", "root":
			currentPath = Path{GroupIDs: []string{}}
			fmt.Println("Returned to root")
		case "15", "save":
			if err := service.Save(); err != nil {
				fmt.Printf("Error saving wallet: %v\n", err)
			} else {
				fmt.Println("Wallet saved successfully!")
			}
		case "16", "q", "quit", "exit":
			// Auto-save before exit
			if err := service.Save(); err != nil {
				fmt.Printf("Error saving wallet: %v\n", err)
			}
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Unknown command. Type 'help' to see available commands.")
		}
	}
}

func readPassword() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

func displayCurrentLocation(service *WalletService, path Path) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	if len(path.GroupIDs) == 0 {
		fmt.Println("Current Location: ROOT")
	} else {
		group, err := FindGroupByPath(service.GetWallet(), path)
		if err == nil {
			fmt.Printf("Current Location: %s (ID: %s)\n", group.Name, group.ID)
		} else {
			fmt.Println("Current Location: ROOT")
		}
	}
	fmt.Println(strings.Repeat("=", 50))
}

func displayMenu() {
	fmt.Println("\nCommands:")
	fmt.Println("  1 (cg)  - Create Group")
	fmt.Println("  2 (ce)  - Create Entry")
	fmt.Println("  3 (l)   - List (groups and entries at current level)")
	fmt.Println("  4 (ug)  - Update Group")
	fmt.Println("  5 (ue)  - Update Entry")
	fmt.Println("  6 (dg)  - Delete Group")
	fmt.Println("  7 (de)  - Delete Entry")
	fmt.Println("  8 (f)   - Traverse Forward (groups one level down)")
	fmt.Println("  9 (b)   - Traverse Backward (go up one level)")
	fmt.Println("  10 (s)  - Search Entry by Title")
	fmt.Println("  11 (t)  - Display Tree (show full hierarchy)")
	fmt.Println("  12 (n)  - Navigate into Group")
	fmt.Println("  13 (r)  - Return to Root")
	fmt.Println("  14      - Save Wallet")
	fmt.Println("  15 (q)  - Quit")
}

func handleCreateGroup(service *WalletService, path Path, scanner *bufio.Scanner) {
	fmt.Print("Enter group name: ")
	if !scanner.Scan() {
		return
	}
	name := strings.TrimSpace(scanner.Text())
	if name == "" {
		fmt.Println("Group name cannot be empty")
		return
	}

	group := &Group{
		Name:    name,
		Groups:  []Group{},
		Entries: []Entry{},
	}

	if err := service.AddGroup(path, group); err != nil {
		fmt.Printf("Error creating group: %v\n", err)
		return
	}

	fmt.Printf("Group '%s' created successfully with ID: %s\n", name, group.ID)
	if err := service.Save(); err != nil {
		fmt.Printf("Error saving: %v\n", err)
	}
}

func handleCreateEntry(service *WalletService, path Path, scanner *bufio.Scanner) {
	if len(path.GroupIDs) == 0 {
		fmt.Println("Cannot create entry at root. Please navigate to a group first.")
		return
	}

	fmt.Print("Enter entry title: ")
	if !scanner.Scan() {
		return
	}
	title := strings.TrimSpace(scanner.Text())
	if title == "" {
		fmt.Println("Entry title cannot be empty")
		return
	}

	fmt.Print("Enter username: ")
	if !scanner.Scan() {
		return
	}
	username := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter password: ")
	if !scanner.Scan() {
		return
	}
	password := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter URL (optional): ")
	if !scanner.Scan() {
		return
	}
	url := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter notes (optional): ")
	if !scanner.Scan() {
		return
	}
	notes := strings.TrimSpace(scanner.Text())

	entry := &Entry{
		Title:    title,
		Username: username,
		Password: password,
		URL:      url,
		Notes:    notes,
	}

	if err := service.AddEntry(path, entry); err != nil {
		fmt.Printf("Error creating entry: %v\n", err)
		return
	}

	fmt.Printf("Entry '%s' created successfully with ID: %s\n", title, entry.ID)
	if err := service.Save(); err != nil {
		fmt.Printf("Error saving: %v\n", err)
	}
}

func handleList(service *WalletService, path Path) {
	var groups []Group
	var entries []Entry

	if len(path.GroupIDs) == 0 {
		// At root - list root groups
		groups = service.GetWallet().Groups
		// No entries at root level
	} else {
		group, err := FindGroupByPath(service.GetWallet(), path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		groups = group.Groups
		entries = group.Entries
	}

	// Display groups
	if len(groups) > 0 {
		fmt.Println("\nGroups:")
		for i, group := range groups {
			fmt.Printf("  %d. %s (ID: %s) - %d subgroups, %d entries\n",
				i+1, group.Name, group.ID, len(group.Groups), len(group.Entries))
		}
	} else {
		fmt.Println("\nNo groups found in current location.")
	}

	// Display entries
	if len(entries) > 0 {
		fmt.Println("\nEntries:")
		for i, entry := range entries {
			fmt.Printf("  %d. %s (ID: %s)\n", i+1, entry.Title, entry.ID)
			fmt.Printf("     Username: %s\n", entry.Username)
			fmt.Printf("     Password: %s\n", entry.Password)
			if entry.URL != "" {
				fmt.Printf("     URL: %s\n", entry.URL)
			}
			if entry.Notes != "" {
				fmt.Printf("     Notes: %s\n", entry.Notes)
			}
		}
	} else {
		if len(path.GroupIDs) > 0 {
			fmt.Println("\nNo entries found in current group.")
		}
	}
}

func handleUpdateGroup(service *WalletService, path Path, scanner *bufio.Scanner) {
	if len(path.GroupIDs) == 0 {
		fmt.Println("Cannot update root groups directly.")
		return
	}

	group, err := FindGroupByPath(service.GetWallet(), path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Current group name: %s\n", group.Name)
	fmt.Print("Enter new group name: ")
	if !scanner.Scan() {
		return
	}
	newName := strings.TrimSpace(scanner.Text())
	if newName == "" {
		fmt.Println("Group name cannot be empty")
		return
	}

	updatedGroup := Group{
		Name:    newName,
		Groups:  group.Groups,
		Entries: group.Entries,
	}

	if err := service.UpdateGroup(path, updatedGroup); err != nil {
		fmt.Printf("Error updating group: %v\n", err)
		return
	}

	fmt.Println("Group updated successfully!")
	if err := service.Save(); err != nil {
		fmt.Printf("Error saving: %v\n", err)
	}
}

func handleUpdateEntry(service *WalletService, path Path, scanner *bufio.Scanner) {
	if len(path.GroupIDs) == 0 {
		fmt.Println("No entries at root level.")
		return
	}

	group, err := FindGroupByPath(service.GetWallet(), path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(group.Entries) == 0 {
		fmt.Println("No entries found in current group.")
		return
	}

	fmt.Println("\nEntries:")
	for i, entry := range group.Entries {
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, entry.Title, entry.ID)
		fmt.Printf("     Username: %s\n", entry.Username)
		fmt.Printf("     Password: %s\n", entry.Password)
		if entry.URL != "" {
			fmt.Printf("     URL: %s\n", entry.URL)
		}
		if entry.Notes != "" {
			fmt.Printf("     Notes: %s\n", entry.Notes)
		}
	}
	fmt.Print("\nEnter entry number to update: ")
	if !scanner.Scan() {
		return
	}

	var entryNum int
	if _, err := fmt.Sscanf(scanner.Text(), "%d", &entryNum); err != nil {
		fmt.Println("Invalid entry number")
		return
	}

	currentGroup, err := FindGroupByPath(service.GetWallet(), path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if entryNum < 1 || entryNum > len(currentGroup.Entries) {
		fmt.Println("Invalid entry number")
		return
	}

	entry := currentGroup.Entries[entryNum-1]
	entryPath := Path{
		GroupIDs: path.GroupIDs,
		EntryID:  entry.ID,
	}

	fmt.Printf("Current entry: %s\n", entry.Title)
	fmt.Print("Enter new title (press Enter to keep current): ")
	if !scanner.Scan() {
		return
	}
	newTitle := strings.TrimSpace(scanner.Text())
	if newTitle == "" {
		newTitle = entry.Title
	}

	fmt.Print("Enter new username (press Enter to keep current): ")
	if !scanner.Scan() {
		return
	}
	newUsername := strings.TrimSpace(scanner.Text())
	if newUsername == "" {
		newUsername = entry.Username
	}

	fmt.Print("Enter new password (press Enter to keep current): ")
	if !scanner.Scan() {
		return
	}
	newPassword := strings.TrimSpace(scanner.Text())
	if newPassword == "" {
		newPassword = entry.Password
	}

	fmt.Print("Enter new URL (press Enter to keep current): ")
	if !scanner.Scan() {
		return
	}
	newURL := strings.TrimSpace(scanner.Text())
	if newURL == "" {
		newURL = entry.URL
	}

	fmt.Print("Enter new notes (press Enter to keep current): ")
	if !scanner.Scan() {
		return
	}
	newNotes := strings.TrimSpace(scanner.Text())
	if newNotes == "" {
		newNotes = entry.Notes
	}

	updatedEntry := Entry{
		Title:    newTitle,
		Username: newUsername,
		Password: newPassword,
		URL:      newURL,
		Notes:    newNotes,
	}

	if err := service.UpdateEntry(entryPath, updatedEntry); err != nil {
		fmt.Printf("Error updating entry: %v\n", err)
		return
	}

	fmt.Println("Entry updated successfully!")
	if err := service.Save(); err != nil {
		fmt.Printf("Error saving: %v\n", err)
	}
}

func handleDeleteGroup(service *WalletService, path Path, scanner *bufio.Scanner) {
	// List groups at current location
	var groups []Group
	var parentPath Path
	if len(path.GroupIDs) == 0 {
		// At root - can delete root groups
		groups = service.GetWallet().Groups
		parentPath = Path{GroupIDs: []string{}}
	} else {
		// At a group - list its subgroups
		parentPath = path
		group, err := FindGroupByPath(service.GetWallet(), path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		groups = group.Groups
	}

	if len(groups) == 0 {
		fmt.Println("No groups to delete.")
		return
	}

	fmt.Println("\nGroups:")
	for i, group := range groups {
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, group.Name, group.ID)
	}

	fmt.Print("\nEnter group number to delete: ")
	if !scanner.Scan() {
		return
	}

	var groupNum int
	if _, err := fmt.Sscanf(scanner.Text(), "%d", &groupNum); err != nil {
		fmt.Println("Invalid group number")
		return
	}

	if groupNum < 1 || groupNum > len(groups) {
		fmt.Println("Invalid group number")
		return
	}

	groupToDelete := groups[groupNum-1]
	deletePath := Path{
		GroupIDs: append(parentPath.GroupIDs, groupToDelete.ID),
	}

	fmt.Printf("Are you sure you want to delete group '%s'? (yes/no): ", groupToDelete.Name)
	if !scanner.Scan() {
		return
	}
	confirm := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if confirm != "yes" {
		fmt.Println("Deletion cancelled")
		return
	}

	if err := service.DeleteGroup(deletePath); err != nil {
		fmt.Printf("Error deleting group: %v\n", err)
		return
	}

	fmt.Println("Group deleted successfully!")
	if err := service.Save(); err != nil {
		fmt.Printf("Error saving: %v\n", err)
	}
}

func handleDeleteEntry(service *WalletService, path Path, scanner *bufio.Scanner) {
	if len(path.GroupIDs) == 0 {
		fmt.Println("No entries at root level.")
		return
	}

	group, err := FindGroupByPath(service.GetWallet(), path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(group.Entries) == 0 {
		fmt.Println("No entries found in current group.")
		return
	}

	fmt.Println("\nEntries:")
	for i, entry := range group.Entries {
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, entry.Title, entry.ID)
		fmt.Printf("     Username: %s\n", entry.Username)
		fmt.Printf("     Password: %s\n", entry.Password)
		if entry.URL != "" {
			fmt.Printf("     URL: %s\n", entry.URL)
		}
		if entry.Notes != "" {
			fmt.Printf("     Notes: %s\n", entry.Notes)
		}
	}
	fmt.Print("\nEnter entry number to delete: ")
	if !scanner.Scan() {
		return
	}

	var entryNum int
	if _, err := fmt.Sscanf(scanner.Text(), "%d", &entryNum); err != nil {
		fmt.Println("Invalid entry number")
		return
	}

	currentGroup, err := FindGroupByPath(service.GetWallet(), path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if entryNum < 1 || entryNum > len(currentGroup.Entries) {
		fmt.Println("Invalid entry number")
		return
	}

	entry := currentGroup.Entries[entryNum-1]
	entryPath := Path{
		GroupIDs: path.GroupIDs,
		EntryID:  entry.ID,
	}

	fmt.Printf("Are you sure you want to delete entry '%s'? (yes/no): ", entry.Title)
	if !scanner.Scan() {
		return
	}
	confirm := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if confirm != "yes" {
		fmt.Println("Deletion cancelled")
		return
	}

	if err := service.DeleteEntry(entryPath); err != nil {
		fmt.Printf("Error deleting entry: %v\n", err)
		return
	}

	fmt.Println("Entry deleted successfully!")
	if err := service.Save(); err != nil {
		fmt.Printf("Error saving: %v\n", err)
	}
}

func handleTraverseForward(service *WalletService, currentPath Path, scanner *bufio.Scanner) Path {
	// Get groups and entries one level down from current location
	var groups []Group
	if len(currentPath.GroupIDs) == 0 {
		// At root - get root groups
		groups = service.GetWallet().Groups
	} else {
		group, err := FindGroupByPath(service.GetWallet(), currentPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return currentPath
		}
		groups = group.Groups
	}

	if len(groups) == 0 {
		fmt.Println("No groups available one level down from current location.")
		return currentPath
	}

	// List available groups
	fmt.Println("\nGroups one level down:")
	for i, group := range groups {
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, group.Name, group.ID)
	}

	fmt.Print("\nEnter group number to navigate into: ")
	if !scanner.Scan() {
		return currentPath
	}

	var groupNum int
	if _, err := fmt.Sscanf(scanner.Text(), "%d", &groupNum); err != nil {
		fmt.Println("Invalid group number")
		return currentPath
	}

	if groupNum < 1 || groupNum > len(groups) {
		fmt.Println("Invalid group number")
		return currentPath
	}

	selectedGroup := groups[groupNum-1]
	newPath := Path{
		GroupIDs: append(currentPath.GroupIDs, selectedGroup.ID),
	}

	fmt.Printf("Navigated to group: %s\n", selectedGroup.Name)
	return newPath
}

func handleTraverseBackward(service *WalletService, currentPath Path) Path {
	// Go up one level in the hierarchy
	if len(currentPath.GroupIDs) == 0 {
		fmt.Println("Already at root level. Cannot go up further.")
		return currentPath
	}

	parentPath := GetParentPath(currentPath)

	// Get parent group name for display
	if len(parentPath.GroupIDs) == 0 {
		fmt.Println("Moved up to root level")
	} else {
		parentGroup, err := FindGroupByPath(service.GetWallet(), parentPath)
		if err == nil {
			fmt.Printf("Moved up to group: %s\n", parentGroup.Name)
		} else {
			fmt.Println("Moved up one level")
		}
	}

	return parentPath
}

func handleSearchEntry(service *WalletService, scanner *bufio.Scanner) {
	fmt.Print("Enter entry title to search: ")
	if !scanner.Scan() {
		return
	}
	searchTitle := strings.TrimSpace(scanner.Text())
	if searchTitle == "" {
		fmt.Println("Search title cannot be empty")
		return
	}

	var foundEntries []PathInfo
	service.TraverseForward(func(info PathInfo) bool {
		if info.IsEntry && info.Entry != nil {
			if strings.Contains(strings.ToLower(info.Entry.Title), strings.ToLower(searchTitle)) {
				foundEntries = append(foundEntries, info)
			}
		}
		return true
	})

	if len(foundEntries) == 0 {
		fmt.Printf("No entries found matching '%s'\n", searchTitle)
		return
	}

	fmt.Printf("\nFound %d entry/entries:\n", len(foundEntries))
	for i, info := range foundEntries {
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, info.Entry.Title, info.Entry.ID)
		fmt.Printf("     Username: %s\n", info.Entry.Username)
		fmt.Printf("     Password: %s\n", info.Entry.Password)
		if info.Entry.URL != "" {
			fmt.Printf("     URL: %s\n", info.Entry.URL)
		}
		if info.Entry.Notes != "" {
			fmt.Printf("     Notes: %s\n", info.Entry.Notes)
		}
		fmt.Printf("     Path: %v\n", info.Path.GroupIDs)
	}
}

func handleDisplayTree(service *WalletService, currentPath Path) {
	wallet := service.GetWallet()
	if wallet == nil {
		fmt.Println("Wallet not loaded")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("WALLET TREE STRUCTURE")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("(Current location marked with >>>)")
	fmt.Println()

	var displayTree func(groups []Group, path Path, depth int, prefix string, isLast bool)
	displayTree = func(groups []Group, path Path, depth int, prefix string, isLast bool) {
		for i, group := range groups {
			isLastGroup := i == len(groups)-1
			currentPrefix := prefix
			if depth > 0 {
				if isLast {
					currentPrefix += "    "
				} else {
					currentPrefix += "│   "
				}
			}

			// Check if this is the current location
			groupPath := Path{GroupIDs: append([]string{}, path.GroupIDs...)}
			groupPath.GroupIDs = append(groupPath.GroupIDs, group.ID)
			isCurrent := len(currentPath.GroupIDs) == len(groupPath.GroupIDs)
			if isCurrent {
				for j := 0; j < len(groupPath.GroupIDs); j++ {
					if j >= len(currentPath.GroupIDs) || groupPath.GroupIDs[j] != currentPath.GroupIDs[j] {
						isCurrent = false
						break
					}
				}
			}

			// Display group
			connector := "├── "
			if isLastGroup {
				connector = "└── "
			}

			marker := ""
			if isCurrent {
				marker = " >>> [CURRENT]"
			}

			fmt.Printf("%s%s%s%s%s\n", prefix, connector, group.Name, marker, fmt.Sprintf(" (%d subgroups, %d entries)", len(group.Groups), len(group.Entries)))

			// Display entries in this group
			newPrefix := prefix
			if depth > 0 {
				if isLast {
					newPrefix += "    "
				} else {
					newPrefix += "│   "
				}
			}
			if isLastGroup {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}

			for j, entry := range group.Entries {
				isLastEntry := j == len(group.Entries)-1 && len(group.Groups) == 0
				entryConnector := "├── "
				if isLastEntry {
					entryConnector = "└── "
				}

				entryPath := Path{
					GroupIDs: append([]string{}, groupPath.GroupIDs...),
					EntryID:  entry.ID,
				}
				isCurrentEntry := len(currentPath.GroupIDs) == len(entryPath.GroupIDs) &&
					currentPath.EntryID == entry.ID
				if isCurrentEntry {
					for k := 0; k < len(entryPath.GroupIDs); k++ {
						if k >= len(currentPath.GroupIDs) || entryPath.GroupIDs[k] != currentPath.GroupIDs[k] {
							isCurrentEntry = false
							break
						}
					}
				}

				entryMarker := ""
				if isCurrentEntry {
					entryMarker = " >>> [CURRENT]"
				}

				fmt.Printf("%s%s%s%s\n", newPrefix, entryConnector, entry.Title, entryMarker)
			}

			// Recursively display nested groups
			nextPrefix := prefix
			if depth > 0 {
				if isLast {
					nextPrefix += "    "
				} else {
					nextPrefix += "│   "
				}
			}
			if isLastGroup {
				nextPrefix += "    "
			} else {
				nextPrefix += "│   "
			}

			displayTree(group.Groups, groupPath, depth+1, nextPrefix, isLastGroup)
		}
	}

	// Check if at root
	isAtRoot := len(currentPath.GroupIDs) == 0
	if isAtRoot {
		fmt.Println("ROOT >>> [CURRENT]")
	} else {
		fmt.Println("ROOT")
	}

	if len(wallet.Groups) == 0 {
		fmt.Println("  (empty)")
	} else {
		displayTree(wallet.Groups, Path{GroupIDs: []string{}}, 0, "", false)
	}

	fmt.Println()
}

func handleNavigateIntoGroup(service *WalletService, currentPath Path, scanner *bufio.Scanner) Path {
	var groups []Group
	if len(currentPath.GroupIDs) == 0 {
		groups = service.GetWallet().Groups
	} else {
		group, err := FindGroupByPath(service.GetWallet(), currentPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return currentPath
		}
		groups = group.Groups
	}

	if len(groups) == 0 {
		fmt.Println("No groups available to navigate into.")
		return currentPath
	}

	fmt.Print("\nEnter group number to navigate into: ")
	if !scanner.Scan() {
		return currentPath
	}

	var groupNum int
	if _, err := fmt.Sscanf(scanner.Text(), "%d", &groupNum); err != nil {
		fmt.Println("Invalid group number")
		return currentPath
	}

	if groupNum < 1 || groupNum > len(groups) {
		fmt.Println("Invalid group number")
		return currentPath
	}

	selectedGroup := groups[groupNum-1]
	newPath := Path{
		GroupIDs: append(currentPath.GroupIDs, selectedGroup.ID),
	}

	fmt.Printf("Navigated into group: %s\n", selectedGroup.Name)
	return newPath
}
