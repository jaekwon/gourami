package server

import (
    "net"
    "os"
    "github.com/jaekwon/gourami/models"
    "github.com/jaekwon/gourami/storage"
)

const RECV_BUF_LEN = 1024

type Server struct {
    Listener Listener
    Identity models.Identity
    Storehouser storage.Storehouser
}
