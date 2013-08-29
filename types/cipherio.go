package types

import (
    //"fmt"
    //"github.com/jaekwon/go-prelude/colors"
    "crypto/rand"
    "io"
    "code.google.com/p/go.crypto/nacl/secretbox"
    "log"
    "errors"
    "fmt"
)

/* A CipherWriter encrypts & writes junk to the internal writer field.
 * It also writes the nonce in the beginning */
type CipherWriter struct {
    Key *[32]byte
    Nonce *[24]byte
    Writer io.Writer
    ChunkSize int64

    // state
    chunkIndex int64
    chunk []byte
    written int64 // written to this.Writer
    currentNonce [24]byte
}

func (this *CipherWriter) Reset () {
    this.chunkIndex = int64(0)
    this.chunk = make([]byte, 0, this.ChunkSize)
    this.written = int64(0)
    copy(this.currentNonce[:], this.Nonce[:])
}

// increment chunk index & increment nonce
func (this *CipherWriter) incrementChunk() {
    this.chunkIndex++
    for i:=0;; {
        this.currentNonce[i]++
        if this.currentNonce[i] != 0 { break }
        i++
        if i == len(this.currentNonce) { break } // nonce becomes zeros
    }
}

// encrypt chunk & write, and increment this.Written
func (this *CipherWriter) encryptWriteChunk() error {
    // write nonce if it hasn't already been written
    if this.written == 0 {
        written, err := this.Writer.Write(this.Nonce[:])
        if err != nil { return err }
        this.written += int64(written)
    }
    // chunk
    cipherChunk := secretbox.Seal(nil, this.chunk, this.Nonce, this.Key)
    written, err := this.Writer.Write(cipherChunk)
    this.written += int64(written)
    return err
}

// encrypt & write through to this.Writer whenever this.chunk is full
func (this *CipherWriter) Write (p []byte) (n int, err error) {
    // write p
    for len(p) > 0 {
        // if chunk is full, encrypt & write
        if len(this.chunk) == cap(this.chunk) {
            err := this.encryptWriteChunk()
            if err != nil { return n, err }
            this.incrementChunk()
        }
        // fill up the chunk
        toCopy := min(cap(this.chunk)-len(this.chunk), len(p))
        this.chunk = append(this.chunk, p[:toCopy]...)
        n += toCopy
        p = p[toCopy:]
    }
    return n, nil
}

// close this writer & encrypt & write through if anything left
// does not close underlying this.Writer.
func (this *CipherWriter) Close () error {
    if len(this.chunk) > 0 {
        err := this.encryptWriteChunk()
        if err != nil { return err }
    }
    return nil
}

/* Create a new Cipher Writer
 * Nonce is generated if nil
 */
func NewCipherWriter(writer io.Writer, key *[32]byte, chunkSize int64) (*CipherWriter) {
    if !(1024 < chunkSize && chunkSize <= 1024*1024) { return nil } // sanity check chunk size
    nonce := &[24]byte{}
    rand.Read(nonce[:])
    cipherWriter := &CipherWriter{}
    cipherWriter.Key = key
    cipherWriter.Nonce = nonce
    cipherWriter.Writer = writer
    cipherWriter.ChunkSize = chunkSize
    cipherWriter.Reset()
    return cipherWriter
}


/* CipherReaderAt decipers and reads the underlying reader
 */
type CipherReaderAt struct {
    Reader io.ReaderAt
    Key *[32]byte
    Nonce *[24]byte
    ChunkSize int64
    Chunk []byte
}

// TODO : consider caching the last deciphered chunk
func (this *CipherReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
    log.Printf("ReadAt len(p):%v, off:%v\n", len(p), off)
    // read Nonce if not yet read
    if this.Nonce == nil {
        this.Nonce = &[24]byte{}
        _, err := this.Reader.ReadAt(this.Nonce[:], 0)
        if err != nil { return 0, err }
    }
    // ...
    for len(p) > 0 {
        chunkIndex := off / (this.ChunkSize - secretbox.Overhead)
        chunkStart := chunkIndex * this.ChunkSize
        offChunk := off - chunkIndex * (this.ChunkSize - secretbox.Overhead)
        log.Printf(" - chunkIndex: %v, chunkStart: %v, offChunk: %v\n", chunkIndex, chunkStart, offChunk)

        // compute nonce, basically a base 256 addition operation
        var nonce [24]byte
        copy(nonce[:], this.Nonce[:])
        {
            ci := chunkIndex // copy
            i := 0
            carry := int16(0)
            for ci > 0 || carry > 0 {
                sum := int16(ci % 256)
                sum += int16(nonce[i])
                sum += int16(carry)
                nonce[i] = byte(sum % 256)
                carry = int16(sum >> 8)
                ci >>= 8
                i++
            }
            log.Printf(" - nonce: %v", nonce)
        }
        numRead, err := this.Reader.ReadAt(this.Chunk, 24+chunkStart)
        if err != nil && err != io.EOF { return n, err }
        openedChunk, ok := secretbox.Open(nil, this.Chunk[:numRead], &nonce, this.Key)
        //fmt.Println(colors.Cyan("nonce:", nonce, " key:", this.Key, " numRead:", numRead))
        if !ok { return n, errors.New(fmt.Sprintf("Failed to decipher chunk %v", chunkIndex)) }
        copied := copy(p, openedChunk[offChunk:])
        p = p[copied:]
        n += copied
        off += int64(copied)
    }
    return n, nil
}

func NewCipherReaderAt(reader io.ReaderAt, key *[32]byte, chunkSize int64) (*CipherReaderAt) {
    return &CipherReaderAt{reader, key, nil, chunkSize, make([]byte, chunkSize)}
}
