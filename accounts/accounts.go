package accounts

/* The accounts package holds code for storage accounts.
 */

import (
    . "github.com/jaekwon/gourami/types"
    "code.google.com/p/go.crypto/nacl/box"
    "crypto/rand"
    "encoding/base64"
    "fmt"
)

type Account struct {
    Id PublicKey
    PrivateKey PrivateKey // usually nil
}

func GenerateAccount() *Account {
    pubKey, priKey, _ := box.GenerateKey(rand.Reader)
    return &Account{pubKey, priKey}
}

func (this *Account) String() string {
    idB64 := base64.URLEncoding.EncodeToString(this.Id[:])
    if this.PrivateKey == nil {
        return fmt.Sprintf("<Account (%v)>", idB64)
    } else {
        return fmt.Sprintf("<Account {%v}>", idB64)
    }
}
