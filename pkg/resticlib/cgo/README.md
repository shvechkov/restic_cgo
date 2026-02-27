# Restic C++ Tester

This directory contains a C++ wrapper and tester for the restic Go library, enabling C++ applications to perform backup and restore operations to S3/Wasabi storage.

## Architecture

```
┌─────────────────────┐
│   C++ Application   │
│   (main.cpp)        │
└──────────┬──────────┘
           │
           │ Calls C API
           ▼
┌─────────────────────┐
│   C Wrapper (CGO)   │
│   (wrapper.go)      │
└──────────┬──────────┘
           │
           │ Uses
           ▼
┌─────────────────────┐
│   Go Library        │
│   (resticlib)       │
└─────────────────────┘
```

## Files

- **wrapper.go** - CGO wrapper that exposes C-compatible API
- **resticlib.h** - C/C++ header file
- **cpp_tester/main.cpp** - C++ test application
- **cpp_tester/Makefile** - Build system

## Building

### Prerequisites

- Go 1.24.0 or later
- GCC/G++ compiler
- Linux environment (tested on Linux)

### Build Steps

```bash
cd cpp_tester
make
```

This will:
1. Build the Go library as a shared object (`librestic.so`)
2. Compile the C++ tester application

## Usage

### Set Environment Variables

```bash
export RESTIC_S3_ENDPOINT="s3.wasabisys.com"
export RESTIC_S3_ACCESS_KEY="your-access-key"
export RESTIC_S3_SECRET_KEY="your-secret-key"
export RESTIC_S3_BUCKET="your-bucket"
export RESTIC_S3_REGION="us-east-1"
export RESTIC_PASSWORD="your-password"
```

### Run Tests

```bash
# Initialize repository
./restic_cpp_tester init

# Run backup and restore test
./restic_cpp_tester test

# List snapshots
./restic_cpp_tester list

# Run all tests
./restic_cpp_tester all

# Or use make
make test
```

## C++ API Reference

### Configuration

```cpp
ResticConfig cfg;
cfg.endpoint = "s3.wasabisys.com";
cfg.access_key_id = "your-key";
cfg.secret_access_key = "your-secret";
cfg.bucket_name = "your-bucket";
cfg.region = "us-east-1";
cfg.password = "your-password";
cfg.prefix = "backups";
cfg.connections = 5;
cfg.use_http = 0;  // 0 = HTTPS, 1 = HTTP
```

### Initialize Repository

```cpp
ResticError* err = ResticInitRepository(&cfg);
if (err != nullptr) {
    std::cerr << "Error: " << err->error_message << std::endl;
    ResticFreeError(err);
}
```

### Create Client

```cpp
int clientID = ResticNewClient(&cfg);
if (clientID < 0) {
    std::cerr << "Failed to create client" << std::endl;
}
```

### Backup

```cpp
// Prepare paths
const char* paths[] = {"/data", "/home"};
int pathCount = 2;

// Prepare tags
const char* tags[] = {"daily", "important"};
int tagCount = 2;

// Perform backup
char* snapshotID = ResticBackup(clientID, 
                                (char**)paths, pathCount,
                                (char**)tags, tagCount,
                                (char*)"my-hostname");

if (strstr(snapshotID, "ERROR:") == snapshotID) {
    std::cerr << "Backup failed: " << snapshotID << std::endl;
} else {
    std::cout << "Backup created: " << snapshotID << std::endl;
}

ResticFreeString(snapshotID);
```

### Restore

```cpp
ResticError* err = ResticRestore(clientID, 
                                 (char*)"snapshot-id-or-latest",
                                 (char*)"/restore/path");
if (err != nullptr) {
    std::cerr << "Restore failed: " << err->error_message << std::endl;
    ResticFreeError(err);
}
```

### List Snapshots

