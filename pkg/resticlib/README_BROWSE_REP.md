# Browsing Repository Contents

The `browse` command lets you explore the files stored inside any snapshot of a restic repository — without performing a full restore.

---

## Quick Reference

```
resticlib-example browse                            # list all snapshots
resticlib-example browse <snapshot-id>:<path>       # list contents of <path>
resticlib-example browse <snapshot-id>:<path> --depth <n>
```

---

## Setup

Build the example binary and export your repository credentials:

```bash
cd pkg/resticlib/examples
go build -o resticlib-example

export RESTIC_S3_ENDPOINT="s3.wasabisys.com"
export RESTIC_S3_ACCESS_KEY="your-access-key"
export RESTIC_S3_SECRET_KEY="your-secret-key"
export RESTIC_S3_BUCKET="your-bucket"
export RESTIC_S3_REGION="us-east-1"
export RESTIC_PASSWORD="your-password"
```

---

## Step 1 — Discover Snapshots

Run `browse` with no arguments to list every snapshot in the repository:

```
$ ./resticlib-example browse

Found 3 snapshot(s):

1. Snapshot a623f3ea
   Time: 2026-04-29T18:52:43-04:00
   Hostname: my-server
   Paths: [/home/alice/documents]
   Tags: [daily automated]

2. Snapshot b8d1c47f
   Time: 2026-04-28T02:00:01-04:00
   Hostname: my-server
   Paths: [/home/alice/documents /etc]

3. Snapshot c91e2a03
   Time: 2026-04-27T02:00:01-04:00
   Hostname: backup-host
   Paths: [/var/www]
```

Note the short snapshot IDs (`a623f3ea`, `b8d1c47f`, …) — these are used in the next steps.

---

## Step 2 — Browse the Root of a Snapshot

Append `:<path>` to the snapshot ID to browse its contents. Use `/` to start at the snapshot root:

```
$ ./resticlib-example browse a623f3ea:/

Browsing snapshot a623f3ea at / (depth=1)...

d drwxr-xr-x         0  2026-04-29 18:52:43  /home
```

The default `--depth 1` shows only the immediate children of the given path.
Increase the depth to see further levels.

---

## Step 3 — Navigate to a Sub-directory

Provide a specific path to jump straight into a directory deep in the snapshot tree.
**Depth is always counted from the path you specify**, not from the snapshot root.

```
$ ./resticlib-example browse a623f3ea:/home/alice/documents --depth 1

Browsing snapshot a623f3ea at /home/alice/documents (depth=1)...

d drwxr-xr-x         0  2026-04-29 18:52:43  projects
d drwxr-xr-x         0  2026-04-29 18:52:43  invoices
- -rw-r--r--      4096  2026-04-29 18:52:43  notes.txt
```

Paths in the output are **relative to the given path**. `projects` means
`/home/alice/documents/projects` in the snapshot.

---

## Step 4 — Control Recursion Depth

| `--depth` | Meaning |
|-----------|---------|
| `1` (default) | Immediate children only |
| `2` | Children and grandchildren |
| `N` | N levels below the given path |
| `0` | Unlimited — the full subtree |

### Examples

Show one level (immediate children only):
```
$ ./resticlib-example browse a623f3ea:/home/alice/documents --depth 1

- notes.txt
d projects
d invoices
```

Show two levels (children and their children):
```
$ ./resticlib-example browse a623f3ea:/home/alice/documents --depth 2

- notes.txt
d projects
- projects/report-q1.pdf
- projects/budget.xlsx
d invoices
d invoices/2025
d invoices/2026
```

Show the full subtree (unlimited depth):
```
$ ./resticlib-example browse a623f3ea:/home/alice/documents --depth 0

- notes.txt
d projects
- projects/report-q1.pdf
- projects/budget.xlsx
d invoices
d invoices/2025
- invoices/2025/inv-001.pdf
d invoices/2026
- invoices/2026/inv-042.pdf
```

---

## Output Format

Each line follows the pattern:

```
<type>  <permissions>   <size>   <mod-time>   <relative-path>
```

| Column | Values |
|--------|--------|
| `type` | `-` file · `d` directory · `l` symlink |
| `permissions` | Unix permission bits |
| `size` | File size in bytes (0 for directories) |
| `mod-time` | Last modification time (`YYYY-MM-DD HH:MM:SS`) |
| `relative-path` | Path relative to the given starting path |

---

## Using the Go API

The same functionality is available directly via the `resticlib` package:

```go
client, err := resticlib.NewClient(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// List all snapshots first
snapshots, err := client.ListSnapshots(ctx)
if err != nil {
    log.Fatal(err)
}
for _, snap := range snapshots {
    fmt.Printf("Snapshot %s  %s\n", snap.ID().Str(), snap.Time.Format(time.RFC3339))
}

// Browse a specific path inside a snapshot
entries, err := client.BrowseSnapshot(ctx, resticlib.BrowseOptions{
    SnapshotID: "a623f3ea",        // short or full ID, or "latest"
    Path:       "/home/alice/documents",
    Depth:      2,                 // 0 = unlimited
})
if err != nil {
    log.Fatal(err)
}

for _, e := range entries {
    fmt.Printf("%s  %s  %d bytes\n", e.Type, e.Path, e.Size)
}
```

### `BrowseOptions` fields

| Field | Type | Description |
|-------|------|-------------|
| `SnapshotID` | `string` | Full ID, short prefix, or `"latest"` |
| `Path` | `string` | Absolute path inside the snapshot (default `"/"`) |
| `Depth` | `int` | Recursion levels below `Path`; `0` = unlimited |

### `FileEntry` fields

| Field | Type | Description |
|-------|------|-------------|
| `Path` | `string` | Absolute path inside the snapshot |
| `Type` | `string` | `"file"`, `"dir"`, or `"symlink"` |
| `Size` | `uint64` | File size in bytes |
| `ModTime` | `time.Time` | Last modification time |
| `Mode` | `os.FileMode` | Unix permission bits |

---

## Typical Workflows

### Find a specific file before restoring

```bash
# Check what's in the documents directory of yesterday's snapshot
./resticlib-example browse b8d1c47f:/home/alice/documents --depth 0 | grep report

# Restore only after confirming the file is there
./resticlib-example restore /tmp/restore b8d1c47f
```

### Compare snapshots across days

```bash
./resticlib-example browse a623f3ea:/etc --depth 1
./resticlib-example browse b8d1c47f:/etc --depth 1
```

### Explore a large backup incrementally

```bash
# Start at the root
./resticlib-example browse a623f3ea:/

# Drill down into a subdirectory
./resticlib-example browse a623f3ea:/var/www --depth 2

# Go deeper
./resticlib-example browse a623f3ea:/var/www/html --depth 0
```
