package types

import (
    "fmt"
    "io"
    "encoding/json"
    "encoding/binary"
    "encoding/base64"
    "errors"
    "code.google.com/p/go.crypto/nacl/box"
    "code.google.com/p/go.crypto/nacl/secretbox"
    "strconv"
    "log"
    "time"
    "crypto/sha512"
    "crypto/rand"
    //"github.com/jaekwon/go-prelude/colors"
)

func min(a, b int) int { if a < b { return a }; return b }

/* A Message is the plaintext message. Servers do not have access to this, but
 *  rather have access to the encrypted message in a CipherMessage struct
 */
type Message struct {
    Header map[string]interface{}
    Content *io.SectionReader
}

func (this *Message) GetHeader(key string) string {
    v := this.Header[key]
    if v == nil { return "" }
    return v.(string)
}

// Is the header valid?
// Error is nil if header is valid.
func (this *Message) ValidateHeader() (err error) {
    // TODO check that all values are strings
    // TODO check that values are of the correct format
    requireHeader := func(key string) error {
        if this.GetHeader(key) == "" {
            return errors.New(fmt.Sprintf("Required header value missing for key %v", key))
        }
        return nil
    }
    if err = requireHeader("ContentType"); err != nil { return err }
    if err = requireHeader("Hash"); err != nil { return err }
    if err = requireHeader("DateTime"); err != nil { return err }
    return nil
}

/* Serialize message to writer in byte format
 * Caller must close the writer
 */
func (this *Message) Serialize(writer io.Writer) error {
    err := this.ValidateHeader()
    if err != nil { return err }
    headerBytes, err := json.Marshal(this.Header)
    if err != nil { return err }
    err = binary.Write(writer, binary.BigEndian, uint64(len(headerBytes)))
    if err != nil { return err }
    _, err = writer.Write(headerBytes)
    if err != nil { return err }
    _, err = this.Content.Seek(int64(0), 0)
    if err != nil { return err }
    err = binary.Write(writer, binary.BigEndian, uint64(this.Content.Size()))
    if err != nil { return err }
    _, err = this.Content.Seek(int64(0), 0)
    if err != nil { return err }
    _, err = io.Copy(writer, this.Content)
    if err != nil { return err }
    return nil
}

/* Make a new Message struct from reader
 */
func DeserializeMessage(reader *io.SectionReader) (*Message, error) {
    var headerSizeBytes, contentSizeBytes [8]byte
    var headerSize, contentSize uint64
    var header map[string]interface{}
    _, err := reader.ReadAt(headerSizeBytes[:], 0)
    if err != nil { return nil, err }
    headerSize = binary.BigEndian.Uint64(headerSizeBytes[:])
    // TODO sanity check header size
    var headerBytes []byte = make([]byte, headerSize)
    _, err = reader.ReadAt(headerBytes, 8)
    if err != nil { return nil, err }
    err = json.Unmarshal(headerBytes, &header)
    if err != nil { return nil, err }
    // TODO sanity check header
    _, err = reader.ReadAt(contentSizeBytes[:], int64(8+headerSize))
    if err != nil { return nil, err }
    contentSize = binary.BigEndian.Uint64(contentSizeBytes[:])
    contentReader := io.NewSectionReader(reader, int64(8+headerSize+8), int64(contentSize))
    return &Message{Header:header, Content:contentReader}, nil
}

/* Makes a new Message given header and reader
 * Required Headers:
 *  ContentType
 * Optional headers: FileName
 *  DateTime, Hash: computed automatically if not present
 */
func NewMessage(header map[string]interface{}, content *io.SectionReader) *Message {
    if header["DateTime"] == nil {
        now := time.Now()
        header["DateTime"] = now.Format(time.RFC3339)
    }
    if header["Hash"] == nil {
        hasher := sha512.New()
        io.Copy(hasher, content)
        hashBytes := hasher.Sum([]byte{})
        header["Hash"] = base64.URLEncoding.EncodeToString(hashBytes)
    }
    return &Message{header, content}
}

type CipherMessage struct {
    Message
    // memoized...
    from *Identity
    key *[32]byte
    nonce *[24]byte
    chunkSize int64
    chunk []byte
    message *Message // memoized decipherd message
}

// Is the header valid?
// Error is nil if header is valid.
func (this *CipherMessage) ValidateHeader() (err error) {
    // TODO check that all values are strings
    // TODO check that values are of the correct format
    requireHeader := func(key string) error {
        if this.GetHeader(key) == "" {
            return errors.New(fmt.Sprintf("Required header value missing for key %v", key))
        }
        return nil
    }
    if err = requireHeader("To"); err != nil { return err }
    if err = requireHeader("From"); err != nil { return err }
    if err = requireHeader("Hash"); err != nil { return err }
    if err = requireHeader("CipherKey"); err != nil { return err }
    // optional: CipherChunkSize, Permit
    return nil
}

