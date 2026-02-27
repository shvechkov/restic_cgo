// Package resticlib provides a high-level Go library for restic backup operations.
//
// This library exposes the major functions of restic for programmatic use,
// with a focus on S3-compatible storage backends (AWS S3, Wasabi, etc.).
//
// # Features
//
//   - Easy-to-use API for backup and restore operations
//   - Full support for S3-compatible storage backends
//   - Repository initialization and management
//   - Snapshot operations (create, list, restore, forget)
//   - Repository integrity checking
//
// # Quick Start
//
// Initialize a repository:
//
//	cfg := &resticlib.Config{
//	    Endpoint:        "s3.wasabisys.com",
//	    AccessKeyID:     "your-access-key",
//	    SecretAccessKey: "your-secret-key",
//	    BucketName:      "your-bucket",
//	    Region:          "us-east-1",
//	    Password:        "your-secure-password",
//	}
//
//	ctx := context.Background()
//	err := resticlib.InitRepository(ctx, cfg)
//
// Create a backup:
//
//	client, err := resticlib.NewClient(ctx, cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
//	    Paths: []string{"/path/to/backup"},
//	    Tags:  []string{"important"},
//	})
//
// Restore from backup:
//
//	err = client.Restore(ctx, resticlib.RestoreOptions{
//	    SnapshotID: "latest",
//	    Target:     "/path/to/restore",
//	})
//
// # Configuration
//
// The Config struct requires the following fields:
//   - Endpoint: S3 endpoint (e.g., "s3.wasabisys.com")
//   - AccessKeyID: S3 access key ID
//   - SecretAccessKey: S3 secret access key
//   - BucketName: S3 bucket name
//   - Region: S3 region (e.g., "us-east-1")
//   - Password: Repository encryption password
//
// Optional fields include Prefix, UseHTTP, Connections, and CACert.
//
// # Examples
//
// See the examples directory for comprehensive usage examples and a test suite.
package resticlib
