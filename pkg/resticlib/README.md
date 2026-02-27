# Resticlib - Restic Library for Go

A Go library that exposes the major functions of restic for programmatic backup and restore operations to S3-compatible storage (including AWS S3 and Wasabi).

## Features

- **Easy-to-use API**: Simple, intuitive interface for backup and restore operations
- **S3/Wasabi Support**: Full support for S3-compatible storage backends
- **Repository Management**: Initialize, check, and manage restic repositories
- **Snapshot Operations**: Create, list, and restore snapshots
- **Comprehensive Testing**: Includes full test suite and examples

## Installation

```bash
go get github.com/restic/restic/pkg/resticlib
```

## Quick Start

### Initialize a Repository

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
        AccessKeyID:     "your-access-key",
        SecretAccessKey: "your-secret-key",
        BucketName:      "your-bucket",
        Region:          "us-east-1",
        Password:        "your-secure-password",
    }
    
    ctx := context.Background()
    
    if err := resticlib.InitRepository(ctx, cfg); err != nil {
        log.Fatalf("Failed to initialize repository: %v", err)
    }
}
```

### Create a Backup

```go
client, err := resticlib.NewClient(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
    Paths:    []string{"/path/to/backup"},
    Tags:     []string{"important", "daily"},
    Hostname: "my-server",
})
if err != nil {
    log.Fatal(err)
}

log.Printf("Backup completed: %s", snapshot.ID.Str())
```

### Restore from Backup

```go
err = client.Restore(ctx, resticlib.RestoreOptions{
    SnapshotID: "latest", // or specific snapshot ID
    Target:     "/path/to/restore",
})
if err != nil {
    log.Fatal(err)
}
```

### List Snapshots

```go
snapshots, err := client.ListSnapshots(ctx)
if err != nil {
    log.Fatal(err)
}

for _, snap := range snapshots {
    log.Printf("Snapshot: %s, Time: %s, Paths: %v", 
        snap.ID.Str(), snap.Time, snap.Paths)
}
```

## Configuration

The `Config` struct supports the following fields:

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `Endpoint` | string | S3 endpoint (e.g., "s3.wasabisys.com") | Yes |
| `AccessKeyID` | string | S3 access key ID | Yes |
| `SecretAccessKey` | string | S3 secret access key | Yes |
| `BucketName` | string | S3 bucket name | Yes |
| `Region` | string | S3 region (e.g., "us-east-1") | Yes |
| `Password` | string | Repository encryption password | Yes |
| `Prefix` | string | Optional prefix for objects in bucket | No |
| `UseHTTP` | bool | Use HTTP instead of HTTPS | No |
| `Connections` | int | Number of concurrent connections (default: 5) | No |
| `CACert` | string | Path to CA certificate file | No |

## API Reference

### Repository Operations

#### `InitRepository(ctx context.Context, cfg *Config) error`

Initializes a new restic repository with the given configuration.

#### `NewClient(ctx context.Context, cfg *Config) (*Client, error)`

Creates a new client to interact with an existing repository.

#### `(*Client) Close() error`

Closes the client and releases resources.

### Backup Operations

#### `(*Client) Backup(ctx context.Context, opts BackupOptions) (*restic.Snapshot, error)`

Creates a new backup snapshot.

**BackupOptions:**
- `Paths []string` - Paths to backup
- `Hostname string` - Hostname for the snapshot
- `Tags []string` - Tags for the snapshot
- `Excludes []string` - Patterns to exclude
- `DryRun bool` - Don't actually upload data
- `TimeStamp *time.Time` - Custom timestamp for the snapshot

### Restore Operations

#### `(*Client) Restore(ctx context.Context, opts RestoreOptions) error`

Restores data from a snapshot.

**RestoreOptions:**
- `SnapshotID string` - Snapshot ID to restore (or "latest")
- `Target string` - Target directory for restore
- `Includes []string` - Patterns to include
- `Excludes []string` - Patterns to exclude
- `DryRun bool` - Don't actually write files
- `Verify bool` - Verify restored files

### Snapshot Operations

#### `(*Client) ListSnapshots(ctx context.Context) ([]*restic.Snapshot, error)`

Returns all snapshots in the repository.

#### `(*Client) Forget(ctx context.Context, opts ForgetOptions) error`

Removes a snapshot from the repository.

**ForgetOptions:**
- `SnapshotID string` - Snapshot ID to forget
- `Prune bool` - Run prune after forgetting

### Repository Maintenance

#### `(*Client) Check(ctx context.Context, checkUnused bool, output io.Writer) error`

Verifies the repository integrity.

## Examples

The `examples` directory contains a comprehensive example application that demonstrates all features:

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
./resticlib-example backup /path/to/backup

# List snapshots
./resticlib-example list

# Restore latest snapshot
./resticlib-example restore /path/to/restore

# Run comprehensive tests
./resticlib-example test
```

