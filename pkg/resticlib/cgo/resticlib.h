#ifndef RESTICLIB_H
#define RESTICLIB_H

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    char* endpoint;
    char* access_key_id;
    char* secret_access_key;
    char* bucket_name;
    char* region;
    char* password;
    char* prefix;
    int connections;
    int use_http;
} ResticConfig;

typedef struct {
    char* snapshot_id;
    char* time;
    char* hostname;
    int path_count;
    char** paths;
    int tag_count;
    char** tags;
} ResticSnapshot;

typedef struct {
    char* error_message;
} ResticError;

// Initialize a new repository
ResticError* ResticInitRepository(ResticConfig* cfg);

// Create a new client (returns client ID or -1 on error)
int ResticNewClient(ResticConfig* cfg);

// Close client and free resources
void ResticCloseClient(int clientID);

// Create a backup (returns snapshot ID or error message starting with "ERROR:")
char* ResticBackup(int clientID, char** paths, int pathCount, char** tags, int tagCount, char* hostname);

// Restore from a snapshot
ResticError* ResticRestore(int clientID, char* snapshotID, char* target);

// List all snapshots (returns array of snapshots, count is set to number of snapshots)
ResticSnapshot** ResticListSnapshots(int clientID, int* count);

// Free snapshot array returned by ResticListSnapshots
void ResticFreeSnapshots(ResticSnapshot** snapshots, int count);

// Free error returned by ResticInitRepository or ResticRestore
void ResticFreeError(ResticError* err);

// Free string returned by ResticBackup
void ResticFreeString(char* str);

#ifdef __cplusplus
}
#endif

#endif // RESTICLIB_H
