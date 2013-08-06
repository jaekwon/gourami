package main

import "fmt"
//import . "github.com/jaekwon/go-prelude"

type StorageUnit uint
type Error uint

type Data interface {
}

type Entity struct {
    name string
}

type StorageA struct {
}

func (*StorageA) Owner () *Entity {
    return nil
}

func (*StorageA) Store (data *Data, trig chan bool) *Error {
    return nil
}

type Storage interface {
    Owner() *Entity
    Store(*Data, chan bool) *Error
}

// The Mail struct represents a piece of Data for somebody
type Mail struct {
    from Entity
    to Entity
    // headers Dict
    data Data
}

func main () {
    fmt.Println("Gourami \033[0;34m<°\\\\<\033[0m (version 0.0)")
}
