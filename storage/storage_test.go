package storage

import (
    "fmt"
    "testing"
    "crypto/rand"
    "encoding/base64"
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
    store := NewOSStore("some_path")

    for i:=0; i<10; i++ {
        id := RandomData(32)
        content := RandomData(64)
        fmt.Println(base64.StdEncoding.EncodeToString(id),
                    base64.StdEncoding.EncodeToString(content))
    }

    fmt.Println(store)
}
