package resticlib_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/restic/restic/pkg/resticlib"
)

// This is an integration test that requires S3 credentials to be set
// Skip this test if credentials are not available
func TestIntegration(t *testing.T) {
	// Check if S3 credentials are available
	endpoint := os.Getenv("RESTIC_S3_ENDPOINT")
	accessKey := os.Getenv("RESTIC_S3_ACCESS_KEY")
	secretKey := os.Getenv("RESTIC_S3_SECRET_KEY")
	bucket := os.Getenv("RESTIC_S3_BUCKET")
	password := os.Getenv("RESTIC_PASSWORD")

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" || password == "" {
		t.Skip("Skipping integration test: S3 credentials not set")
	}

	ctx := context.Background()

	cfg := &resticlib.Config{
		Endpoint:        endpoint,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		BucketName:      bucket,
		Region:          "us-east-1",
		Password:        password,
		Prefix:          "resticlib-test",
		Connections:     5,
	}

	// Test 1: Initialize repository (may already exist, so we ignore errors)
	t.Log("Initializing repository...")
	_ = resticlib.InitRepository(ctx, cfg)

	// Test 2: Create client
	t.Log("Creating client...")
	client, err := resticlib.NewClient(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test 3: Create test data
	t.Log("Creating test data...")
	testDir, err := createTestData(t)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Test 4: Backup
	t.Log("Creating backup...")
	snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
		Paths:    []string{testDir},
		Tags:     []string{"test"},
		Hostname: "test-host",
	})
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}
	snapID := "unknown"
	if snapshot.ID() != nil {
		snapID = snapshot.ID().Str()
	}
	t.Logf("Backup created: %s", snapID)

	// Test 5: List snapshots
	t.Log("Listing snapshots...")
	snapshots, err := client.ListSnapshots(ctx)
	if err != nil {
		t.Fatalf("Failed to list snapshots: %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatal("No snapshots found")
	}
	t.Logf("Found %d snapshot(s)", len(snapshots))

	// Test 6: Restore
	t.Log("Restoring backup...")
	restoreDir := filepath.Join(os.TempDir(), "resticlib-test-restore")
	defer os.RemoveAll(restoreDir)

	err = client.Restore(ctx, resticlib.RestoreOptions{
		SnapshotID: snapID,
		Target:     restoreDir,
	})
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Test 7: Verify restored files exist
	t.Log("Verifying restored files...")
	if err := verifyFiles(testDir, restoreDir); err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	t.Log("All tests passed!")
}

func TestConfigValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		cfg     *resticlib.Config
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "missing endpoint",
			cfg: &resticlib.Config{
				BucketName:      "bucket",
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
				Password:        "password",
			},
			wantErr: true,
		},
		{
			name: "missing bucket",
			cfg: &resticlib.Config{
				Endpoint:        "s3.example.com",
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
				Password:        "password",
			},
			wantErr: true,
		},
		{
			name: "missing access key",
			cfg: &resticlib.Config{
				Endpoint:        "s3.example.com",
				BucketName:      "bucket",
				SecretAccessKey: "secret",
				Password:        "password",
			},
			wantErr: true,
		},
		{
			name: "missing secret key",
			cfg: &resticlib.Config{
				Endpoint:    "s3.example.com",
				BucketName:  "bucket",
				AccessKeyID: "key",
				Password:    "password",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			cfg: &resticlib.Config{
				Endpoint:        "s3.example.com",
				BucketName:      "bucket",
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resticlib.InitRepository(ctx, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func createTestData(t *testing.T) (string, error) {
	dir, err := os.MkdirTemp("", "resticlib-test-*")
	if err != nil {
		return "", err
	}

	files := map[string]string{
		"file1.txt":        "Test content 1",
		"file2.txt":        "Test content 2",
		"subdir/file3.txt": "Test content 3",
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

func verifyFiles(originalDir, restoreDir string) error {
	return filepath.Walk(originalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		relPath, err := filepath.Rel(originalDir, path)
		if err != nil {
			return err
		}

		restoredPath := filepath.Join(restoreDir, relPath)
		if _, err := os.Stat(restoredPath); err != nil {
			return err
		}

		return nil
	})
}
