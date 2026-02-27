#include <iostream>
#include <cstdlib>
#include <cstring>
#include <vector>
#include <string>
#include <fstream>
#include <sys/stat.h>
#include <unistd.h>
#include "../resticlib.h"

class ResticClient {
private:
    int clientID;
    
public:
    ResticClient(const ResticConfig& cfg) {
        clientID = ResticNewClient(const_cast<ResticConfig*>(&cfg));
        if (clientID < 0) {
            throw std::runtime_error("Failed to create restic client");
        }
    }
    
    ~ResticClient() {
        if (clientID >= 0) {
            ResticCloseClient(clientID);
        }
    }
    
    std::string backup(const std::vector<std::string>& paths, 
                      const std::vector<std::string>& tags = {},
                      const std::string& hostname = "") {
        // Convert paths to C array
        std::vector<char*> cPaths;
        for (const auto& path : paths) {
            cPaths.push_back(const_cast<char*>(path.c_str()));
        }
        
        // Convert tags to C array
        std::vector<char*> cTags;
        for (const auto& tag : tags) {
            cTags.push_back(const_cast<char*>(tag.c_str()));
        }
        
        char* hostnamePtr = hostname.empty() ? nullptr : const_cast<char*>(hostname.c_str());
        
        char* result = ResticBackup(clientID, 
                                    cPaths.data(), cPaths.size(),
                                    cTags.empty() ? nullptr : cTags.data(), cTags.size(),
                                    hostnamePtr);
        
        std::string snapshotID(result);
        ResticFreeString(result);
        
        if (snapshotID.find("ERROR:") == 0) {
            throw std::runtime_error(snapshotID.substr(7));
        }
        
        return snapshotID;
    }
    
    void restore(const std::string& snapshotID, const std::string& target) {
        ResticError* err = ResticRestore(clientID, 
                                         const_cast<char*>(snapshotID.c_str()),
                                         const_cast<char*>(target.c_str()));
        if (err != nullptr) {
            std::string errMsg(err->error_message);
            ResticFreeError(err);
            throw std::runtime_error("Restore failed: " + errMsg);
        }
    }
    
    std::vector<ResticSnapshot*> listSnapshots() {
        int count = 0;
        ResticSnapshot** snapshots = ResticListSnapshots(clientID, &count);
        
        std::vector<ResticSnapshot*> result;
        if (snapshots != nullptr && count > 0) {
            for (int i = 0; i < count; i++) {
                result.push_back(snapshots[i]);
            }
            // Free the original array (allocated by Go), but keep the snapshot structs
            free(snapshots);
        }
        
        return result;
    }
    
    void freeSnapshotList(const std::vector<ResticSnapshot*>& snapshots) {
        if (!snapshots.empty()) {
            ResticFreeSnapshots(const_cast<ResticSnapshot**>(snapshots.data()), 
                              snapshots.size());
        }
    }
};

ResticConfig getConfigFromEnv() {
    ResticConfig cfg;
    
    const char* endpoint = std::getenv("RESTIC_S3_ENDPOINT");
    const char* accessKey = std::getenv("RESTIC_S3_ACCESS_KEY");
    const char* secretKey = std::getenv("RESTIC_S3_SECRET_KEY");
    const char* bucket = std::getenv("RESTIC_S3_BUCKET");
    const char* region = std::getenv("RESTIC_S3_REGION");
    const char* password = std::getenv("RESTIC_PASSWORD");
    
    if (!endpoint || !accessKey || !secretKey || !bucket || !password) {
        throw std::runtime_error("Missing required environment variables");
    }
    
    cfg.endpoint = strdup(endpoint);
    cfg.access_key_id = strdup(accessKey);
    cfg.secret_access_key = strdup(secretKey);
    cfg.bucket_name = strdup(bucket);
    cfg.region = region ? strdup(region) : strdup("us-east-1");
    cfg.password = strdup(password);
    cfg.prefix = strdup("restic-cpp-test");
    cfg.connections = 5;
    cfg.use_http = 0;
    
    return cfg;
}

