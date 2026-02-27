#!/bin/bash
set -x

# Clean up
rm -rf /tmp/restic-cpp-test-data /tmp/restic-cpp-test-restore

# Create test data
mkdir -p /tmp/restic-cpp-test-data
echo "test content" > /tmp/restic-cpp-test-data/file1.txt

# Run the test but don't verify
./restic_cpp_tester test 2>&1 | grep -v "Data verification failed" || true

# Check what was restored
echo "=== Checking restore directory ==="
if [ -d /tmp/restic-cpp-test-restore ]; then
  find /tmp/restic-cpp-test-restore -type f
  echo "=== Content structure ==="
  ls -laR /tmp/restic-cpp-test-restore
else
  echo "Restore directory not found (already cleaned up)"
fi
