package models

import (
    . "github.com/jaekwon/go-prelude"
)

/**
 * Entity
 */

type Entity struct {
    Name string
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
