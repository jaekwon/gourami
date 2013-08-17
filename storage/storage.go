package storage

import (
    "os"
    "io/ioutil"
    "fmt"
    "errors"
    "strconv"
    "path/filepath"
    "encoding/base64"
    . "github.com/jaekwon/go-prelude"
    "github.com/jaekwon/go-prelude/fs"
    . "github.com/jaekwon/gourami/types"
)

/* Storer is an interface... TODO
 */
type Storer interface {
    Owner() *Identity
    Size() (used int64, capacity int64)
    Store(id Id, data []byte) error
    Delete() error
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
    DataDir string
    Index *Index
}

func (*OSStore) Owner() *Identity {
    return nil
}

func (this *OSStore) Size() (int64, int64) {
    used := int64(0)
    capstring, err := this.Index.Get(MetaCapacity)
    if err != nil { return used, int64(-1) }
    capacity, err := strconv.ParseInt(capstring, 10, 64)
    if err != nil { return used, int64(-1) }
    return used, capacity
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

func (this *OSStore) PathForId(id Id) (string, error) {
    if len(id) != 32 {
        return "", errors.New(fmt.Sprintf("Id was of the wrong length (expected 32, got %v).", len(id)))
    }
    idB64 := base64.URLEncoding.EncodeToString(id)
    path := filepath.Join(this.DataDir, idB64)
    return path, nil
}

func (this *OSStore) OpenDataDir() (*os.File, error) {
    return fs.EnsureDirOpen(this.DataDir)
}

func (this *OSStore) Iterate(ch chan Tuple2) {
    defer close(ch)
    dataDirFile, err := this.OpenDataDir()
    if err != nil {
        ch <- Tuple2{nil, err}
        return
    }
    defer dataDirFile.Close()
    files, err := dataDirFile.Readdirnames(0)
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

func (this *OSStore) Delete() error {
    err := this.Index.Close()
    if err != nil { return err }
    err = os.RemoveAll(this.RootDir)
    return err
}

func NewOSStore(rootDir string, owner *Identity, capacity int64) (Storer, error) {
    var err error
    dataDir := filepath.Join(rootDir, "data")

    if owner == nil { return nil, errors.New("NewOSStore expected non-nil owner") }

    _, err = fs.EnsureDir(rootDir)
    if err != nil { return nil, err }
    _, err = fs.EnsureDir(dataDir)
    if err != nil { return nil, err }
    indexPath := filepath.Join(rootDir, "index.sqlite")
    index, err := NewIndex(indexPath)
    if err != nil { return nil, err }

    // set meta
    index.Set(MetaCapacity, strconv.FormatInt(capacity, 10))
    index.Set(MetaOwner, owner.PublicKey.String())

    return &OSStore{
        RootDir: rootDir,
        DataDir: dataDir,
        Index: index,
    }, nil
}

/* A Storehouser manages many Storers
 */
type Storehouser interface {
    AllocateStorer(owner *Identity, capacity int64) (Storer, error)
    GetStorer(owner *Identity) (Storer, error)
}

type OSStorehouser struct {
    RootDir string
}

func (this *OSStorehouser) AllocateStorer(owner *Identity, capacity int64) (Storer, error) {
    return nil, nil
}

func (this *OSStorehouser) GetStorer(owner *Identity) (Storer, error) {
    return nil, nil
}

func NewOSStorehouser(rootDir string) (Storehouser, error) {
    _, err := fs.EnsureDir(rootDir)
    if err != nil { return nil, err }
    return &OSStorehouser{rootDir}, nil
}
