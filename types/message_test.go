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
)

func hashString(str string) string {
    hasher := sha512.New()
    io.WriteString(hasher, str)
    hashBytes := hasher.Sum([]byte{})
    hashString := base64.URLEncoding.EncodeToString(hashBytes)
    return hashString
}

func TestSerialize(t *testing.T) {

    // construct message
    messageStr := "hello world!"
    messageHash := hashString(messageStr)
    header := map[string]interface{}{
        "ContentType":"binary",
        "Hash":messageHash,
        "DateTime":"1937-01-01T12:00:27.87+00:20",
        "Filename":"testfile.txt",
    }
    messageReader := io.NewSectionReader(strings.NewReader(messageStr), 0, int64(len(messageStr)))
    message := &Message{header, messageReader}

    // write to buffer
    var b bytes.Buffer
    err := message.Serialize(&b)
    if err != nil { t.Fatal(err) }
    fmt.Println("Dump of serialized message:")
    fmt.Println(hex.Dump(b.Bytes()))

    // deserialize serialized bytes
    messageBytes := b.Bytes()
    message2, err := Deserialize(io.NewSectionReader(bytes.NewReader(messageBytes), 0, int64(len(messageBytes))))
    if err != nil { t.Fatal(err) }

    // test for equality between message and message2
    if !reflect.DeepEqual(message.Header, message2.Header) {
        t.Fatal("message & message2 headers were not equal")
    }
    message2Bytes, err := ioutil.ReadAll(message2.Content)
    if messageStr != string(message2Bytes) {
        t.Fatal("message & message2 contents were not equal")
    }

}
