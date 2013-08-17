package identity

/* The identity package holds code for server & account identities.
 */

import (
    . "github.com/jaekwon/gourami/types"
    "code.google.com/p/go.crypto/nacl/box"
    "crypto/rand"
    "encoding/base64"
    "fmt"
)

type Identity struct {
    Id PublicKey
    PrivateKey PrivateKey // usually nil
}

func GenerateIdentity() *Identity {
    pubKey, priKey, _ := box.GenerateKey(rand.Reader)
    return &Identity{pubKey, priKey}
}

func (this *Identity) String() string {
    idB64 := base64.URLEncoding.EncodeToString(this.Id[:])
    if this.PrivateKey == nil {
        return fmt.Sprintf("<Identity (%v)>", idB64)
    } else {
        return fmt.Sprintf("<Identnty {%v}>", idB64)
    }
}
