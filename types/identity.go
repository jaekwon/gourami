package types

/* The identity package holds code for server & account identities.
 */

import (
    "code.google.com/p/go.crypto/nacl/box"
    "crypto/rand"
    //"encoding/base64"
    "fmt"
)

type Identity struct {
    PublicKey  *PublicKey
    PrivateKey *PrivateKey // usually nil
}

func GenerateIdentity() *Identity {
    pubKey, priKey, _ := box.GenerateKey(rand.Reader)
    return &Identity{(*PublicKey)(pubKey), (*PrivateKey)(priKey)}
}

func (this *Identity) String() string {
    idB64 := this.PublicKey.String()
    if this.PrivateKey == nil {
        return fmt.Sprintf("<Identity (%v)>", idB64)
    } else {
        return fmt.Sprintf("<Identnty {%v}>", idB64)
    }
}
