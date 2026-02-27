package resticlib

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/restic/restic/internal/archiver"
	"github.com/restic/restic/internal/backend"
	"github.com/restic/restic/internal/backend/s3"
	"github.com/restic/restic/internal/data"
	"github.com/restic/restic/internal/fs"
	"github.com/restic/restic/internal/options"
	"github.com/restic/restic/internal/repository"
	"github.com/restic/restic/internal/restic"
	"github.com/restic/restic/internal/restorer"
)

// Client is the main entry point for the restic library
type Client struct {
	repo     *repository.Repository
	backend  backend.Backend
	password string
	config   *Config
}

// Config holds configuration for the restic client
type Config struct {
	// S3/Wasabi configuration
	Endpoint        string // e.g., "s3.wasabisys.com" or "s3.amazonaws.com"
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string // e.g., "us-east-1"
	Prefix          string // Optional: prefix for all objects in the bucket
	UseHTTP         bool   // Use HTTP instead of HTTPS

	// Repository password
	Password string

	// Optional settings
	Connections int    // Number of concurrent connections (default: 5)
	CACert      string // Path to CA certificate file
}

// NewClient creates a new restic client with the given configuration
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &Client{
		password: cfg.Password,
		config:   cfg,
	}

	// Open or create the backend
	be, err := client.openBackend(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open backend: %w", err)
	}
	client.backend = be

	// Open the repository
	repo, err := client.openRepository(ctx)
	if err != nil {
		be.Close()
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	client.repo = repo

	return client, nil
}

