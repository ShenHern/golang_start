package pkg

import (
	"errors"
)

// WalletService provides high-level operations on the wallet
type WalletService struct {
	wallet   *Wallet
	filepath string
	password string
}

// NewWalletService creates a new wallet service instance
func NewWalletService(filepath string, password string) *WalletService {
	return &WalletService{
		filepath: filepath,
		password: password,
	}
}

// Load loads the wallet from the file
func (ws *WalletService) Load() error {
	wallet, err := LoadWallet(ws.filepath, ws.password)
	if err != nil {
		return err
	}
	ws.wallet = wallet
	return nil
}

// Save saves the wallet to the file
func (ws *WalletService) Save() error {
	if ws.wallet == nil {
		return errors.New("wallet not loaded")
	}
	return SaveWallet(ws.wallet, ws.filepath, ws.password)
}

// CreateNew creates a new wallet and saves it
func (ws *WalletService) CreateNew() error {
	ws.wallet = CreateNewWallet()
	return ws.Save()
}

// GetWallet returns the current wallet
func (ws *WalletService) GetWallet() *Wallet {
	return ws.wallet
}

// AddGroup adds a group at the specified path
func (ws *WalletService) AddGroup(path Path, group *Group) error {
	if ws.wallet == nil {
		return errors.New("wallet not loaded")
	}

	// Auto-generate ID if not provided or empty
	if group.ID == "" {
		group.ID = generateGroupID()
	} else {
		// Check if ID already exists
		if checkGroupIDExists(ws.wallet, group.ID) {
			return errors.New("group ID already exists")
		}
	}

	// Check if group name already exists
	if checkGroupNameExists(ws.wallet, group.Name, "") {
		return errors.New("group name already exists")
	}

	// Initialize empty slices if nil
	if group.Groups == nil {
		group.Groups = []Group{}
	}
	if group.Entries == nil {
		group.Entries = []Entry{}
	}

	// If path is empty, add to root
	if len(path.GroupIDs) == 0 {
		ws.wallet.Groups = append(ws.wallet.Groups, *group)
		return nil
	}

	// Find parent group
	parentGroup, err := FindGroupByPath(ws.wallet, path)
	if err != nil {
		return err
	}

	parentGroup.Groups = append(parentGroup.Groups, *group)
	return nil
}

// AddEntry adds an entry to the group at the specified path
func (ws *WalletService) AddEntry(path Path, entry *Entry) error {
	if ws.wallet == nil {
		return errors.New("wallet not loaded")
	}

	if path.EntryID != "" {
		return errors.New("path should point to a group, not an entry")
	}

	// Auto-generate ID if not provided or empty
	if entry.ID == "" {
		entry.ID = generateEntryID()
	} else {
		// Check if ID already exists
		if checkEntryIDExists(ws.wallet, entry.ID) {
			return errors.New("entry ID already exists")
		}
	}

	// Check if entry title already exists
	if checkEntryTitleExists(ws.wallet, entry.Title, "") {
		return errors.New("entry title already exists")
	}

	group, err := FindGroupByPath(ws.wallet, path)
	if err != nil {
		return err
	}

	group.Entries = append(group.Entries, *entry)
	return nil
}

// UpdateGroup updates a group at the specified path
func (ws *WalletService) UpdateGroup(path Path, updatedGroup Group) error {
	if ws.wallet == nil {
		return errors.New("wallet not loaded")
	}

	if len(path.GroupIDs) == 0 {
		return errors.New("cannot update root groups directly")
	}

	targetID := path.GroupIDs[len(path.GroupIDs)-1]

	// Check if the new name conflicts with another group (excluding the current one)
	if checkGroupNameExists(ws.wallet, updatedGroup.Name, targetID) {
		return errors.New("group name already exists")
	}

	// Get parent path
	parentPath := GetParentPath(path)
	if len(parentPath.GroupIDs) == 0 {
		// Update root-level group
		for i := range ws.wallet.Groups {
			if ws.wallet.Groups[i].ID == targetID {
				updatedGroup.ID = ws.wallet.Groups[i].ID
				updatedGroup.Groups = ws.wallet.Groups[i].Groups
				updatedGroup.Entries = ws.wallet.Groups[i].Entries
				ws.wallet.Groups[i] = updatedGroup
				return nil
			}
		}
		return errors.New("group not found")
	}

	parentGroup, err := FindGroupByPath(ws.wallet, parentPath)
	if err != nil {
		return err
	}

	for i := range parentGroup.Groups {
		if parentGroup.Groups[i].ID == targetID {
			updatedGroup.ID = parentGroup.Groups[i].ID
			updatedGroup.Groups = parentGroup.Groups[i].Groups
			updatedGroup.Entries = parentGroup.Groups[i].Entries
			parentGroup.Groups[i] = updatedGroup
			return nil
		}
	}

	return errors.New("group not found")
}

