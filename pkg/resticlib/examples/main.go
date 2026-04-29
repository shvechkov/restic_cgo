package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/restic/restic/pkg/resticlib"
)

// Example demonstrates how to use the resticlib package
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		if err := exampleInit(); err != nil {
			log.Fatalf("Init failed: %v", err)
		}
	case "backup":
		if err := exampleBackup(); err != nil {
			log.Fatalf("Backup failed: %v", err)
		}
	case "restore":
		if err := exampleRestore(); err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	case "list":
		if err := exampleList(); err != nil {
			log.Fatalf("List failed: %v", err)
		}
	case "test":
		if err := runTests(); err != nil {
			log.Fatalf("Tests failed: %v", err)
		}
	case "browse":
		if err := exampleBrowse(os.Args[2:]); err != nil {
			log.Fatalf("Browse failed: %v", err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: resticlib-example <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init     - Initialize a new repository")
	fmt.Println("  backup   - Create a backup")
	fmt.Println("  restore  - Restore from backup")
	fmt.Println("  list     - List snapshots")
	fmt.Println("  browse   - Browse snapshot contents")
	fmt.Println("            Without arguments: lists all snapshots")
	fmt.Println("            With <snapshot-id>:<path>: lists files at that path")
	fmt.Println("            --depth <num>: limit directory recursion depth (default: 1)")
	fmt.Println("  test     - Run comprehensive tests")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  RESTIC_S3_ENDPOINT       - S3/Wasabi endpoint (e.g., s3.wasabisys.com)")
	fmt.Println("  RESTIC_S3_ACCESS_KEY     - S3 access key ID")
	fmt.Println("  RESTIC_S3_SECRET_KEY     - S3 secret access key")
	fmt.Println("  RESTIC_S3_BUCKET         - S3 bucket name")
	fmt.Println("  RESTIC_S3_REGION         - S3 region (default: us-east-1)")
	fmt.Println("  RESTIC_PASSWORD          - Repository password")
}

// getConfig reads configuration from environment variables
func getConfig() *resticlib.Config {
	region := os.Getenv("RESTIC_S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	return &resticlib.Config{
		Endpoint:        os.Getenv("RESTIC_S3_ENDPOINT"),
		AccessKeyID:     os.Getenv("RESTIC_S3_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("RESTIC_S3_SECRET_KEY"),
		BucketName:      os.Getenv("RESTIC_S3_BUCKET"),
		Region:          region,
		Password:        os.Getenv("RESTIC_PASSWORD"),
		Prefix:          "restic-test", // Optional prefix
		Connections:     5,
	}
}

// exampleInit initializes a new repository
func exampleInit() error {
	fmt.Println("Initializing repository...")

	cfg := getConfig()
	ctx := context.Background()

	err := resticlib.InitRepository(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	fmt.Println("✓ Repository initialized successfully")
	return nil
}

// exampleBackup creates a backup
func exampleBackup() error {
	fmt.Println("Creating backup...")

	cfg := getConfig()
	ctx := context.Background()

	// Create client
	client, err := resticlib.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Determine what to backup
	paths := os.Args[2:]
	if len(paths) == 0 {
		// Default to current directory
		paths = []string{"."}
	}

	// Create backup
	snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
		Paths:    paths,
		Tags:     []string{"example", "automated"},
		Hostname: "example-host",
	})
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Printf("✓ Backup completed successfully\n")
	if snapshot.ID() != nil {
		fmt.Printf("  Snapshot ID: %s\n", snapshot.ID().Str())
	}
	fmt.Printf("  Time: %s\n", snapshot.Time.Format(time.RFC3339))
	if snapshot.Summary != nil {
		fmt.Printf("  Files: %d new, %d changed\n", snapshot.Summary.FilesNew, snapshot.Summary.FilesChanged)
		fmt.Printf("  Data added: %d bytes\n", snapshot.Summary.DataAdded)
	}

	return nil
}

// exampleRestore restores from a backup
func exampleRestore() error {
	fmt.Println("Restoring from backup...")

	cfg := getConfig()
	ctx := context.Background()

	// Create client
	client, err := resticlib.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Determine target directory
	target := "./restore-output"
	if len(os.Args) > 2 {
		target = os.Args[2]
	}

	// Determine snapshot ID
	snapshotID := "latest"
	if len(os.Args) > 3 {
		snapshotID = os.Args[3]
	}

	// Create target directory
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Restore
	err = client.Restore(ctx, resticlib.RestoreOptions{
		SnapshotID: snapshotID,
		Target:     target,
	})
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Printf("✓ Restore completed successfully to %s\n", target)
	return nil
}

// exampleBrowse browses snapshot contents.
// Without arguments it lists all snapshots.
// With <snapshot-id>:<path> it lists files at that path.
// --depth <num> controls recursion depth (default 1; 0 = unlimited).
func exampleBrowse(args []string) error {
	// Parse --depth flag from args.
	depth := 1
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--depth" {
			if i+1 >= len(args) {
				return fmt.Errorf("--depth requires a numeric argument")
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				return fmt.Errorf("invalid --depth value %q: %w", args[i], err)
			}
			depth = n
		} else {
			remaining = append(remaining, args[i])
		}
	}

	cfg := getConfig()
	ctx := context.Background()

	client, err := resticlib.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// No positional argument: list snapshots.
	if len(remaining) == 0 {
		snapshots, err := client.ListSnapshots(ctx)
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}
		if len(snapshots) == 0 {
			fmt.Println("No snapshots found")
			return nil
		}
		fmt.Printf("Found %d snapshot(s):\n\n", len(snapshots))
		for i, snap := range snapshots {
			snapID := "unknown"
			if snap.ID() != nil {
				snapID = snap.ID().Str()
			}
			fmt.Printf("%d. Snapshot %s\n", i+1, snapID)
			fmt.Printf("   Time: %s\n", snap.Time.Format(time.RFC3339))
			fmt.Printf("   Hostname: %s\n", snap.Hostname)
			fmt.Printf("   Paths: %v\n", snap.Paths)
			if len(snap.Tags) > 0 {
				fmt.Printf("   Tags: %v\n", snap.Tags)
			}
			fmt.Println()
		}
		return nil
	}

	// Positional argument: <snapshot-id>:<path>
	arg := remaining[0]
	colonIdx := strings.Index(arg, ":")
	var snapshotID, browsePath string
	if colonIdx < 0 {
		snapshotID = arg
		browsePath = "/"
	} else {
		snapshotID = arg[:colonIdx]
		browsePath = arg[colonIdx+1:]
		if browsePath == "" {
			browsePath = "/"
		}
	}

	fmt.Printf("Browsing snapshot %s at %s (depth=%d)...\n\n", snapshotID, browsePath, depth)

	entries, err := client.BrowseSnapshot(ctx, resticlib.BrowseOptions{
		SnapshotID: snapshotID,
		Path:       browsePath,
		Depth:      depth,
	})
	if err != nil {
		return fmt.Errorf("browse failed: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("(empty)")
		return nil
	}

	// Paths in entries are absolute. Strip the browsePath prefix so that
	// depth is visually anchored to the given path, not the filesystem root.
	pathPrefix := strings.TrimRight(browsePath, "/")

	for _, e := range entries {
		typeChar := '-'
		if e.Type == "dir" {
			typeChar = 'd'
		} else if e.Type == "symlink" {
			typeChar = 'l'
		}

		// Show paths relative to the starting path.
		displayPath := e.Path
		if pathPrefix != "" && strings.HasPrefix(e.Path, pathPrefix+"/") {
			displayPath = e.Path[len(pathPrefix)+1:]
		}

		fmt.Printf("%c %s  %8d  %s  %s\n",
			typeChar,
			e.Mode,
			e.Size,
			e.ModTime.Format("2006-01-02 15:04:05"),
			displayPath,
		)
	}

	return nil
}


