package types

import (
    "encoding/base64"
)

// storage Ids
type Id []byte

func (this Id) String() string {
    if len(this) != 32 {
        return base64.URLEncoding.EncodeToString(this) + "(Invalid Id)"
    }
    return base64.URLEncoding.EncodeToString(this)
}

// account keys
type PublicKey *[32]byte
type PrivateKey *[32]byte
