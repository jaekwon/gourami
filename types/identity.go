package types

/* The identity package holds code for server & account identities.
 */

import (
    "code.google.com/p/go.crypto/nacl/box"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "fmt"
)

type Identity struct {
    PublicKey  *[32]byte
    PrivateKey *[32]byte // usually nil
}

func GenerateIdentity() *Identity {
    pubKey, priKey, _ := box.GenerateKey(rand.Reader)
    return &Identity{pubKey, priKey}
}

func (this *Identity) String() string {
    if this.PrivateKey == nil {
        return fmt.Sprintf("<Identity %v>", KeyToString(this.PublicKey))
    } else {
        return fmt.Sprintf("<Identnty %v!>", KeyToString(this.PublicKey))
    }
}

func NewIdentity(publicKey []byte, privateKey []byte) (*Identity, error) {
    if len(publicKey) != 32 {
        return nil, errors.New("Invalid public key length") }
    if privateKey != nil && len(privateKey) != 32 {
        return nil, errors.New("Invalid private key length") }
    if privateKey != nil {
        var publicKeyA, privateKeyA [32]byte
        copy(publicKeyA[:], publicKey)
        copy(privateKeyA[:], privateKey)
        return &Identity{&publicKeyA, &privateKeyA}, nil
    } else {
        var publicKeyA [32]byte
        copy(publicKeyA[:], publicKey)
        return &Identity{&publicKeyA, nil}, nil
    }
}

func KeyToString(key *[32]byte) string {
    return base64.URLEncoding.EncodeToString(key[:])
}
