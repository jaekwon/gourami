package storage

import (
    "testing"
    "github.com/jaekwon/go-prelude/fs"
)

func TestMakeIndex(t *testing.T) {

    fs.EnsureDir("../.testStore")

    index, err := NewIndex("../.testStore/index.sqlite")
    if err != nil { t.Fatal(err) }

    err = index.Initialize()
    if err != nil { t.Fatal(err) }
}