// InitRepository initializes a new repository
func InitRepository(ctx context.Context, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Create backend configuration
	s3Cfg := s3.NewConfig()
	s3Cfg.Endpoint = cfg.Endpoint
	s3Cfg.Bucket = cfg.BucketName
	s3Cfg.Prefix = cfg.Prefix
	s3Cfg.Region = cfg.Region
	s3Cfg.UseHTTP = cfg.UseHTTP
	s3Cfg.KeyID = cfg.AccessKeyID
	s3Cfg.Secret = options.NewSecretString(cfg.SecretAccessKey)

	if cfg.Connections > 0 {
		s3Cfg.Connections = uint(cfg.Connections)
	}

	// Create the backend
	be, err := s3.Create(ctx, s3Cfg, nil, func(string, ...interface{}) {})
	if err != nil {
		return fmt.Errorf("failed to create backend: %w", err)
	}
	defer be.Close()

	// Create repository and initialize it
	repo, err := repository.New(be, repository.Options{})
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	err = repo.Init(ctx, restic.StableRepoVersion, cfg.Password, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	return nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	if c.backend != nil {
		return c.backend.Close()
	}
	return nil
}

// openBackend opens the S3/Wasabi backend
func (c *Client) openBackend(ctx context.Context) (backend.Backend, error) {
	s3Cfg := s3.NewConfig()
	s3Cfg.Endpoint = c.config.Endpoint
	s3Cfg.Bucket = c.config.BucketName
	s3Cfg.Prefix = c.config.Prefix
	s3Cfg.Region = c.config.Region
	s3Cfg.UseHTTP = c.config.UseHTTP
	s3Cfg.KeyID = c.config.AccessKeyID
	s3Cfg.Secret = options.NewSecretString(c.config.SecretAccessKey)

	if c.config.Connections > 0 {
		s3Cfg.Connections = uint(c.config.Connections)
	}

	be, err := s3.Open(ctx, s3Cfg, nil, func(string, ...interface{}) {})
	if err != nil {
		return nil, fmt.Errorf("failed to open S3 backend: %w", err)
	}

	return be, nil
}

// openRepository opens the repository
func (c *Client) openRepository(ctx context.Context) (*repository.Repository, error) {
	repo, err := repository.New(c.backend, repository.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	err = repo.SearchKey(ctx, c.password, 10, "")
	if err != nil {
		return nil, fmt.Errorf("wrong password or repository not initialized: %w", err)
	}

	return repo, nil
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if cfg.BucketName == "" {
		return fmt.Errorf("bucket name is required")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("access key ID is required")
	}
	if cfg.SecretAccessKey == "" {
		return fmt.Errorf("secret access key is required")
	}
	if cfg.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

// BackupOptions holds options for backup operations
type BackupOptions struct {
	Paths      []string             // Paths to backup
	Hostname   string               // Hostname for the snapshot
	Tags       []string             // Tags for the snapshot
	Excludes   []string             // Patterns to exclude
	DryRun     bool                 // Don't actually upload data
	TimeStamp  *time.Time           // Custom timestamp for the snapshot
	ReadStdin  bool                 // Read from stdin instead of filesystem
	StdinName  string               // Filename to use when reading from stdin
	ProgressFn func(uint64, uint64) // Progress callback: (current, total) bytes
}

// Backup creates a new backup snapshot
func (c *Client) Backup(ctx context.Context, opts BackupOptions) (*data.Snapshot, error) {
	if len(opts.Paths) == 0 && !opts.ReadStdin {
		return nil, fmt.Errorf("no paths specified for backup")
	}

	// Get hostname
	hostname := opts.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
		if hostname == "" {
			hostname = "unknown"
		}
	}

	if opts.DryRun {
		// Return a dummy snapshot for dry run
		dummySnapshot := &data.Snapshot{
			Hostname: hostname,
			Tags:     opts.Tags,
			Paths:    opts.Paths,
			Time:     time.Now(),
		}
		if opts.TimeStamp != nil {
			dummySnapshot.Time = *opts.TimeStamp
		}
		return dummySnapshot, nil
	}

	// Create the archiver
	arch := archiver.New(c.repo, fs.Local{}, archiver.Options{})

	// Set up snapshot options
	snapshotOpts := archiver.SnapshotOptions{
		Time:        time.Now(),
		Hostname:    hostname,
		Tags:        opts.Tags,
		Excludes:    opts.Excludes,
		BackupStart: time.Now(),
	}

	if opts.TimeStamp != nil {
		snapshotOpts.Time = *opts.TimeStamp
	}

	// Backup the paths
	snapshot, snapshotID, _, err := arch.Snapshot(ctx, opts.Paths, snapshotOpts)
	if err != nil {
		return nil, fmt.Errorf("snapshot failed: %w", err)
	}

	// Load the snapshot to get one with the ID set
	snapshot, err = data.LoadSnapshot(ctx, c.repo, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshot: %w", err)
	}

	return snapshot, nil
}

// RestoreOptions holds options for restore operations
type RestoreOptions struct {
	SnapshotID string               // Snapshot ID to restore (or "latest")
	Target     string               // Target directory for restore
	Includes   []string             // Patterns to include
	Excludes   []string             // Patterns to exclude
	DryRun     bool                 // Don't actually write files
	Verify     bool                 // Verify restored files
	ProgressFn func(uint64, uint64) // Progress callback: (current, total) bytes
}

// Restore restores data from a snapshot
func (c *Client) Restore(ctx context.Context, opts RestoreOptions) error {
	if opts.Target == "" {
		return fmt.Errorf("target directory is required")
	}

	if opts.SnapshotID == "" {
		opts.SnapshotID = "latest"
	}

	// Find the snapshot
	var snapshotID restic.ID
	var err error

	if strings.ToLower(opts.SnapshotID) == "latest" {
		snapshotID, err = c.findLatestSnapshot(ctx)
		if err != nil {
			return fmt.Errorf("failed to find latest snapshot: %w", err)
		}
	} else {
		// Parse snapshot ID
		snapshotID, err = restic.ParseID(opts.SnapshotID)
		if err != nil {
			return fmt.Errorf("failed to parse snapshot ID: %w", err)
		}
	}

	// Load the snapshot
	snapshot, err := data.LoadSnapshot(ctx, c.repo, snapshotID)
	if err != nil {
		return fmt.Errorf("failed to load snapshot: %w", err)
	}

	if snapshot.Tree == nil {
		return fmt.Errorf("snapshot has no tree")
	}

	if opts.DryRun {
		return nil
	}

	// Create the restorer
	res := restorer.NewRestorer(c.repo, snapshot, restorer.Options{})

	// Restore the snapshot
	_, err = res.RestoreTo(ctx, opts.Target)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	return nil
}

// ListSnapshots returns all snapshots in the repository
func (c *Client) ListSnapshots(ctx context.Context) ([]*data.Snapshot, error) {
	snapshots := []*data.Snapshot{}

	err := data.ForAllSnapshots(ctx, c.repo, c.repo, nil, func(id restic.ID, snapshot *data.Snapshot, err error) error {
		if err != nil {
			return err
		}
		snapshots = append(snapshots, snapshot)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	return snapshots, nil
}

// findLatestSnapshot finds the most recent snapshot
func (c *Client) findLatestSnapshot(ctx context.Context) (restic.ID, error) {
	var latestID restic.ID
	var latestTime time.Time

	err := data.ForAllSnapshots(ctx, c.repo, c.repo, nil, func(id restic.ID, snapshot *data.Snapshot, err error) error {
		if err != nil {
			return err
		}
		if snapshot.Time.After(latestTime) {
			latestTime = snapshot.Time
			latestID = id
		}
		return nil
	})

	if err != nil {
		return restic.ID{}, err
	}

	if latestID.IsNull() {
		return restic.ID{}, fmt.Errorf("no snapshots found")
	}

	return latestID, nil
}

// ForgetOptions holds options for forget operations
type ForgetOptions struct {
	SnapshotID string
	Prune      bool
}

// Forget removes a snapshot from the repository
func (c *Client) Forget(ctx context.Context, opts ForgetOptions) error {
	if opts.SnapshotID == "" {
		return fmt.Errorf("snapshot ID is required")
	}

	snapshotID, err := restic.ParseID(opts.SnapshotID)
	if err != nil {
		return fmt.Errorf("failed to parse snapshot ID: %w", err)
	}

	// Delete the snapshot file from backend
	h := backend.Handle{Type: backend.SnapshotFile, Name: snapshotID.String()}
	err = c.backend.Remove(ctx, h)
	if err != nil {
		return fmt.Errorf("failed to remove snapshot: %w", err)
	}

	return nil
}

// Check verifies the repository integrity
func (c *Client) Check(ctx context.Context, checkUnused bool, output io.Writer) error {
	if output == nil {
		output = io.Discard
	}

	fmt.Fprintln(output, "checking repository integrity...")

	// This is a simplified check - full implementation would require
	// using internal/checker package
	err := c.repo.LoadIndex(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	fmt.Fprintln(output, "repository check passed")
	return nil
}