func exampleList() error {
	fmt.Println("Listing snapshots...")

	cfg := getConfig()
	ctx := context.Background()

	// Create client
	client, err := resticlib.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// List snapshots
	snapshots, err := client.ListSnapshots(ctx)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		fmt.Println("No snapshots found")
		return nil
	}

	fmt.Printf("Found %d snapshot(s):\n\n", len(snapshots))
	for i, snap := range snapshots {
		snapID := "unknown"
		if snap.ID() != nil {
			snapID = snap.ID().Str()
		}
		fmt.Printf("%d. Snapshot %s\n", i+1, snapID)
		fmt.Printf("   Time: %s\n", snap.Time.Format(time.RFC3339))
		fmt.Printf("   Hostname: %s\n", snap.Hostname)
		fmt.Printf("   Paths: %v\n", snap.Paths)
		if len(snap.Tags) > 0 {
			fmt.Printf("   Tags: %v\n", snap.Tags)
		}
		fmt.Println()
	}

	return nil
}

// runTests runs comprehensive tests of the library
func runTests() error {
	fmt.Println("Running comprehensive tests...")
	fmt.Println()

	cfg := getConfig()
	ctx := context.Background()

	// Test 1: Initialize repository
	fmt.Println("Test 1: Initialize repository")
	err := resticlib.InitRepository(ctx, cfg)
	if err != nil {
		// It might already exist, try to continue
		fmt.Printf("  ⚠ Repository may already exist: %v\n", err)
	} else {
		fmt.Println("  ✓ Repository initialized")
	}

	// Test 2: Create client
	fmt.Println("\nTest 2: Create client")
	client, err := resticlib.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()
	fmt.Println("  ✓ Client created successfully")

	// Test 3: Create test data
	fmt.Println("\nTest 3: Create test data")
	testDir, err := createTestData()
	if err != nil {
		return fmt.Errorf("failed to create test data: %w", err)
	}
	defer os.RemoveAll(testDir)
	fmt.Printf("  ✓ Test data created in %s\n", testDir)

	// Test 4: Create backup
	fmt.Println("\nTest 4: Create backup")
	snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
		Paths:    []string{testDir},
		Tags:     []string{"test", "automated"},
		Hostname: "test-host",
	})
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	snapID := "unknown"
	if snapshot.ID() != nil {
		snapID = snapshot.ID().Str()
	}
	fmt.Printf("  ✓ Backup created: %s\n", snapID)

	// Test 5: List snapshots
	fmt.Println("\nTest 5: List snapshots")
	snapshots, err := client.ListSnapshots(ctx)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}
	fmt.Printf("  ✓ Found %d snapshot(s)\n", len(snapshots))

	// Test 6: Restore backup
	fmt.Println("\nTest 6: Restore backup")
	restoreDir := filepath.Join(os.TempDir(), "restic-restore-test")
	defer os.RemoveAll(restoreDir)

	err = client.Restore(ctx, resticlib.RestoreOptions{
		SnapshotID: snapID,
		Target:     restoreDir,
	})
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}
	fmt.Printf("  ✓ Restored to %s\n", restoreDir)

	// Test 7: Verify restored data
	fmt.Println("\nTest 7: Verify restored data")
	if err := verifyRestoreData(testDir, restoreDir); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}
	fmt.Println("  ✓ Data verified successfully")

	// Test 8: Check repository
	fmt.Println("\nTest 8: Check repository")
	if err := client.Check(ctx, false, os.Stdout); err != nil {
		return fmt.Errorf("check failed: %w", err)
	}
	fmt.Println("  ✓ Repository check passed")

	fmt.Println("\n✓ All tests passed successfully!")
	return nil
}