## Testing

The library includes comprehensive tests:

```bash
cd pkg/resticlib/examples
go run main.go test
```

This will:
1. Initialize a repository
2. Create test data
3. Perform a backup
4. List snapshots
5. Restore the backup
6. Verify data integrity
7. Check repository health

## Using with Wasabi

Wasabi is an S3-compatible cloud storage service. Configuration example:

```go
cfg := &resticlib.Config{
    Endpoint:        "s3.wasabisys.com",  // or regional endpoint
    AccessKeyID:     "your-wasabi-access-key",
    SecretAccessKey: "your-wasabi-secret-key",
    BucketName:      "your-bucket",
    Region:          "us-east-1",  // or other Wasabi region
    Password:        "your-repository-password",
}
```

Available Wasabi regions:
- `us-east-1` - US East (N. Virginia)
- `us-east-2` - US East (N. Virginia)
- `us-central-1` - US Central (Texas)
- `us-west-1` - US West (Oregon)
- `eu-central-1` - EU Central (Amsterdam)
- `eu-west-1` - EU West (London)
- `eu-west-2` - EU West (Paris)
- `ap-northeast-1` - Asia Pacific (Tokyo)
- `ap-northeast-2` - Asia Pacific (Osaka)
- `ap-southeast-1` - Asia Pacific (Singapore)
- `ap-southeast-2` - Asia Pacific (Sydney)

## Using with AWS S3

```go
cfg := &resticlib.Config{
    Endpoint:        "s3.amazonaws.com",
    AccessKeyID:     "your-aws-access-key",
    SecretAccessKey: "your-aws-secret-key",
    BucketName:      "your-bucket",
    Region:          "us-east-1",  // your AWS region
    Password:        "your-repository-password",
}
```

## Error Handling

All functions return detailed errors. Always check for errors:

```go
client, err := resticlib.NewClient(ctx, cfg)
if err != nil {
    // Handle error - could be:
    // - Invalid configuration
    // - Network issues
    // - Authentication problems
    // - Repository not initialized
    log.Fatalf("Failed to create client: %v", err)
}
defer client.Close()
```

## Performance Considerations

- **Concurrent Connections**: Adjust `Config.Connections` based on your network and S3 service
- **Large Files**: The library handles large files efficiently through chunking
- **Network Issues**: Restic includes automatic retry logic for transient failures

## Security

- **Passwords**: Repository passwords are used for encryption. Use strong, unique passwords
- **Credentials**: Store S3 credentials securely (e.g., environment variables, secrets manager)
- **Data Encryption**: All data is encrypted before upload using AES-256

## License

This library is part of the restic project and follows the same BSD 2-Clause License.

## Contributing

Contributions are welcome! Please ensure that:
- Code follows Go best practices
- New features include tests
- Documentation is updated
- No modifications to existing restic code (this is an add-on library)

## Support

- [Restic Documentation](https://restic.readthedocs.io/)
- [Restic Forum](https://forum.restic.net/)
- [GitHub Issues](https://github.com/restic/restic/issues)
