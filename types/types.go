package types

import (
    "encoding/base64"
    "errors"
)

// storage Ids
type Id []byte

// this is just representation.
// if you want typechecking, call ToString() instead
func (this Id) String() string {
    if len(this) != 32 {
        return base64.URLEncoding.EncodeToString(this) + "(Invalid Id)"
    }
    return base64.URLEncoding.EncodeToString(this)
}

func (this Id) ToString() (string, error) {
    if len(this) != 32 {
        return "", errors.New("Invalid Id length")
    }
    return base64.URLEncoding.EncodeToString(this), nil
}

func StringToId(s string) (Id, error) {
    if len(s) != 44 {
        return nil, errors.New("Invalid Id string length")
    }
    idSlice, err := base64.URLEncoding.DecodeString(s)
    return Id(idSlice), err
}