func (this *CipherMessage) From() (*Identity, error) {
    if this.from != nil { return this.from, nil }
    fromIdentBytes, err := base64.URLEncoding.DecodeString(this.GetHeader("From"))
    if err != nil {
        return nil, errors.New("Invalid From identity base64") }
    this.from, err = NewIdentity(fromIdentBytes, nil)
    if err != nil {
        return nil, errors.New("Invalid From identity: "+err.Error()) }
    return this.from, nil
}

func (this *CipherMessage) Nonce() (*[24]byte, error) {
    if this.nonce != nil { return this.nonce, nil }
    nonce := [24]byte{}
    _, err := this.Content.ReadAt(nonce[:], 0)
    if err != nil { return nil, err }
    this.nonce = &nonce
    return this.nonce, nil
}

func (this *CipherMessage) ChunkSize() (chunkSize int64, err error) {
    chunkSizeS := this.GetHeader("CipherChunkSize")
    if chunkSizeS == "" {
        chunkSize = this.Content.Size()
    } else {
        chunkSizeI, err := strconv.Atoi(chunkSizeS)
        if err != nil {
            return -1, errors.New("Invalid chunk size: " + err.Error()) }
        chunkSize = int64(chunkSizeI)
    }
    this.chunkSize = chunkSize
    this.chunk = make([]byte, this.chunkSize)
    return chunkSize, nil
}

/* Decipher the cipher key to reveal the symmetrical key
 */
func (this *CipherMessage) DecipherKey(ident *Identity) (*[32]byte, error) {
    newError := func(err string) error { return errors.New("Cannot decipher cipher key: " + err) }
    if KeyToString(ident.PublicKey) != this.GetHeader("To") {
        return nil, newError("Wrong identity") }
    if ident.PrivateKey == nil {
        return nil, newError("Identity lacks PrivateKey") }
    from, err := this.From()
    if err != nil {
        return nil, newError(err.Error()) }

    // decipher the CipherKey to get the symmetric key
    nonceCipherKey, err := base64.URLEncoding.DecodeString(this.GetHeader("CipherKey"))
    if err != nil {
        return nil, newError("Invalid CipherKey base64") }
    if len(nonceCipherKey) != 24+32+box.Overhead { // 24 byte nonce + 32 byte encrypted bytes + overhead bytes
        return nil, newError("Invalid length.") }
    var nonce [24]byte
    copy(nonce[:], nonceCipherKey[:24])
    cipherKey := nonceCipherKey[24:]
    var key [32]byte
    _, ok := box.Open(key[:0], cipherKey, &nonce, from.PublicKey, ident.PrivateKey)
    if !ok {
        return nil, newError("Failure") }
    this.key = &key
    //fmt.Println(colors.Green(this.key))
    return this.key, nil
}

// Reads decrypted bytes. Do not use this directly, but use DecipherMessage instead.
// TODO : consider caching the last deciphered chunk
func (this *CipherMessage) ReadAt(p []byte, off int64) (n int, err error) {
    log.Printf("ReadAt len(p):%v, off:%v\n", len(p), off)
    for len(p) > 0 {
        chunkIndex := off / (this.chunkSize - secretbox.Overhead)
        chunkStart := chunkIndex * this.chunkSize
        offChunk := off - chunkIndex * (this.chunkSize - secretbox.Overhead)
        log.Printf(" - chunkIndex: %v, chunkStart: %v, offChunk: %v\n", chunkIndex, chunkStart, offChunk)

        // compute nonce, basically a base 256 addition operation
        var nonce [24]byte
        copy(nonce[:], this.nonce[:])
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
        //fmt.Println(colors.Cyan(len(this.chunk), cap(this.chunk)))
        numRead, err := this.Content.ReadAt(this.chunk, 24+chunkStart)
        if err != nil && err != io.EOF { return n, err }
        openedChunk, ok := secretbox.Open(nil, this.chunk[:numRead], &nonce, this.key)
        //fmt.Println(colors.Cyan("nonce:", nonce, " key:", this.key, " numRead:", numRead))
        if !ok { return n, errors.New(fmt.Sprintf("Failed to decipher chunk %v", chunkIndex)) }
        copied := copy(p, openedChunk[offChunk:])
        p = p[copied:]
        n += copied
        off += int64(copied)
    }
    return n, nil
}

/* Decipher the Content and return a message.
 */
