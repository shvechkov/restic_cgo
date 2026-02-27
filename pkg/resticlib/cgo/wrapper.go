package main

/*
#include <stdlib.h>
#include <string.h>

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
*/
import "C"
import (
	"context"
	"time"
	"unsafe"

	"github.com/restic/restic/pkg/resticlib"
)

var (
	clients = make(map[int]*resticlib.Client)
	nextID  = 1
)

//export ResticInitRepository
func ResticInitRepository(cfg *C.ResticConfig) *C.ResticError {
	goConfig := cConfigToGo(cfg)
	
	ctx := context.Background()
	err := resticlib.InitRepository(ctx, goConfig)
	
	if err != nil {
		return makeError(err.Error())
	}
	
	return nil
}

//export ResticNewClient
func ResticNewClient(cfg *C.ResticConfig) C.int {
	goConfig := cConfigToGo(cfg)
	
	ctx := context.Background()
	client, err := resticlib.NewClient(ctx, goConfig)
	
	if err != nil {
		return -1
	}
	
	clientID := nextID
	nextID++
	clients[clientID] = client
	
	return C.int(clientID)
}

//export ResticCloseClient
func ResticCloseClient(clientID C.int) {
	if client, ok := clients[int(clientID)]; ok {
		client.Close()
		delete(clients, int(clientID))
	}
}

//export ResticBackup
func ResticBackup(clientID C.int, paths **C.char, pathCount C.int, tags **C.char, tagCount C.int, hostname *C.char) *C.char {
	client, ok := clients[int(clientID)]
	if !ok {
		return C.CString("ERROR: invalid client ID")
	}
	
	// Convert paths
	goPaths := make([]string, int(pathCount))
	pathSlice := unsafe.Slice(paths, pathCount)
	for i, path := range pathSlice {
		goPaths[i] = C.GoString(path)
	}
	
	// Convert tags
	goTags := make([]string, int(tagCount))
	if tagCount > 0 {
		tagSlice := unsafe.Slice(tags, tagCount)
		for i, tag := range tagSlice {
			goTags[i] = C.GoString(tag)
		}
	}
	
	goHostname := ""
	if hostname != nil {
		goHostname = C.GoString(hostname)
	}
	
	ctx := context.Background()
	snapshot, err := client.Backup(ctx, resticlib.BackupOptions{
		Paths:    goPaths,
		Tags:     goTags,
		Hostname: goHostname,
	})
	
	if err != nil {
		return C.CString("ERROR: " + err.Error())
	}
	
	if snapshot.ID() != nil {
		return C.CString(snapshot.ID().String())
	}
	
	return C.CString("unknown")
}

//export ResticRestore
func ResticRestore(clientID C.int, snapshotID *C.char, target *C.char) *C.ResticError {
	client, ok := clients[int(clientID)]
	if !ok {
		return makeError("invalid client ID")
	}
	
	goSnapshotID := C.GoString(snapshotID)
	goTarget := C.GoString(target)
	
	ctx := context.Background()
	err := client.Restore(ctx, resticlib.RestoreOptions{
		SnapshotID: goSnapshotID,
		Target:     goTarget,
	})
	
	if err != nil {
		return makeError(err.Error())
	}
	
	return nil
}