// UpdateEntry updates an entry at the specified path
func (ws *WalletService) UpdateEntry(path Path, updatedEntry Entry) error {
	if ws.wallet == nil {
		return errors.New("wallet not loaded")
	}

	if path.EntryID == "" {
		return errors.New("path must include an entry ID")
	}

	entry, err := FindEntryByPath(ws.wallet, path)
	if err != nil {
		return err
	}

	// Check if the new title conflicts with another entry (excluding the current one)
	if checkEntryTitleExists(ws.wallet, updatedEntry.Title, path.EntryID) {
		return errors.New("entry title already exists")
	}

	updatedEntry.ID = entry.ID
	*entry = updatedEntry
	return nil
}

// DeleteGroup deletes a group at the specified path
func (ws *WalletService) DeleteGroup(path Path) error {
	if ws.wallet == nil {
		return errors.New("wallet not loaded")
	}

	if len(path.GroupIDs) == 0 {
		return errors.New("cannot delete root groups directly")
	}

	parentPath := GetParentPath(path)
	if len(parentPath.GroupIDs) == 0 {
		// Delete root-level group
		for i, group := range ws.wallet.Groups {
			if group.ID == path.GroupIDs[len(path.GroupIDs)-1] {
				ws.wallet.Groups = append(ws.wallet.Groups[:i], ws.wallet.Groups[i+1:]...)
				return nil
			}
		}
		return errors.New("group not found")
	}

	parentGroup, err := FindGroupByPath(ws.wallet, parentPath)
	if err != nil {
		return err
	}

	targetID := path.GroupIDs[len(path.GroupIDs)-1]
	for i, group := range parentGroup.Groups {
		if group.ID == targetID {
			parentGroup.Groups = append(parentGroup.Groups[:i], parentGroup.Groups[i+1:]...)
			return nil
		}
	}

	return errors.New("group not found")
}

// DeleteEntry deletes an entry at the specified path
func (ws *WalletService) DeleteEntry(path Path) error {
	if ws.wallet == nil {
		return errors.New("wallet not loaded")
	}

	if path.EntryID == "" {
		return errors.New("path must include an entry ID")
	}

	group, err := FindGroupByPath(ws.wallet, path)
	if err != nil {
		return err
	}

	for i, entry := range group.Entries {
		if entry.ID == path.EntryID {
			group.Entries = append(group.Entries[:i], group.Entries[i+1:]...)
			return nil
		}
	}

	return errors.New("entry not found")
}

// FindGroupByID finds a group by its ID and returns its path
func (ws *WalletService) FindGroupByID(groupID string) (Path, *Group, error) {
	if ws.wallet == nil {
		return Path{}, nil, errors.New("wallet not loaded")
	}

	path, err := GetPathToGroup(ws.wallet, groupID)
	if err != nil {
		return Path{}, nil, err
	}

	group, err := FindGroupByPath(ws.wallet, path)
	if err != nil {
		return Path{}, nil, err
	}

	return path, group, nil
}

// FindEntryByID finds an entry by its ID and returns its path
func (ws *WalletService) FindEntryByID(entryID string) (Path, *Entry, error) {
	if ws.wallet == nil {
		return Path{}, nil, errors.New("wallet not loaded")
	}

	path, err := GetPathToEntry(ws.wallet, entryID)
	if err != nil {
		return Path{}, nil, err
	}

	entry, err := FindEntryByPath(ws.wallet, path)
	if err != nil {
		return Path{}, nil, err
	}

	return path, entry, nil
}

// TraverseForward performs forward traversal with a callback
func (ws *WalletService) TraverseForward(callback func(info PathInfo) bool) {
	if ws.wallet == nil {
		return
	}
	TraverseForward(ws.wallet, callback)
}

// TraverseBackward performs backward traversal with a callback
func (ws *WalletService) TraverseBackward(callback func(info PathInfo) bool) {
	if ws.wallet == nil {
		return
	}
	TraverseBackward(ws.wallet, callback)
}
