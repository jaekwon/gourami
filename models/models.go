package models

import (
    . "github.com/jaekwon/go-prelude"
    "fmt"
)

/**
 * Entity
 */

type Entity struct {
    Name string
}

func (this *Entity) Serialize() []byte {
    return []byte(fmt.Sprintf("<Entity %s>", this.Name))
}


/**
 * Mail struct represents data for some entity
 */

type Mail struct {
    from    *Entity
    to      *Entity
    headers Dict
    body    []byte
}

func (*Mail) Serialize() []byte {
    // serialize everything, including internal body data
    return []byte("<Mail>")
}
