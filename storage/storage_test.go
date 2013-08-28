package storage

import (
    "fmt"
    "testing"
    "crypto/rand"
    "syscall"
    . "github.com/jaekwon/go-prelude"
    "github.com/jaekwon/gourami/types"
)

var TestIdentity *types.Identity

func init() {
    TestIdentity = types.GenerateIdentity()
}

func RandomData(length uint) []byte {
    d := make([]byte, length)
    _, err := rand.Read(d)
    if err != nil {
        fmt.Println("error:", err)
        return nil
    }
    return d
}

func TestStoreMany(t *testing.T) {
    store, err := NewOSStore("../.testStore", TestIdentity, 999)

    if err != nil {
        t.Fatal("Could not create new OSStore:", err)
    }

    // test that capacity is set
    used, capacity := store.Size()
    if used != 0 {
        t.Fatal(fmt.Sprintf("Wrong used. Expected 0, got %v", used))
    }
    if capacity != 999 {
        t.Fatal(fmt.Sprintf("Wrong capacity. Expected 999, got %v", capacity))
    }

    // TODO test that the identity is correct

    store.Delete()
}

func TestIterate(t *testing.T) {
    store, err := NewOSStore("../.testStore", TestIdentity, 999)

    if err != nil {
        t.Fatal("Could not create new OSStore:", err)
    }

    for i:=0; i<10; i++ {
        id := RandomData(32)
        data := RandomData(64)
        err := store.Store(id, data)
        if err != nil {
            fmt.Println("Error! ", err)
        }
    }

    ch := make(chan Tuple2)
    go store.(*OSStore).Iterate(ch)
    for tuple := range ch {
        id_, err := tuple.Get()
        id := id_.(types.Id)
        if err != nil {
            fmt.Println(err)
        }
        path, err := store.(*OSStore).PathForId(id)
        if err != nil {
            t.Fatal(err)
        }
        var stat syscall.Stat_t
        err = syscall.Stat(path, &stat)
        if err != nil {
            t.Fatal(err)
        }
        fmt.Printf("----> %v (size: %v, blksize:%v)\n", id, stat.Size, stat.Blksize)
    }

    store.Delete()
}
