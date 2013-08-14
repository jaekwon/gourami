package storage

import (
    "fmt"
    "testing"
    "crypto/rand"
    . "github.com/jaekwon/go-prelude"
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

    if (err != nil) {
        t.Fatal("Could not create new OSStore:", err)
    }

    for i:=0; i<10; i++ {
        id := RandomData(32)
        data := RandomData(64)
        err := store.Store(id, data)
        if (err != nil) {
            fmt.Println("Error! ", err)
        }
    }

    fmt.Println(store, err)

    store.Close()
}

func TestIterate(t *testing.T) {
    store, err := NewOSStore("../.testStore")

    if (err != nil) {
        t.Fatal("Could not create new OSStore:", err)
    }

    ch := make(chan Tuple2)
    go store.(*OSStore).Iterate(ch)
    for t := range ch {
        id, err := t.Get()
        if err != nil {
            fmt.Println(err)
        }
        fmt.Println("----> ", id)
    }

    store.Close()
}
