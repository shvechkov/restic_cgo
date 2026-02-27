# Restic Library - Summary

## Created Files

This implementation adds a new library package to restic that exposes backup and restore functionality for S3/Wasabi storage. **No existing restic code has been modified.**

### Library Files

1. **pkg/resticlib/client.go** (main library file)
   - `Config` struct for S3/Wasabi configuration
   - `Client` struct for managing repository connections
   - `InitRepository()` - Initialize a new repository
   - `NewClient()` - Create a client to an existing repository  
   - `Backup()` - Create backups with full options
   - `Restore()` - Restore from snapshots
   - `ListSnapshots()` - List all snapshots
   - `Forget()` - Delete snapshots
   - `Check()` - Verify repository integrity
   - `Close()` - Clean up resources

2. **pkg/resticlib/doc.go**
   - Package documentation with examples

3. **pkg/resticlib/README.md**
   - Comprehensive library documentation
   - API reference
   - Quick start guide
   - Configuration examples for S3 and Wasabi
   - Security considerations

4. **pkg/resticlib/USAGE.md**
   - Detailed usage guide
   - Code examples for all operations
   - Best practices
   - Troubleshooting
   - Advanced patterns

5. **pkg/resticlib/client_test.go**
   - Integration tests
   - Configuration validation tests
   - Test helpers

6. **pkg/resticlib/examples/main.go**
   - Complete working example application
   - Commands: init, backup, restore, list, test
   - Comprehensive test suite
   - Environment variable configuration

## Features

### Core Functionality
- ✅ Repository initialization on S3/Wasabi
- ✅ Backup files and directories
- ✅ Restore from snapshots (including "latest")
- ✅ List all snapshots
- ✅ Delete snapshots
- ✅ Repository integrity checking
- ✅ Full S3 and Wasabi support

### Configuration Options
- S3/Wasabi endpoint configuration
- Access key management
- Region selection
- Optional path prefix
- Concurrent connection control
- HTTP/HTTPS selection

### Example Application
- Initialize repositories
- Create backups with tags
- Restore to specified locations
- List snapshots with details
- Comprehensive test suite

## Usage

### Quick Start

```bash
# Set environment variables
export RESTIC_S3_ENDPOINT="s3.wasabisys.com"
export RESTIC_S3_ACCESS_KEY="your-access-key"
export RESTIC_S3_SECRET_KEY="your-secret-key"
export RESTIC_S3_BUCKET="your-bucket"
export RESTIC_S3_REGION="us-east-1"
export RESTIC_PASSWORD="your-password"

# Build the example
cd pkg/resticlib/examples
go build -o resticlib-example

# Initialize repository
./resticlib-example init

# Create a backup
./resticlib-example backup /path/to/data

# List snapshots
./resticlib-example list

# Restore latest snapshot
./resticlib-example restore /path/to/restore

# Run comprehensive tests
./resticlib-example test
```

### Programmatic Usage

```go
package main

import (
    "context"
    "log"
    "github.com/restic/restic/pkg/resticlib"
)

func main() {
    cfg := &resticlib.Config{
        Endpoint:        "s3.wasabisys.com",
        AccessKeyID:     "your-key",
        SecretAccessKey: "your-secret",
        BucketName:      "your-bucket",
        Region:          "us-east-1",
        Password:        "your-password",
    }
    
    ctx := context.Background()
    
    // Initialize repository (once)
    resticlib.InitRepository(ctx, cfg)
    
    // Create client
    client, _ := resticlib.NewClient(ctx, cfg)
    defer client.Close()
    
    // Backup
    client.Backup(ctx, resticlib.BackupOptions{
        Paths: []string{"/data"},
        Tags:  []string{"important"},
    })
    
    // Restore
    client.Restore(ctx, resticlib.RestoreOptions{
        SnapshotID: "latest",
        Target:     "/restore",
    })
}
```

## Architecture

The library is built on top of restic's internal packages:
- `internal/backend/s3` - S3 backend implementation
- `internal/repository` - Repository management
- `internal/archiver` - Backup operations
- `internal/restorer` - Restore operations
- `internal/data` - Snapshot data structures

## Testing

The library includes:
1. Unit tests for configuration validation
2. Integration tests (requires S3 credentials)
3. Comprehensive test suite in the example application

Run tests:
```bash
# Unit tests
cd pkg/resticlib
go test -v

# Integration tests (with credentials)
export RESTIC_S3_ENDPOINT="..."
go test -v

# Example test suite
cd examples
./resticlib-example test
```

## Compatibility

- Works with any S3-compatible storage (AWS S3, Wasabi, MinIO, etc.)
- Uses restic's stable repository format
- Compatible with repositories created by restic CLI
- Go 1.24.0+

## Security

- All data is encrypted using AES-256
- Repository password is required for all operations
- S3 credentials are not stored, only used at runtime
- Supports HTTPS for secure transmission

## Future Enhancements

Possible additions (not implemented):
- Prune operations
- Repository statistics
- Snapshot filtering by tags/time
- Progress callbacks
- Parallel backup/restore
- Snapshot comparison/diff

## Files Created

```
pkg/resticlib/
├── client.go           # Main library implementation
├── client_test.go      # Tests
├── doc.go              # Package documentation
├── README.md           # Library documentation
├── USAGE.md            # Usage guide
└── examples/
    └── main.go         # Example application and tester
```

## Building

```bash
# Build library
cd pkg/resticlib
go build

# Build example
cd examples
go build -o resticlib-example
```

## Notes

- This is a NEW library added to restic
- NO existing restic code has been modified
- All files are in the `pkg/` directory
- The library uses restic's internal packages
- Fully documented with examples and tests
