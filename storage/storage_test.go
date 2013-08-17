package storage

import (
    "fmt"
    "testing"
    "crypto/rand"
    "syscall"
    . "github.com/jaekwon/go-prelude"
    "github.com/jaekwon/gourami/types"
)

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
    store, err := NewOSStore("../.testStore")

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
}

func TestIterate(t *testing.T) {
    store, err := NewOSStore("../.testStore")

    if err != nil {
        t.Fatal("Could not create new OSStore:", err)
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
}
