# Safe Wallet Go

A secure password storage application written in Go that encrypts password files using password-based encryption with salt and hashing.

## Features

- **Password-based encryption**: Uses AES-256-GCM with PBKDF2 key derivation
- **Nested group structure**: Supports hierarchical organization of password entries
- **Path-aware traversal**: Forward and backward traversal of nested groups
- **Secure storage**: Encrypted file storage with proper file permissions

## Structure

The wallet uses a JSON structure with nested groups and entries:

```json
{
  "version": 1,
  "groups": [
    {
      "id": "grp-root",
      "name": "Root",
      "groups": [...],
      "entries": [...]
    }
  ]
}
```

## Files

- `models.go`: Data structures for Wallet, Group, Entry, and Path
- `crypto.go`: Encryption/decryption functions using AES-GCM
- `storage.go`: File read/write operations
- `traversal.go`: Path-aware traversal functions (forward and backward)
- `wallet.go`: High-level service API for wallet operations
- `main.go`: Example usage

## Usage

### Basic Operations

```go
// Create a new wallet service
service := NewWalletService("wallet.dat", "your-password")

// Create a new wallet
service.CreateNew()

// Or load an existing wallet
service.Load()

// Add a group
group := Group{
    ID:   "grp-1",
    Name: "Work",
}
path := Path{GroupIDs: []string{"grp-root"}}
service.AddGroup(path, group)

// Add an entry
entry := Entry{
    ID:       "ent-1",
    Title:    "AWS Console",
    Username: "admin",
    Password: "password123",
    URL:      "https://aws.amazon.com",
    Notes:    "MFA enabled",
}
entryPath := Path{GroupIDs: []string{"grp-root", "grp-1"}}
service.AddEntry(entryPath, entry)

// Save the wallet
service.Save()
```

### Traversal

```go
// Forward traversal (depth-first)
service.TraverseForward(func(info PathInfo) bool {
    if info.IsEntry {
        fmt.Println("Entry:", info.Entry.Title)
    } else {
        fmt.Println("Group:", info.Group.Name)
    }
    return true // Continue traversal
})

// Backward traversal (reverse depth-first)
service.TraverseBackward(func(info PathInfo) bool {
    // Process items in reverse order
    return true
})
```

### Path Operations

```go
// Find group by ID
path, group, err := service.FindGroupByID("grp-1")

// Find entry by ID
path, entry, err := service.FindEntryByID("ent-1")

// Get parent path
parentPath := GetParentPath(path)
```

## Security

- Uses AES-256-GCM for encryption
- PBKDF2 with 100,000 iterations for key derivation
- Random salt (32 bytes) and nonce (12 bytes) for each encryption
- File permissions set to 0600 (read/write for owner only)

## Building

```bash
cd safe-wallet-go
go mod tidy
go build
```

## Running

```bash
go run .
```

