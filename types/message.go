package types

import (
    "io"
    "encoding/json"
    "encoding/binary"
)

type Message struct {
    Header map[string]interface{}
    Content *io.SectionReader
}

// Are the headers valid?
// Error is nil if header is valid.
func (this *Message) ValidateHeader() error {
    // check that all values are strings
    // check for ContentType, Hash, and Datetime
    // TODO
    return nil
}

// Serialize message to writer in byte format
// Caller must close the writer
func (this *Message) Serialize(writer io.Writer) error {
    err := this.ValidateHeader()
    if err != nil { return err }
    headerBytes, err := json.Marshal(this.Header)
    if err != nil { return err }
    err = binary.Write(writer, binary.BigEndian, uint64(len(headerBytes)))
    if err != nil { return err }
    _, err = writer.Write(headerBytes)
    if err != nil { return err }
    err = binary.Write(writer, binary.BigEndian, uint64(this.Content.Size()))
    if err != nil { return err }
    buf := make([]byte, 1024)
    for {
        n, err := this.Content.Read(buf)
        if err != nil && err != io.EOF { return err }
        if n == 0 { break }
        _, err = writer.Write(buf[:n])
        if err != nil { return err }
    }
    return nil
}

// Make a new Message struct from reader
func Deserialize(reader *io.SectionReader) (*Message, error) {
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
