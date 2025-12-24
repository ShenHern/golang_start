package main

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

// Entry represents a password entry
type Entry struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
	Notes    string `json:"notes"`
}

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
