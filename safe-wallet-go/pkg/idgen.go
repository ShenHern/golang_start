package pkg

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// generateGroupID generates a unique group ID
func generateGroupID() string {
	return generateUniqueID("grp")
}

// generateEntryID generates a unique entry ID
func generateEntryID() string {
	return generateUniqueID("ent")
}

// generateUniqueID generates a unique ID with the given prefix
func generateUniqueID(prefix string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(bytes))
}

// checkGroupIDExists checks if a group ID already exists in the wallet
func checkGroupIDExists(wallet *Wallet, groupID string) bool {
	var exists bool
	TraverseForward(wallet, func(info PathInfo) bool {
		if !info.IsEntry && info.Group != nil && info.Group.ID == groupID {
			exists = true
			return false // Stop traversal
		}
		return true
	})
	return exists
}

// checkEntryIDExists checks if an entry ID already exists in the wallet
func checkEntryIDExists(wallet *Wallet, entryID string) bool {
	var exists bool
	TraverseForward(wallet, func(info PathInfo) bool {
		if info.IsEntry && info.Entry != nil && info.Entry.ID == entryID {
			exists = true
			return false // Stop traversal
		}
		return true
	})
	return exists
}

// checkGroupNameExists checks if a group name already exists in the wallet
func checkGroupNameExists(wallet *Wallet, groupName string, excludeID string) bool {
	var exists bool
	TraverseForward(wallet, func(info PathInfo) bool {
		if !info.IsEntry && info.Group != nil && info.Group.Name == groupName {
			// If excludeID is set, skip the group with that ID (for updates)
			if excludeID == "" || info.Group.ID != excludeID {
				exists = true
				return false // Stop traversal
			}
		}
		return true
	})
	return exists
}

// checkEntryTitleExists checks if an entry title already exists in the wallet
func checkEntryTitleExists(wallet *Wallet, entryTitle string, excludeID string) bool {
	var exists bool
	TraverseForward(wallet, func(info PathInfo) bool {
		if info.IsEntry && info.Entry != nil && info.Entry.Title == entryTitle {
			// If excludeID is set, skip the entry with that ID (for updates)
			if excludeID == "" || info.Entry.ID != excludeID {
				exists = true
				return false // Stop traversal
			}
		}
		return true
	})
	return exists
}