// createTestData creates test files for backup
func createTestData() (string, error) {
	dir, err := os.MkdirTemp("", "restic-test-*")
	if err != nil {
		return "", err
	}

	// Create some test files
	files := map[string]string{
		"file1.txt":                   "This is test file 1",
		"file2.txt":                   "This is test file 2",
		"subdir/file3.txt":            "This is test file 3 in a subdirectory",
		"subdir/file4.txt":            "This is test file 4 in a subdirectory",
		"subdir/nested/file5.txt":     "This is test file 5 in a nested subdirectory",
	}

	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return dir, err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return dir, err
		}
	}

	return dir, nil
}

// verifyRestoreData verifies that restored data matches original
func verifyRestoreData(originalDir, restoreDir string) error {
	// Walk through original directory
	return filepath.Walk(originalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(originalDir, path)
		if err != nil {
			return err
		}

		// Read original file
		originalContent, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read original file %s: %w", path, err)
		}

		// Read restored file
		restoredPath := filepath.Join(restoreDir, relPath)
		restoredContent, err := os.ReadFile(restoredPath)
		if err != nil {
			return fmt.Errorf("failed to read restored file %s: %w", restoredPath, err)
		}

		// Compare content
		if string(originalContent) != string(restoredContent) {
			return fmt.Errorf("content mismatch for file %s", relPath)
		}

		return nil
	})
}
