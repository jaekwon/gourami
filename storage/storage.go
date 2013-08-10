package storage

import (
    "os"
    "fmt"
    // . "github.com/jaekwon/go-prelude"
    "github.com/jaekwon/gourami/models"
    "github.com/golang/groupcache/lru"
)

// Enumeration type
type StorageStatus uint
const (
    StorageStatusSuccess StorageStatus = iota
    StorageStatusFailure
)

/* Storer is an interface... TODO
 */
type Storer interface {
    Owner() *models.Entity
    Size() (used uint, total uint)
    //Store(*Serializable, chan StorageStatus)
}

type Serializable interface {
    Serialize() []byte
}

/**
 * Storer Implementation
 * NOTE: There is a basic OSStore provided, but users my want
 *  to implement a wrapper to use OpenStack Swift.
 */

/* The OSStore works with the OS's filesystem to implement Storer.
 * We use the native file system handles defragmentation issues.
 * Since filesystems typically support a maximum number of files per directory,
 *  we handle indexing these files via directory structures.
 *
 * NOTE: Consider checking dir_index to see when to split directories.
 * Seems that we should do it around 20K files with dir_index, 5K without.
 * Maybe better to measure performance and do what is optimal.
 */
type OSStore struct {
    RootDir string
    Root *os.File
    DirCache *lru.Cache
}

func NewOSStore(rootDir string) (Storer, error) {
    maxEntries := 100
    s := OSStore{RootDir:rootDir}

    // Set s.Root which is the root directory *File
    var err error
    s.Root, err = os.OpenFile(rootDir, os.O_RDONLY, os.ModeDir)
    if os.IsNotExist(err) {
        // Does not exist so make new dir
        err = os.Mkdir(rootDir, os.ModeDir | 0755)
        s.Root, err = os.OpenFile(rootDir, os.O_RDONLY, os.ModeDir)
        if err != nil {
            return nil, err
        }
    }
    if err != nil {
        return nil, err
    }

    // Check s.Root is a valid directory
    stat, err := s.Root.Stat()
    if err != nil {
        s.Root.Close()
        s.Root = nil
        return nil, err
    }
    if !stat.IsDir() {
        return nil, fmt.Errorf("%v is not a directory", rootDir)
    }

    // Set DirCache, for caching directory entries
    s.DirCache = lru.New(maxEntries)
    return &s, nil
}

func (*OSStore) Owner() *models.Entity {
    return nil
}

func (*OSStore) Size() (uint, uint) {
    return 0, 0
}
