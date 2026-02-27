.PHONY: all clean lib cpp_tester test

# Build all
all: lib cpp_tester

# Build the Go shared library
lib:
	@echo "Building Go shared library..."
	#cd ../.. && go build -buildmode=c-shared -o cgo/librestic.so ./cgo/wrapper.go
	cd ../.. && go build -buildmode=c-shared -gcflags="all=-N -l" -o cgo/librestic.so ./cgo/wrapper.go
	@echo "✓ Library built: librestic.so"

# Build the C++ tester
cpp_tester: lib
	@echo "Building C++ tester..."
	g++ -g -O0 -o restic_cpp_tester main.cpp -I.. -L.. -lrestic -Wl,-rpath,'$$ORIGIN/..' -std=c++11
	@echo "✓ C++ tester built: restic_cpp_tester"

# Run tests (requires environment variables to be set)
test: cpp_tester
	@echo "Running C++ tests..."
	./restic_cpp_tester all

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f ../librestic.so ../librestic.h restic_cpp_tester
	@echo "✓ Cleaned"

help:
	@echo "Makefile for Restic C++ Tester"
	@echo ""
	@echo "Targets:"
	@echo "  all         - Build library and C++ tester (default)"
	@echo "  lib         - Build Go shared library only"
	@echo "  cpp_tester  - Build C++ tester only"
	@echo "  test        - Run all tests"
	@echo "  clean       - Remove build artifacts"
	@echo "  help        - Show this help"
	@echo ""
	@echo "Environment variables (required for testing):"
	@echo "  RESTIC_S3_ENDPOINT"
	@echo "  RESTIC_S3_ACCESS_KEY"
	@echo "  RESTIC_S3_SECRET_KEY"
	@echo "  RESTIC_S3_BUCKET"
	@echo "  RESTIC_S3_REGION (default: us-east-1)"
	@echo "  RESTIC_PASSWORD"
	@echo ""
	@echo "Example:"
	@echo "  export RESTIC_S3_ENDPOINT=s3.wasabisys.com"
	@echo "  export RESTIC_S3_ACCESS_KEY=your-key"
	@echo "  export RESTIC_S3_SECRET_KEY=your-secret"
	@echo "  export RESTIC_S3_BUCKET=your-bucket"
	@echo "  export RESTIC_PASSWORD=your-password"
	@echo "  make test"
