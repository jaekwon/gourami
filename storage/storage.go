package storage

import (
    //"os"
    //. "github.com/jaekwon/go-prelude"
    "github.com/jaekwon/gourami/models"
    "github.com/golang/groupcache/lru"
)

// Enumeration type
type StorageStatus uint
const (
    StorageStatusSuccess StorageStatus = iota
    StorageStatusFailure
)

/* Storage is an interface... TODO
 */
type Storage interface {
    Owner() *models.Entity
    Size() (used uint, total uint)
    //Store(*Serializable, chan StorageStatus)
}

type Serializable interface {
    Serialize() []byte
}

/**
 * Storage Implementation
 */

/* The OSStore works with the OS's filesystem to implement Storage.
 * We use the native file system handles defragmentation issues.
 * Since filesystems typically support a maximum number of files per directory,
 *  we handle indexing these files via directory structures.
 *
 * NOTE: Consider checking dir_index to see when to split directories.
 * Seems that we should do it around 20K files with dir_index, 5K without.
 * Maybe better to measure performance and do what is optimal.
 */
type OSStore struct {
    DirCache *lru.Cache
}

func NewOSStore() Storage {
    maxEntries := 100
    s := OSStore{}
    s.DirCache = lru.New(maxEntries)
    return &s
}

func (*OSStore) Owner() *models.Entity {
    return nil
}

func (*OSStore) Size() (uint, uint) {
    return 0, 0
}
