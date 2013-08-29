package types

import (
    "testing"
    "fmt"
    "io"
    "io/ioutil"
    "encoding/base64"
    "crypto/sha512"
    "strings"
    "bytes"
    "encoding/hex"
    "reflect"
    "github.com/jaekwon/go-prelude/colors"
)

func hashString(str string) string {
    hasher := sha512.New()
    io.WriteString(hasher, str)
    hashBytes := hasher.Sum([]byte{})
    hashString := base64.URLEncoding.EncodeToString(hashBytes)
    return hashString
}

func stringSectionReader(str string) *io.SectionReader {
    return io.NewSectionReader(strings.NewReader(str), 0, int64(len(str)))
}

func TestSerialize(t *testing.T) {

    // construct message
    messageStr := "hello world!"
    message := NewMessage(map[string]interface{}{
            "ContentType": "application/octet-stream",
        }, stringSectionReader(messageStr))

    // write to buffer
    var b bytes.Buffer
    err := message.Serialize(&b)
    if err != nil { t.Fatal(err) }
    t.Log("Dump of serialized message:")
    t.Log(hex.Dump(b.Bytes()))

    // deserialize serialized bytes
    messageBytes := b.Bytes()
    message2, err := DeserializeMessage(io.NewSectionReader(bytes.NewReader(messageBytes), 0, int64(len(messageBytes))))
    if err != nil { t.Fatal(err) }

    // test for equality between message and message2
    if !reflect.DeepEqual(message.Header, message2.Header) {
        t.Fatal("message & message2 headers were not equal")
    }
    message2Bytes, err := ioutil.ReadAll(message2.Content)
    if messageStr != string(message2Bytes) {
        t.Fatal("message & message2 contents were not equal.\n Expected: " +
            colors.Green(messageStr) + "\n Got: " +
            colors.Red(string(message2Bytes)))
    }

}

func TestEncrypt(t *testing.T) {

    // construct message
    messageStr := "hello world!"
    message := NewMessage(map[string]interface{}{
            "ContentType": "application/octet-stream",
        }, stringSectionReader(messageStr))

    // create identities
    from := GenerateIdentity()
    to := GenerateIdentity()

    // write cipher message
    var b bytes.Buffer
    err := WriteCipherMessage(&b, message, from, to, "")
    if err != nil { t.Fatal(err) }
    t.Log("Dump of serialized message:")
    t.Log(hex.Dump(b.Bytes()))

    // read cipher message
    reader := bytes.NewReader(b.Bytes())
    sectionReader := io.NewSectionReader(reader, 0, int64(b.Len()))
    cipherMessage, err := DeserializeCipherMessage(sectionReader)
    if err != nil { t.Fatal(err) }
    t.Log("CipherMessage: ", cipherMessage)

    // decipher message
    message2, err := cipherMessage.DecipherMessage(to)
    if err != nil { t.Fatal(err) }
    var m2c bytes.Buffer
    io.Copy(&m2c, message2.Content)
    if m2c.String() != messageStr {
        t.Fatal(fmt.Sprintf("Deciphered message was wrong.\n Expected: %v,\n got: %v", messageStr, m2c.String())) }
}
