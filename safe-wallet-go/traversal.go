package main

import (
	"errors"
)

// FindGroupByPath finds a group by its path (list of group IDs)
func FindGroupByPath(wallet *Wallet, path Path) (*Group, error) {
	if len(path.GroupIDs) == 0 {
		return nil, errors.New("empty path")
	}

	currentGroups := wallet.Groups
	var targetGroup *Group

	// Traverse forward through the path
	for i, groupID := range path.GroupIDs {
		found := false
		for j := range currentGroups {
			if currentGroups[j].ID == groupID {
				targetGroup = &currentGroups[j]
				if i < len(path.GroupIDs)-1 {
					// Not the last group in path, continue traversing
					currentGroups = currentGroups[j].Groups
					found = true
					break
				} else {
					// This is the target group
					return targetGroup, nil
				}
			}
		}
		if !found {
			return nil, errors.New("group not found in path")
		}
	}

	return targetGroup, nil
}

// FindEntryByPath finds an entry by its path (group IDs + entry ID)
func FindEntryByPath(wallet *Wallet, path Path) (*Entry, error) {
	if path.EntryID == "" {
		return nil, errors.New("entry ID is empty")
	}

	// Find the parent group
	group, err := FindGroupByPath(wallet, path)
	if err != nil {
		return nil, err
	}

	// Search for the entry in the group
	for i := range group.Entries {
		if group.Entries[i].ID == path.EntryID {
			return &group.Entries[i], nil
		}
	}

	return nil, errors.New("entry not found")
}

// TraverseForward performs forward traversal (depth-first) of all groups and entries
func TraverseForward(wallet *Wallet, callback func(info PathInfo) bool) {
	var traverse func(groups []Group, currentPath Path, depth int) bool
	traverse = func(groups []Group, currentPath Path, depth int) bool {
		for i := range groups {
			group := &groups[i]

			// Process group
			info := PathInfo{
				Path:    currentPath,
				Group:   group,
				Depth:   depth,
				IsEntry: false,
			}
			if !callback(info) {
				return false
			}

			// Process entries in this group
			for j := range group.Entries {
				entryPath := Path{
					GroupIDs: append([]string{}, currentPath.GroupIDs...),
					EntryID:  group.Entries[j].ID,
				}
				entryInfo := PathInfo{
					Path:    entryPath,
					Entry:   &group.Entries[j],
					Depth:   depth,
					IsEntry: true,
				}
				if !callback(entryInfo) {
					return false
				}
			}

			// Recursively traverse nested groups
			newPath := Path{
				GroupIDs: append(currentPath.GroupIDs, group.ID),
			}
			if !traverse(group.Groups, newPath, depth+1) {
				return false
			}
		}
		return true
	}

	rootPath := Path{GroupIDs: []string{}}
	traverse(wallet.Groups, rootPath, 0)
}

// TraverseBackward performs backward traversal (reverse depth-first) of all groups and entries
func TraverseBackward(wallet *Wallet, callback func(info PathInfo) bool) {
	// First, collect all items in forward order
	var items []PathInfo
	TraverseForward(wallet, func(info PathInfo) bool {
		items = append(items, info)
		return true
	})

	// Then process them in reverse order
	for i := len(items) - 1; i >= 0; i-- {
		if !callback(items[i]) {
			break
		}
	}
}

// GetPathToGroup returns the path to a group by its ID
func GetPathToGroup(wallet *Wallet, targetGroupID string) (Path, error) {
	var foundPath Path
	found := false

	var findPath func(groups []Group, currentPath Path) bool
	findPath = func(groups []Group, currentPath Path) bool {
		for i := range groups {
			group := &groups[i]
			newPath := Path{
				GroupIDs: append(currentPath.GroupIDs, group.ID),
			}

			if group.ID == targetGroupID {
				foundPath = newPath
				found = true
				return true
			}

			if findPath(group.Groups, newPath) {
				return true
			}
		}
		return false
	}

	rootPath := Path{GroupIDs: []string{}}
	findPath(wallet.Groups, rootPath)

	if !found {
		return Path{}, errors.New("group not found")
	}

	return foundPath, nil
}

// GetPathToEntry returns the path to an entry by its ID
func GetPathToEntry(wallet *Wallet, targetEntryID string) (Path, error) {
	var foundPath Path
	found := false

	var findPath func(groups []Group, currentPath Path) bool
	findPath = func(groups []Group, currentPath Path) bool {
		for i := range groups {
			group := &groups[i]
			newPath := Path{
				GroupIDs: append(currentPath.GroupIDs, group.ID),
			}

			// Check entries in this group
			for j := range group.Entries {
				if group.Entries[j].ID == targetEntryID {
					foundPath = Path{
						GroupIDs: newPath.GroupIDs,
						EntryID:  targetEntryID,
					}
					found = true
					return true
				}
			}

			if findPath(group.Groups, newPath) {
				return true
			}
		}
		return false
	}

	rootPath := Path{GroupIDs: []string{}}
	findPath(wallet.Groups, rootPath)

	if !found {
		return Path{}, errors.New("entry not found")
	}

	return foundPath, nil
}

// GetParentPath returns the path to the parent of the given path
func GetParentPath(path Path) Path {
	if len(path.GroupIDs) == 0 {
		return Path{GroupIDs: []string{}}
	}
	return Path{
		GroupIDs: path.GroupIDs[:len(path.GroupIDs)-1],
	}
}

// GetRootGroups returns all root-level groups
func GetRootGroups(wallet *Wallet) []Group {
	return wallet.Groups
}
