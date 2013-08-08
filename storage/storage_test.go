package storage

import (
    "fmt"
    "testing"
    "crypto/rand"
)

func RandomData() {
}

func TestStoreMany(t *testing.T) {
    store := NewOSStore()

    for i:=0; i<1000; i++ {
        //data := RandomData()
    }

    fmt.Println(store)
}
