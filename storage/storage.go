package storage

import (
    "os"
    "io/ioutil"
    "fmt"
    "errors"
    "path/filepath"
    "encoding/base64"
    . "github.com/jaekwon/go-prelude"
    . "github.com/jaekwon/gourami/types"
    "github.com/jaekwon/gourami/models"
    "github.com/golang/groupcache/lru"
)

/* Storer is an interface... TODO
 */
type Storer interface {
    Owner() *models.Entity
    Size() (used uint, total uint)
    Store(id Id, data []byte) error
    Close() error
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
    RootFile *os.File
    DirCache *lru.Cache
}

func (*OSStore) Owner() *models.Entity {
    return nil
}

func (*OSStore) Size() (uint, uint) {
    return 0, 0
}

func (this *OSStore) Store(id Id, data []byte) error {
    path, err := this.PathForId(id)
    if err != nil {
        return err
    }
    // if file exists...
    if _, err := os.Stat(path); err == nil {
        return errors.New(fmt.Sprintf("Could not store. File already exists: %v", path))
    }
    err = ioutil.WriteFile(path, data, 0600)
    return err
}

func (this *OSStore) Close() error {
    return this.RootFile.Close()
}

func (this *OSStore) PathForId(id Id) (string, error) {
    if len(id) != 32 {
        return "", errors.New(fmt.Sprintf("Id was of the wrong length (expected 32, got %v).", len(id)))
    }
    idB64 := base64.URLEncoding.EncodeToString(id)
    path := filepath.Join(this.RootDir, idB64)
    return path, nil
}

func (this *OSStore) Iterate(ch chan Tuple2) {
    defer close(ch)
    files, err := this.RootFile.Readdirnames(0)
    if err != nil {
        ch <- Tuple2{nil, err}
        return
    }
    for _, file := range files {
        idBytes, _ := base64.URLEncoding.DecodeString(file)
        id := Id(idBytes)
        ch <- Tuple2{id, nil}
    }
    return
}

func (this *OSStore) GetFile(id Id) (*os.File, error) {
    path, err := this.PathForId(id)
    if err != nil { return nil, err }
    return os.Open(path)
}

func NewOSStore(rootDir string) (Storer, error) {
    maxEntries := 100
    s := OSStore{RootDir:rootDir}

    // Set s.RootFile which is the root directory *os.File
    var err error
    s.RootFile, err = os.OpenFile(rootDir, os.O_RDONLY, os.ModeDir)
    if os.IsNotExist(err) {
        // Does not exist so make new dir
        err = os.Mkdir(rootDir, os.ModeDir | 0755)
        s.RootFile, err = os.OpenFile(rootDir, os.O_RDONLY, os.ModeDir)
        if err != nil {
            return nil, err
        }
    }
    if err != nil {
        return nil, err
    }

    // Check s.RootFile is a valid directory
    stat, err := s.RootFile.Stat()
    if err != nil {
        s.RootFile.Close()
        s.RootFile = nil
        return nil, err
    }
    if !stat.IsDir() {
        return nil, fmt.Errorf("%v is not a directory", rootDir)
    }

    // Set DirCache, for caching directory entries
    s.DirCache = lru.New(maxEntries)
    return &s, nil
}