```cpp
int count = 0;
ResticSnapshot** snapshots = ResticListSnapshots(clientID, &count);

for (int i = 0; i < count; i++) {
    ResticSnapshot* snap = snapshots[i];
    std::cout << "Snapshot: " << snap->snapshot_id << std::endl;
    std::cout << "  Time: " << snap->time << std::endl;
    std::cout << "  Hostname: " << snap->hostname << std::endl;
    
    // List paths
    for (int j = 0; j < snap->path_count; j++) {
        std::cout << "  Path: " << snap->paths[j] << std::endl;
    }
    
    // List tags
    for (int j = 0; j < snap->tag_count; j++) {
        std::cout << "  Tag: " << snap->tags[j] << std::endl;
    }
}

ResticFreeSnapshots(snapshots, count);
```

### Close Client

```cpp
ResticCloseClient(clientID);
```

## C++ Wrapper Class

The tester includes a convenient C++ wrapper class:

```cpp
#include "resticlib.h"

class ResticClient {
public:
    ResticClient(const ResticConfig& cfg);
    ~ResticClient();
    
    std::string backup(const std::vector<std::string>& paths, 
                      const std::vector<std::string>& tags = {},
                      const std::string& hostname = "");
    
    void restore(const std::string& snapshotID, 
                const std::string& target);
    
    std::vector<ResticSnapshot*> listSnapshots();
    void freeSnapshotList(const std::vector<ResticSnapshot*>& snapshots);
};

// Usage
ResticConfig cfg = getConfigFromEnv();
ResticClient client(cfg);

// Backup
std::vector<std::string> paths = {"/data"};
std::vector<std::string> tags = {"daily"};
std::string snapshotID = client.backup(paths, tags, "hostname");

// Restore
client.restore("latest", "/restore/path");

// List
auto snapshots = client.listSnapshots();
// ... use snapshots ...
client.freeSnapshotList(snapshots);
```

## Memory Management

**Important**: The C API uses malloc/free for memory management. Always free resources:

- `ResticFreeError()` - Free error objects
- `ResticFreeString()` - Free strings returned by ResticBackup
- `ResticFreeSnapshots()` - Free snapshot arrays
- `ResticCloseClient()` - Close clients

The C++ wrapper class handles most memory management automatically.

## Testing

The tester performs comprehensive tests:

1. **Repository Initialization** - Creates a new repository
2. **Backup Test** - Creates test data and backs it up
3. **Restore Test** - Restores the backup to a new location
4. **Verification** - Verifies restored data matches original
5. **List Test** - Lists all snapshots in the repository

## Building Your Own Application

### 1. Link Against the Library

```bash
g++ -o myapp myapp.cpp -I/path/to/resticlib/cgo \
    -L/path/to/resticlib/cgo -lrestic \
    -Wl,-rpath,/path/to/resticlib/cgo
```

### 2. Include the Header

```cpp
#include "resticlib.h"
```

### 3. Use the API

See examples above or check `main.cpp` for complete examples.

## Troubleshooting

### Library Not Found

If you get "error while loading shared libraries":

```bash
export LD_LIBRARY_PATH=/path/to/resticlib/cgo:$LD_LIBRARY_PATH
```

Or use `-Wl,-rpath` when compiling (already included in Makefile).

### Build Errors

Make sure you have:
- Go 1.24.0+ installed
- GCC/G++ compiler
- All Go dependencies (`go mod download` in restic root)

### Runtime Errors

Check environment variables are set correctly:
```bash
./restic_cpp_tester init
```

## Platform Support

- **Linux**: Fully supported and tested
- **macOS**: Should work with minor Makefile adjustments (.dylib instead of .so)
- **Windows**: Requires additional work (DLL building, different paths)

## Performance

The C++ wrapper adds minimal overhead:
- Function calls go through CGO (small overhead)
- Data conversion between C and Go types
- Most time is spent in actual backup/restore operations

## Examples

See `main.cpp` for complete working examples of:
- Repository initialization
- Creating backups with tags
- Restoring snapshots
- Listing and displaying snapshots
- Error handling
- Memory management

## License

Same as restic (BSD 2-Clause License)
