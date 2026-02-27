# Resticlib Usage Guide

This guide provides detailed information on using the resticlib package for backup and restore operations.

## Table of Contents

1. [Installation](#installation)
2. [Configuration](#configuration)
3. [Repository Initialization](#repository-initialization)
4. [Backup Operations](#backup-operations)
5. [Restore Operations](#restore-operations)
6. [Snapshot Management](#snapshot-management)
7. [Error Handling](#error-handling)
8. [Best Practices](#best-practices)
9. [Advanced Usage](#advanced-usage)

## Installation

```bash
go get github.com/restic/restic/pkg/resticlib
```

## Configuration

### Basic Configuration

Create a configuration for S3/Wasabi storage:

```go
cfg := &resticlib.Config{
    Endpoint:        "s3.wasabisys.com",
    AccessKeyID:     "your-access-key-id",
    SecretAccessKey: "your-secret-access-key",
    BucketName:      "your-bucket-name",
    Region:          "us-east-1",
    Password:        "your-repository-password",
}
```

### Configuration from Environment Variables

```go
import "os"

func getConfigFromEnv() *resticlib.Config {
    return &resticlib.Config{
        Endpoint:        os.Getenv("RESTIC_S3_ENDPOINT"),
        AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
        SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
        BucketName:      os.Getenv("RESTIC_S3_BUCKET"),
        Region:          os.Getenv("AWS_REGION"),
        Password:        os.Getenv("RESTIC_PASSWORD"),
    }
}
```

### Advanced Configuration

```go
cfg := &resticlib.Config{
    Endpoint:        "s3.amazonaws.com",
    AccessKeyID:     "your-access-key",
    SecretAccessKey: "your-secret-key",
    BucketName:      "your-bucket",
    Region:          "us-west-2",
    Password:        "your-password",
    
    // Optional: Prefix for all objects in the bucket
    Prefix:          "backups/production",
    
    // Optional: Number of concurrent connections (default: 5)
    Connections:     10,
    
    // Optional: Use HTTP instead of HTTPS
    UseHTTP:         false,
    
    // Optional: Path to CA certificate
    CACert:          "/path/to/ca-cert.pem",
}
```

## Repository Initialization

Initialize a new repository before first use:

```go
import (
    "context"
    "log"
    
    "github.com/restic/restic/pkg/resticlib"
)

func initRepo() {
    cfg := &resticlib.Config{
        // ... configuration ...
    }
    
    ctx := context.Background()
    
    if err := resticlib.InitRepository(ctx, cfg); err != nil {
        log.Fatalf("Failed to initialize repository: %v", err)
    }
    
    log.Println("Repository initialized successfully")
}
```

**Note**: Only call `InitRepository` once per repository. Subsequent operations should use `NewClient`.

## Backup Operations

### Simple Backup

```go
func simpleBackup() error {
    cfg := getConfig()
    ctx := context.Background()
    
    client, err := resticlib.NewClient(ctx, cfg)
    if err != nil {
        return err
    }
    defer client.Close()
    
    snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
        Paths: []string{"/home/user/documents"},
    })
    if err != nil {
        return err
    }
    
    log.Printf("Backup completed: %s", snapshot.ID.Str())
    return nil
}
```

### Backup with Tags and Hostname

```go
snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
    Paths:    []string{"/var/www", "/etc"},
    Tags:     []string{"production", "daily", "web-server"},
    Hostname: "web-01.example.com",
})
```

### Backup Multiple Paths

```go
snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
    Paths: []string{
        "/home/user/documents",
        "/home/user/pictures",
        "/home/user/videos",
    },
    Tags: []string{"personal", "important"},
})
```

### Backup with Custom Timestamp

```go
import "time"

timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
    Paths:     []string{"/data"},
    TimeStamp: &timestamp,
})
```

### Dry Run Backup

Test a backup without actually uploading data:

```go
snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
    Paths:  []string{"/data"},
    DryRun: true,
})
```

## Restore Operations

### Restore Latest Snapshot

```go
err := client.Restore(ctx, resticlib.RestoreOptions{
    SnapshotID: "latest",
    Target:     "/restore/path",
})
```

### Restore Specific Snapshot

```go
err := client.Restore(ctx, resticlib.RestoreOptions{
    SnapshotID: "a1b2c3d4", // Or full snapshot ID
    Target:     "/restore/path",
})
```

### Restore with Includes/Excludes

```go
err := client.Restore(ctx, resticlib.RestoreOptions{
    SnapshotID: "latest",
    Target:     "/restore/path",
    Includes:   []string{"*.txt", "*.pdf"},
})
```

### Dry Run Restore

Test a restore without writing files:

```go
err := client.Restore(ctx, resticlib.RestoreOptions{
    SnapshotID: "latest",
    Target:     "/restore/path",
    DryRun:     true,
})
```

### Restore with Verification

```go
err := client.Restore(ctx, resticlib.RestoreOptions{
    SnapshotID: "latest",
    Target:     "/restore/path",
    Verify:     true, // Verify restored files
})
```

## Snapshot Management

### List All Snapshots

```go
snapshots, err := client.ListSnapshots(ctx)
if err != nil {
    return err
}

for _, snap := range snapshots {
    fmt.Printf("ID: %s\n", snap.ID.Str())
    fmt.Printf("Time: %s\n", snap.Time)
    fmt.Printf("Hostname: %s\n", snap.Hostname)
    fmt.Printf("Paths: %v\n", snap.Paths)
    fmt.Printf("Tags: %v\n", snap.Tags)
    fmt.Println()
}
```

### Delete a Snapshot

```go
err := client.Forget(ctx, resticlib.ForgetOptions{
    SnapshotID: "a1b2c3d4",
    Prune:      true, // Also run prune
})
```

### Check Repository Integrity

```go
import "os"

err := client.Check(ctx, false, os.Stdout)
if err != nil {
    log.Printf("Repository check failed: %v", err)
}
```

## Error Handling

### Comprehensive Error Handling

```go
func backupWithErrorHandling() error {
    cfg := getConfig()
    ctx := context.Background()
    
    client, err := resticlib.NewClient(ctx, cfg)
    if err != nil {
        return fmt.Errorf("failed to create client: %w", err)
    }
    defer func() {
        if cerr := client.Close(); cerr != nil {
            log.Printf("Warning: failed to close client: %v", cerr)
        }
    }()
    
    snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
        Paths: []string{"/data"},
    })
    if err != nil {
        return fmt.Errorf("backup failed: %w", err)
    }
    
    log.Printf("Backup successful: %s", snapshot.ID.Str())
    return nil
}
```

### Retry Logic

```go
import "time"

func backupWithRetry(maxRetries int) error {
    cfg := getConfig()
    ctx := context.Background()
    
    client, err := resticlib.NewClient(ctx, cfg)
    if err != nil {
        return err
    }
    defer client.Close()
    
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        _, err := client.Backup(ctx, resticlib.BackupOptions{
            Paths: []string{"/data"},
        })
        if err == nil {
            return nil
        }
        
        lastErr = err
        log.Printf("Backup attempt %d failed: %v", i+1, err)
        time.Sleep(time.Minute * time.Duration(i+1))
    }
    
    return fmt.Errorf("backup failed after %d attempts: %w", maxRetries, lastErr)
}
```

## Best Practices

### 1. Always Close the Client

```go
client, err := resticlib.NewClient(ctx, cfg)
if err != nil {
    return err
}
defer client.Close()
```

### 2. Use Context for Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
defer cancel()

client, err := resticlib.NewClient(ctx, cfg)
// ...
```

### 3. Tag Your Backups

```go
snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
    Paths: []string{"/data"},
    Tags:  []string{
        "environment:production",
        "application:myapp",
        "frequency:daily",
    },
})
```

### 4. Store Credentials Securely

Never hardcode credentials:

```go
// Good: Use environment variables
cfg := &resticlib.Config{
    AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
    SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
    Password:        os.Getenv("RESTIC_PASSWORD"),
}
```

### 5. Regular Repository Checks

```go
func regularCheck() error {
    client, err := resticlib.NewClient(ctx, cfg)
    if err != nil {
        return err
    }
    defer client.Close()
    
    return client.Check(ctx, false, os.Stdout)
}
```

## Advanced Usage

### Automated Backup Script

```go
package main

import (
    "context"
    "log"
    "os"
    "time"
    
    "github.com/restic/restic/pkg/resticlib"
)

func main() {
    cfg := getConfig()
    ctx := context.Background()
    
    client, err := resticlib.NewClient(ctx, cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Create backup
    snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
        Paths:    []string{"/data"},
        Tags:     []string{"automated", time.Now().Format("2006-01-02")},
        Hostname: getHostname(),
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Backup completed: %s", snapshot.ID.Str())
    
    // Clean up old snapshots (keep last 7 days)
    snapshots, _ := client.ListSnapshots(ctx)
    cutoff := time.Now().AddDate(0, 0, -7)
    
    for _, snap := range snapshots {
        if snap.Time.Before(cutoff) {
            log.Printf("Removing old snapshot: %s", snap.ID.Str())
            client.Forget(ctx, resticlib.ForgetOptions{
                SnapshotID: snap.ID.Str(),
            })
        }
    }
}

func getHostname() string {
    hostname, _ := os.Hostname()
    if hostname == "" {
        hostname = "unknown"
    }
    return hostname
}
```

### Concurrent Backups

```go
import "golang.org/x/sync/errgroup"

func backupMultipleSources() error {
    cfg := getConfig()
    ctx := context.Background()
    
    sources := []string{"/home", "/var", "/etc"}
    
    g, ctx := errgroup.WithContext(ctx)
    
    for _, source := range sources {
        source := source // Capture variable
        g.Go(func() error {
            client, err := resticlib.NewClient(ctx, cfg)
            if err != nil {
                return err
            }
            defer client.Close()
            
            _, err = client.Backup(ctx, resticlib.BackupOptions{
                Paths: []string{source},
                Tags:  []string{"concurrent"},
            })
            return err
        })
    }
    
    return g.Wait()
}
```

## Troubleshooting

### Common Issues

**Issue**: "repository not found" error
- **Solution**: Initialize the repository first with `InitRepository`

**Issue**: "wrong password" error
- **Solution**: Verify the password matches the one used during initialization

**Issue**: Network timeout errors
- **Solution**: Increase context timeout or check network connectivity

**Issue**: S3 authentication errors
- **Solution**: Verify access keys, secret keys, and bucket permissions

### Enable Debug Logging

Set the `RESTIC_DEBUG` environment variable:

```bash
export RESTIC_DEBUG=1
```

## See Also

- [README.md](README.md) - Main documentation
- [examples/main.go](examples/main.go) - Complete examples
- [Restic Documentation](https://restic.readthedocs.io/)
