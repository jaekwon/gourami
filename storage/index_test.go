package storage

import (
    "testing"
)

func TestMakeIndex(t *testing.T) {
    index, err := NewIndex("../.testStore/index.sqlite")
    if err != nil {
        t.Fatal(err)
    }
    err = index.Initialize()
    if err != nil {
        t.Fatal(err)
    }
}