//export ResticListSnapshots
func ResticListSnapshots(clientID C.int, count *C.int) **C.ResticSnapshot {
	client, ok := clients[int(clientID)]
	if !ok {
		*count = 0
		return nil
	}
	
	ctx := context.Background()
	snapshots, err := client.ListSnapshots(ctx)
	
	if err != nil || len(snapshots) == 0 {
		*count = 0
		return nil
	}
	
	*count = C.int(len(snapshots))
	
	// Allocate array of snapshot pointers
	cSnapshots := C.malloc(C.size_t(len(snapshots)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	cSnapshotsSlice := unsafe.Slice((**C.ResticSnapshot)(cSnapshots), len(snapshots))
	
	for i, snap := range snapshots {
		cSnap := (*C.ResticSnapshot)(C.malloc(C.sizeof_ResticSnapshot))
		
		// Set snapshot ID
		if snap.ID() != nil {
			cSnap.snapshot_id = C.CString(snap.ID().Str())
		} else {
			cSnap.snapshot_id = C.CString("unknown")
		}
		
		// Set time
		cSnap.time = C.CString(snap.Time.Format(time.RFC3339))
		
		// Set hostname
		cSnap.hostname = C.CString(snap.Hostname)
		
		// Set paths
		cSnap.path_count = C.int(len(snap.Paths))
		if len(snap.Paths) > 0 {
			pathsArray := C.malloc(C.size_t(len(snap.Paths)) * C.size_t(unsafe.Sizeof(uintptr(0))))
			pathsSlice := unsafe.Slice((**C.char)(pathsArray), len(snap.Paths))
			for j, path := range snap.Paths {
				pathsSlice[j] = C.CString(path)
			}
			cSnap.paths = (**C.char)(pathsArray)
		}
		
		// Set tags
		cSnap.tag_count = C.int(len(snap.Tags))
		if len(snap.Tags) > 0 {
			tagsArray := C.malloc(C.size_t(len(snap.Tags)) * C.size_t(unsafe.Sizeof(uintptr(0))))
			tagsSlice := unsafe.Slice((**C.char)(tagsArray), len(snap.Tags))
			for j, tag := range snap.Tags {
				tagsSlice[j] = C.CString(tag)
			}
			cSnap.tags = (**C.char)(tagsArray)
		}
		
		cSnapshotsSlice[i] = cSnap
	}
	
	return (**C.ResticSnapshot)(cSnapshots)
}

//export ResticFreeSnapshots
func ResticFreeSnapshots(snapshots **C.ResticSnapshot, count C.int) {
	if snapshots == nil || count == 0 {
		return
	}
	
	snapshotsSlice := unsafe.Slice(snapshots, count)
	for _, snap := range snapshotsSlice {
		if snap != nil {
			C.free(unsafe.Pointer(snap.snapshot_id))
			C.free(unsafe.Pointer(snap.time))
			C.free(unsafe.Pointer(snap.hostname))
			
			if snap.path_count > 0 && snap.paths != nil {
				pathsSlice := unsafe.Slice(snap.paths, snap.path_count)
				for _, path := range pathsSlice {
					C.free(unsafe.Pointer(path))
				}
				C.free(unsafe.Pointer(snap.paths))
			}
			
			if snap.tag_count > 0 && snap.tags != nil {
				tagsSlice := unsafe.Slice(snap.tags, snap.tag_count)
				for _, tag := range tagsSlice {
					C.free(unsafe.Pointer(tag))
				}
				C.free(unsafe.Pointer(snap.tags))
			}
			
			C.free(unsafe.Pointer(snap))
		}
	}
	
	// Note: We don't free the snapshots array itself here because
	// it may be owned by the caller (e.g., C++ vector's internal storage)
}

//export ResticFreeError
func ResticFreeError(err *C.ResticError) {
	if err != nil && err.error_message != nil {
		C.free(unsafe.Pointer(err.error_message))
		C.free(unsafe.Pointer(err))
	}
}

//export ResticFreeString
func ResticFreeString(str *C.char) {
	if str != nil {
		C.free(unsafe.Pointer(str))
	}
}

func cConfigToGo(cfg *C.ResticConfig) *resticlib.Config {
	return &resticlib.Config{
		Endpoint:        C.GoString(cfg.endpoint),
		AccessKeyID:     C.GoString(cfg.access_key_id),
		SecretAccessKey: C.GoString(cfg.secret_access_key),
		BucketName:      C.GoString(cfg.bucket_name),
		Region:          C.GoString(cfg.region),
		Password:        C.GoString(cfg.password),
		Prefix:          C.GoString(cfg.prefix),
		Connections:     int(cfg.connections),
		UseHTTP:         int(cfg.use_http) != 0,
	}
}

func makeError(msg string) *C.ResticError {
	err := (*C.ResticError)(C.malloc(C.sizeof_ResticError))
	err.error_message = C.CString(msg)
	return err
}

func main() {
	// Required for building as shared library
}
