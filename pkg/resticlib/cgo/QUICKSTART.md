# Restic C++ Integration - Quick Start

This guide will help you quickly get started with the C++ wrapper for restic.

## Quick Build

```bash
cd pkg/resticlib/cgo/cpp_tester
make
```

## Quick Test

```bash
# Set credentials
export RESTIC_S3_ENDPOINT="s3.wasabisys.com"
export RESTIC_S3_ACCESS_KEY="your-key"
export RESTIC_S3_SECRET_KEY="your-secret"
export RESTIC_S3_BUCKET="your-bucket"
export RESTIC_PASSWORD="your-password"

# Run all tests
./restic_cpp_tester all
```

## Minimal C++ Example

```cpp
#include "resticlib.h"
#include <iostream>
#include <cstring>

int main() {
    // Configure
    ResticConfig cfg;
    cfg.endpoint = strdup("s3.wasabisys.com");
    cfg.access_key_id = strdup("your-key");
    cfg.secret_access_key = strdup("your-secret");
    cfg.bucket_name = strdup("your-bucket");
    cfg.region = strdup("us-east-1");
    cfg.password = strdup("your-password");
    cfg.prefix = strdup("backups");
    cfg.connections = 5;
    cfg.use_http = 0;
    
    // Initialize repository (once)
    ResticError* err = ResticInitRepository(&cfg);
    if (err) {
        std::cout << "Note: " << err->error_message << std::endl;
        ResticFreeError(err);
    }
    
    // Create client
    int client = ResticNewClient(&cfg);
    if (client < 0) {
        std::cerr << "Failed to create client" << std::endl;
        return 1;
    }
    
    // Backup
    const char* paths[] = {"/tmp/test"};
    const char* tags[] = {"test"};
    char* snapshotID = ResticBackup(client, (char**)paths, 1, 
                                    (char**)tags, 1, (char*)"test-host");
    std::cout << "Backup: " << snapshotID << std::endl;
    ResticFreeString(snapshotID);
    
    // Restore
    err = ResticRestore(client, (char*)"latest", (char*)"/tmp/restore");
    if (err) {
        std::cerr << "Restore failed: " << err->error_message << std::endl;
        ResticFreeError(err);
    }
    
    // List snapshots
    int count = 0;
    ResticSnapshot** snapshots = ResticListSnapshots(client, &count);
    std::cout << "Found " << count << " snapshots" << std::endl;
    ResticFreeSnapshots(snapshots, count);
    
    // Cleanup
    ResticCloseClient(client);
    free(cfg.endpoint);
    free(cfg.access_key_id);
    free(cfg.secret_access_key);
    free(cfg.bucket_name);
    free(cfg.region);
    free(cfg.password);
    free(cfg.prefix);
    
    return 0;
}
```

## Build Your Application

```bash
g++ -o myapp myapp.cpp \
    -I/path/to/restic/pkg/resticlib/cgo \
    -L/path/to/restic/pkg/resticlib/cgo \
    -lrestic \
    -Wl,-rpath,/path/to/restic/pkg/resticlib/cgo \
    -std=c++11
```

## Using the C++ Wrapper Class

```cpp
#include "resticlib.h"
#include <iostream>
#include <vector>

// Copy the ResticClient class from main.cpp

int main() {
    try {
        ResticConfig cfg = getConfigFromEnv();
        ResticClient client(cfg);
        
        // Backup
        std::vector<std::string> paths = {"/data"};
        std::vector<std::string> tags = {"daily"};
        std::string snapID = client.backup(paths, tags, "my-host");
        std::cout << "Backup: " << snapID << std::endl;
        
        // Restore
        client.restore("latest", "/restore");
        
        // List
        auto snapshots = client.listSnapshots();
        std::cout << "Snapshots: " << snapshots.size() << std::endl;
        client.freeSnapshotList(snapshots);
        
        freeConfig(cfg);
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    
    return 0;
}
```

## Directory Structure

```
pkg/resticlib/cgo/
├── wrapper.go           # Go CGO wrapper
├── resticlib.h          # C/C++ header
├── librestic.so         # Compiled shared library
├── librestic.h          # Auto-generated CGO header
├── README.md            # Full documentation
└── cpp_tester/
    ├── main.cpp         # C++ test application
    ├── Makefile         # Build system
    └── restic_cpp_tester # Compiled tester
```

## Common Operations

### Initialize Once
```bash
./restic_cpp_tester init
```

### Backup Directory
```bash
./restic_cpp_tester test
```

### List All Snapshots
```bash
./restic_cpp_tester list
```

### Run Full Test Suite
```bash
./restic_cpp_tester all
```

## Environment Variables

Required:
- `RESTIC_S3_ENDPOINT` - S3 endpoint (e.g., "s3.wasabisys.com")
- `RESTIC_S3_ACCESS_KEY` - Access key ID
- `RESTIC_S3_SECRET_KEY` - Secret access key
- `RESTIC_S3_BUCKET` - Bucket name
- `RESTIC_PASSWORD` - Repository password

Optional:
- `RESTIC_S3_REGION` - AWS region (default: "us-east-1")

## Troubleshooting

### "Library not found" Error

```bash
export LD_LIBRARY_PATH=/path/to/restic/pkg/resticlib/cgo:$LD_LIBRARY_PATH
```

Or rebuild with rpath:
```bash
make clean && make
```

### Build Errors

Ensure you have:
1. Go 1.24.0+
2. GCC/G++ compiler
3. restic dependencies installed

```bash
cd /path/to/restic
go mod download
```

## Next Steps

1. Read `README.md` for complete API reference
2. Check `main.cpp` for detailed examples
3. Build your own application using the examples above

## Support

- Full documentation: `pkg/resticlib/cgo/README.md`
- Go library docs: `pkg/resticlib/README.md`
- Example code: `pkg/resticlib/cgo/cpp_tester/main.cpp`
