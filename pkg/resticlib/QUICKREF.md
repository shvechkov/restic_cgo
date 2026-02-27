# Restic Library - Quick Reference

## Installation

```go
import "github.com/restic/restic/pkg/resticlib"
```

## Configuration

```go
cfg := &resticlib.Config{
    Endpoint:        "s3.wasabisys.com",        // Required
    AccessKeyID:     "your-access-key",         // Required
    SecretAccessKey: "your-secret-key",         // Required
    BucketName:      "your-bucket",             // Required
    Region:          "us-east-1",               // Required
    Password:        "repository-password",     // Required
    Prefix:          "backups/prod",            // Optional
    Connections:     5,                         // Optional (default: 5)
    UseHTTP:         false,                     // Optional (default: false)
}
```

## Initialize Repository (Once)

```go
ctx := context.Background()
err := resticlib.InitRepository(ctx, cfg)
```

## Create Client

```go
client, err := resticlib.NewClient(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## Backup

### Simple Backup
```go
snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
    Paths: []string{"/data"},
})
```

### Backup with Options
```go
snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
    Paths:    []string{"/data", "/home"},
    Tags:     []string{"production", "daily"},
    Hostname: "server-01",
    Excludes: []string{"*.tmp", "*.log"},
})
```

## Restore

### Restore Latest
```go
err := client.Restore(ctx, resticlib.RestoreOptions{
    SnapshotID: "latest",
    Target:     "/restore/path",
})
```

### Restore Specific Snapshot
```go
err := client.Restore(ctx, resticlib.RestoreOptions{
    SnapshotID: "a1b2c3d4",
    Target:     "/restore/path",
    Verify:     true,
})
```

## List Snapshots

```go
snapshots, err := client.ListSnapshots(ctx)
for _, snap := range snapshots {
    if snap.ID() != nil {
        fmt.Printf("ID: %s, Time: %s, Paths: %v\n", 
            snap.ID().Str(), snap.Time, snap.Paths)
    }
}
```

## Delete Snapshot

```go
err := client.Forget(ctx, resticlib.ForgetOptions{
    SnapshotID: "a1b2c3d4",
})
```

## Check Repository

```go
err := client.Check(ctx, false, os.Stdout)
```

## S3 Providers

### AWS S3
```go
cfg := &resticlib.Config{
    Endpoint:   "s3.amazonaws.com",
    Region:     "us-west-2",
    // ... rest of config
}
```

### Wasabi
```go
cfg := &resticlib.Config{
    Endpoint:   "s3.wasabisys.com",
    Region:     "us-east-1",
    // ... rest of config
}
```

### MinIO
```go
cfg := &resticlib.Config{
    Endpoint:   "minio.example.com:9000",
    UseHTTP:    true,  // If not using TLS
    Region:     "us-east-1",
    // ... rest of config
}
```

## Environment Variables

```bash
export RESTIC_S3_ENDPOINT="s3.wasabisys.com"
export RESTIC_S3_ACCESS_KEY="your-access-key"
export RESTIC_S3_SECRET_KEY="your-secret-key"
export RESTIC_S3_BUCKET="your-bucket"
export RESTIC_S3_REGION="us-east-1"
export RESTIC_PASSWORD="your-password"
```

## Complete Example

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/restic/restic/pkg/resticlib"
)

func main() {
    cfg := &resticlib.Config{
        Endpoint:        os.Getenv("RESTIC_S3_ENDPOINT"),
        AccessKeyID:     os.Getenv("RESTIC_S3_ACCESS_KEY"),
        SecretAccessKey: os.Getenv("RESTIC_S3_SECRET_KEY"),
        BucketName:      os.Getenv("RESTIC_S3_BUCKET"),
        Region:          os.Getenv("RESTIC_S3_REGION"),
        Password:        os.Getenv("RESTIC_PASSWORD"),
    }
    
    ctx := context.Background()
    
    // Initialize (once per repository)
    if err := resticlib.InitRepository(ctx, cfg); err != nil {
        log.Printf("Init failed (may already exist): %v", err)
    }
    
    // Create client
    client, err := resticlib.NewClient(ctx, cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Backup
    snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
        Paths: []string{"/data"},
        Tags:  []string{"automated"},
    })
    if err != nil {
        log.Fatal(err)
    }
    if snapshot.ID() != nil {
        log.Printf("Backup created: %s", snapshot.ID().Str())
    }
    
    // List
    snapshots, err := client.ListSnapshots(ctx)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Found %d snapshots", len(snapshots))
    
    // Restore
    err = client.Restore(ctx, resticlib.RestoreOptions{
        SnapshotID: "latest",
        Target:     "/restore",
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Error Handling

```go
client, err := resticlib.NewClient(ctx, cfg)
if err != nil {
    if strings.Contains(err.Error(), "wrong password") {
        // Handle password error
    } else if strings.Contains(err.Error(), "not initialized") {
        // Repository needs initialization
    } else {
        // Other error
    }
}
```

## Common Patterns

### Daily Backup Script
```go
snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
    Paths: []string{"/data"},
    Tags:  []string{"daily", time.Now().Format("2006-01-02")},
})
```

### Backup with Retry
```go
var snapshot *data.Snapshot
var err error
for i := 0; i < 3; i++ {
    snapshot, err = client.Backup(ctx, opts)
    if err == nil {
        break
    }
    time.Sleep(time.Minute)
}
```

### List Recent Snapshots
```go
snapshots, _ := client.ListSnapshots(ctx)
cutoff := time.Now().AddDate(0, 0, -7)  // Last 7 days
for _, snap := range snapshots {
    if snap.Time.After(cutoff) {
        // Process recent snapshot
    }
}
```

## Documentation

- **README.md** - Complete documentation
- **USAGE.md** - Detailed usage guide
- **SUMMARY.md** - Project overview
- **examples/main.go** - Working examples

## Testing

```bash
# Run unit tests
cd pkg/resticlib
go test -v

# Run example test suite
cd examples
./resticlib-example test
```

## Getting Help

- Check documentation in `pkg/resticlib/`
- Run example: `./resticlib-example test`
- See [restic documentation](https://restic.readthedocs.io/)