void freeConfig(ResticConfig& cfg) {
    free(cfg.endpoint);
    free(cfg.access_key_id);
    free(cfg.secret_access_key);
    free(cfg.bucket_name);
    free(cfg.region);
    free(cfg.password);
    free(cfg.prefix);
}

bool createTestData(const std::string& path) {
    // Create directory
    mkdir(path.c_str(), 0755);
    
    // Create test files
    std::vector<std::string> files = {
        path + "/file1.txt",
        path + "/file2.txt",
        path + "/file3.txt"
    };
    
    for (size_t i = 0; i < files.size(); i++) {
        std::ofstream file(files[i]);
        if (!file) return false;
        file << "Test content for file " << (i + 1) << std::endl;
        file << "This is a test file created by the C++ tester." << std::endl;
        file.close();
    }
    
    // Create subdirectory
    std::string subdir = path + "/subdir";
    mkdir(subdir.c_str(), 0755);
    
    std::ofstream subfile(subdir + "/file4.txt");
    if (!subfile) return false;
    subfile << "Test file in subdirectory" << std::endl;
    subfile.close();
    
    return true;
}

void removeDirectory(const std::string& path) {
    std::string cmd = "rm -rf " + path;
    system(cmd.c_str());
}

bool verifyRestore(const std::string& originalPath, const std::string& restorePath) {
    // Check if file exists - restic restores with full original path
    std::string testFile = restorePath + originalPath + "/file1.txt";
    std::ifstream file(testFile);
    if (!file) {
        std::cerr << "Failed to find restored file: " << testFile << std::endl;
        return false;
    }
    
    std::string line;
    std::getline(file, line);
    file.close();
    
    return line.find("Test content") != std::string::npos;
}

void testInit() {
    std::cout << "Test 1: Initialize repository" << std::endl;
    
    ResticConfig cfg = getConfigFromEnv();
    
    ResticError* err = ResticInitRepository(&cfg);
    if (err != nullptr) {
        std::cout << "  ⚠ Repository may already exist: " << err->error_message << std::endl;
        ResticFreeError(err);
    } else {
        std::cout << "  ✓ Repository initialized" << std::endl;
    }
    
    freeConfig(cfg);
}

void testBackupRestore() {
    std::cout << "\nTest 2: Backup and Restore" << std::endl;
    
    ResticConfig cfg = getConfigFromEnv();
    
    try {
        // Create client
        std::cout << "  Creating client..." << std::endl;
        ResticClient client(cfg);
        std::cout << "  ✓ Client created" << std::endl;
        
        // Create test data
        std::string testDir = "/tmp/restic-cpp-test-data";
        std::cout << "  Creating test data..." << std::endl;
        removeDirectory(testDir);
        if (!createTestData(testDir)) {
            throw std::runtime_error("Failed to create test data");
        }
        std::cout << "  ✓ Test data created in " << testDir << std::endl;
        
        // Perform backup
        std::cout << "  Performing backup..." << std::endl;
        std::vector<std::string> paths = {testDir};
        std::vector<std::string> tags = {"cpp-test", "automated"};
        std::string snapshotID = client.backup(paths, tags, "cpp-tester");
        std::cout << "  ✓ Backup completed: " << snapshotID << std::endl;
        
        // Restore backup
        std::string restoreDir = "/tmp/restic-cpp-test-restore";
        removeDirectory(restoreDir);
        mkdir(restoreDir.c_str(), 0755);
        
        std::cout << "  Restoring backup..." << std::endl;
        client.restore(snapshotID, restoreDir);
        std::cout << "  ✓ Restore completed to " << restoreDir << std::endl;
        
        // Verify restored data
        std::cout << "  Verifying restored data..." << std::endl;
        if (verifyRestore(testDir, restoreDir)) {
            std::cout << "  ✓ Data verified successfully" << std::endl;
        } else {
            std::cout << "  ✗ Data verification failed" << std::endl;
        }
        
        // Clean up
        removeDirectory(testDir);
        removeDirectory(restoreDir);
        
    } catch (const std::exception& e) {
        std::cerr << "  ✗ Error: " << e.what() << std::endl;
        freeConfig(cfg);
        throw;
    }
    
    freeConfig(cfg);
}

