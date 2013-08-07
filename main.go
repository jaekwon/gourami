package main

import "fmt"
//import "reflect"
import . "github.com/jaekwon/go-prelude"

type StorageUnit uint
type Status uint
type Data []byte

/*
 * The Storage interface is an Gourami flavored interface to the OS filesystem.
 * We use the native file system handles defragmentation issues.
 * Since filesystems typically support a maximum number of files per directory,
 * we handle indexing these files via directory structures.
 *
 * NOTE: Consider checking dir_index to see when to split directories.
 * Seems that we should do it around 20K files with dir_index, 5K without.
 * Maybe better to measure performance and do what is optimal.
 */
type Storage interface {
    Owner() *Entity
    Size() (used StorageUnit, total StorageUnit)
    Store(*Serializable, chan Status)
}

type Serializable interface {
    Serialize() Data
}

/**
 * Entity
 */

type Entity struct {
    name string
}

func (this *Entity) Serialize() Data {
    return Data(fmt.Sprintf("<Entity %s>", this.name))
}


/**
 * Mail struct represents Data for some entity
 */

type Mail struct {
    from Entity
    to Entity
    headers Dict
    body Data
}

func (*Mail) Serialize() Data {
    // serialize everything, including internal body data
    return Data("<Mail>")
}

func main () {
    fmt.Println("Gourami \033[0;34m<°\\\\<\033[0m (version 0.0)")

    user := Entity{name:"Jae"}
    fmt.Println(user)
    fmt.Println(string(user.Serialize()))
}