func (this *CipherMessage) DecipherMessage(ident *Identity) (*Message, error) {
    newError := func(err error) error { return errors.New("Cannot decipher message: " + err.Error()) }
    _, err := this.DecipherKey(ident)
    if err != nil { return nil, newError(err) }
    chunkSize, err := this.ChunkSize()
    if err != nil { return nil, newError(err) }
    _, err = this.Nonce()
    if err != nil { return nil, newError(err) }
    numChunks := (this.Content.Size() - 24) / chunkSize
    messageSize := this.Content.Size() - 24 - numChunks*secretbox.Overhead
    // a CipherMessage is a ReaderAt, so the returned Message has a Content SectionReader
    //  derived from `this`.
    return DeserializeMessage(io.NewSectionReader(this, 0, messageSize))
}

/* Encrypt & write message
 * Note that there is no way to directly convert a Message into CipherMessage struct,
 *  but you can use this function to serialize one to any writer (e.g. file or network)
 */
func WriteCipherMessage(writer io.Writer, message *Message, from, to *Identity, permit string) error {
    newError := func(err error) error { return errors.New("Cannot encrypt message: " + err.Error()) }
    // need to fill: To, From, Hash, CipherKey, CipherChunkSize, Permit
    // generate a new symmetric key
    var key [32]byte
    _, err := rand.Read(key[:])
    if err != nil { return newError(err) }
    // generate Nonce
    var nonce [24]byte
    _, err = rand.Read(nonce[:])
    if err != nil { return newError(err) }
    // determine appropriate ChunkSize
    // for now, just use 10K bytes
    chunkSize := int64(10240)
    chunkSizeString := strconv.Itoa(int(chunkSize))
    // calculate CipherText hash
    hasher := sha512.New()
    cipherWriter := NewCipherWriter(hasher, &key, &nonce, chunkSize)
    err = message.Serialize(cipherWriter)
    if err != nil { return newError(err) }
    err = cipherWriter.Close()
    if err != nil { return newError(err) }
    cipherMessageSize := cipherWriter.written
    hashBytes := hasher.Sum([]byte{})
    hashString := base64.URLEncoding.EncodeToString(hashBytes)
    // encrypt key to CipherKey
    cipherKey := box.Seal(nil, key[:], &nonce, to.PublicKey, from.PrivateKey)
    cipherKey = append(nonce[:], cipherKey...)
    cipherKeyString := base64.URLEncoding.EncodeToString(cipherKey)
    // make header
    header := map[string]interface{}{
        "To": KeyToString(to.PublicKey),
        "From": KeyToString(from.PublicKey),
        "Hash": hashString,
        "CipherKey": cipherKeyString,
        "CipherChunkSize": chunkSizeString,
        "Permit": permit,
    }
    headerBytes, err := json.Marshal(header)
    if err != nil { return newError(err) }

    // write!
    err = binary.Write(writer, binary.BigEndian, uint64(len(headerBytes)))
    if err != nil { return newError(err) }
    _, err = writer.Write(headerBytes)
    if err != nil { return newError(err) }
    err = binary.Write(writer, binary.BigEndian, uint64(len(nonce))+uint64(cipherMessageSize))
    if err != nil { return newError(err) }
    _, err = writer.Write(nonce[:])
    if err != nil { return newError(err) }
    cipherWriter.Reset()
    cipherWriter.Writer = writer
    err = message.Serialize(cipherWriter)
    if err != nil { return newError(err) }
    err = cipherWriter.Close()
    if err != nil { return newError(err) }
    return nil
}

/* Make a new CipherMessage struct from reader
 */
func DeserializeCipherMessage(reader *io.SectionReader) (*CipherMessage, error) {
    message, err := DeserializeMessage(reader)
    if err != nil { return nil, err }
    return &CipherMessage{Message:*message}, nil
}

/* A CipherWriter encrypts & writes junk to the internal writer field */
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
    cipherChunk := secretbox.Seal(nil, this.chunk, this.Nonce, this.Key)
    //fmt.Println(colors.Blue("nonce:", this.Nonce, " key:", this.Key))
    written, err := this.Writer.Write(cipherChunk)
    this.written += int64(written)
    return err
}

// encrypt & write through to this.Writer whenever this.chunk is full
func (this *CipherWriter) Write (p []byte) (n int, err error) {
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

func NewCipherWriter(writer io.Writer, key *[32]byte, nonce *[24]byte, chunkSize int64) (*CipherWriter) {
    if !(1024 < chunkSize && chunkSize <= 1024*1024) { return nil } // sanity check chunk size
    cipherWriter := &CipherWriter{}
    cipherWriter.Key = key
    cipherWriter.Nonce = nonce
    //fmt.Println(colors.Red("Nonce:", nonce))
    cipherWriter.Writer = writer
    cipherWriter.ChunkSize = chunkSize
    cipherWriter.Reset()
    return cipherWriter
}