void testListSnapshots() {
    std::cout << "\nTest 3: List snapshots" << std::endl;
    
    ResticConfig cfg = getConfigFromEnv();
    
    try {
        ResticClient client(cfg);
        
        auto snapshots = client.listSnapshots();
        
        if (snapshots.empty()) {
            std::cout << "  No snapshots found" << std::endl;
        } else {
            std::cout << "  Found " << snapshots.size() << " snapshot(s):" << std::endl;
            
            for (size_t i = 0; i < snapshots.size(); i++) {
                const auto& snap = snapshots[i];
                std::cout << "\n  " << (i + 1) << ". Snapshot " << snap->snapshot_id << std::endl;
                std::cout << "     Time: " << snap->time << std::endl;
                std::cout << "     Hostname: " << snap->hostname << std::endl;
                
                if (snap->path_count > 0) {
                    std::cout << "     Paths: ";
                    for (int j = 0; j < snap->path_count; j++) {
                        if (j > 0) std::cout << ", ";
                        std::cout << snap->paths[j];
                    }
                    std::cout << std::endl;
                }
                
                if (snap->tag_count > 0) {
                    std::cout << "     Tags: ";
                    for (int j = 0; j < snap->tag_count; j++) {
                        if (j > 0) std::cout << ", ";
                        std::cout << snap->tags[j];
                    }
                    std::cout << std::endl;
                }
            }
            
            client.freeSnapshotList(snapshots);
        }
        
    } catch (const std::exception& e) {
        std::cerr << "  ✗ Error: " << e.what() << std::endl;
        freeConfig(cfg);
        throw;
    }
    
    freeConfig(cfg);
}

void printUsage() {
    std::cout << "Usage: restic_cpp_tester <command>" << std::endl;
    std::cout << std::endl;
    std::cout << "Commands:" << std::endl;
    std::cout << "  init     - Initialize repository" << std::endl;
    std::cout << "  test     - Run backup and restore test" << std::endl;
    std::cout << "  list     - List snapshots" << std::endl;
    std::cout << "  all      - Run all tests" << std::endl;
    std::cout << std::endl;
    std::cout << "Environment variables:" << std::endl;
    std::cout << "  RESTIC_S3_ENDPOINT       - S3/Wasabi endpoint" << std::endl;
    std::cout << "  RESTIC_S3_ACCESS_KEY     - S3 access key ID" << std::endl;
    std::cout << "  RESTIC_S3_SECRET_KEY     - S3 secret access key" << std::endl;
    std::cout << "  RESTIC_S3_BUCKET         - S3 bucket name" << std::endl;
    std::cout << "  RESTIC_S3_REGION         - S3 region (default: us-east-1)" << std::endl;
    std::cout << "  RESTIC_PASSWORD          - Repository password" << std::endl;
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        printUsage();
        return 1;
    }
    
    std::string command = argv[1];
    
    try {
        if (command == "init") {
            testInit();
        } else if (command == "test") {
            testBackupRestore();
        } else if (command == "list") {
            testListSnapshots();
        } else if (command == "all") {
            testInit();
            testBackupRestore();
            testListSnapshots();
            std::cout << "\n✓ All tests completed successfully!" << std::endl;
        } else {
            std::cerr << "Unknown command: " << command << std::endl;
            printUsage();
            return 1;
        }
    } catch (const std::exception& e) {
        std::cerr << "\n✗ Test failed: " << e.what() << std::endl;
        return 1;
    }
    
    return 0;
}
