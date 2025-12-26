package pkg

// Wallet represents the root structure of the password storage
type Wallet struct {
	Version int     `json:"version"`
	Groups  []Group `json:"groups"`
}

// Group represents a group that can contain other groups and entries
type Group struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Groups  []Group `json:"groups"`
	Entries []Entry `json:"entries"`
}

// Entry represents a password entry with flexible, user-defined fields
type Entry struct {
	ID     string       `json:"id"`
	Title  string       `json:"title"`
	Fields []EntryField `json:"fields"`
}

// EntryField represents a key-value pair for an entry's field
type EntryField struct {
	Name  string    `json:"name"`
	Value string    `json:"value"`
	Type  FieldType `json:"type"`
}

// FieldType defines the type of an entry field
type FieldType string

const (
	// FieldTypeGeneral is for general purpose fields like username, notes, etc.
	FieldTypeGeneral FieldType = "general"
	// FieldTypePassword is for password fields that should be masked
	FieldTypePassword FieldType = "password"
	// FieldTypePIN is for PIN fields that should be masked and only accept numeric
	FieldTypePIN FieldType = "pin"
)

// Path represents a path to a group or entry
type Path struct {
	GroupIDs []string // Path of group IDs from root to target
	EntryID  string   // Entry ID if targeting an entry, empty if targeting a group
}

// PathInfo contains information about a traversed item
type PathInfo struct {
	Path    Path
	Group   *Group
	Entry   *Entry
	Depth   int
	IsEntry bool
}
