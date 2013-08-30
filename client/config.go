package client

import (
    "os"
    "github.com/jaekwon/gourami/types"
    "errors"
    "fmt"
    "encoding/json"
)

type Config struct {
    Version string              `json:"version"`
    Identity *types.Identity    `json:"identity"`
}

func (this *Config) Save(filepath string, password string) error {
    // TODO if err, delete the file
    if _, err := os.Stat(filepath); !os.IsNotExist(err) {
        return errors.New("file already exists")
    }
    file, err := os.OpenFile(filepath, os.O_CREATE | os.O_WRONLY, 0600)
    if err != nil { return err }
    defer file.Close()
    key := &[32]byte{}
    if len(password) > 32 { return errors.New("password too long. max 32 characters.") }
    copy(key[:], []byte(password))
    fmt.Println("key: ", key)
    chunkSize := 10240
    cipherWriter := types.NewCipherWriter(file, key, int64(chunkSize))
    encoder := json.NewEncoder(cipherWriter)
    err = encoder.Encode(this)
    if err == nil {
        cipherWriter.Close() }
    if err != nil {
        os.Remove(filepath)
        return err }
    return file.Sync() // is this necessary?
}

func GenerateConfig() (*Config, error) {
    version := "0"
    identity := types.GenerateIdentity()
    return &Config{version, identity}, nil
}

func NewConfig(filepath string, password string) *Config, error {
    file, err := os.OpenFile(filepath)
    
}
