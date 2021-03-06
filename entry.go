package ovhlogs

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
)

// Entry represent a log entry to push to OVH logs PAAS
type Entry struct {
	Version      string  `json:"version"`
	Host         string  `json:"host"`
	ShortMessage string  `json:"short_message"`
	FullMessage  string  `json:"full_message"`
	Timestamp    float64 `json:"time_stamp"`
	Level        uint8   `json:"level"`
	Line         uint    `json:"line"`
	OvhToken     string  `json:"_X-OVH-TOKEN"`
}

func (e Entry) send(conn net.Conn, proto Protocol, compression CompressAlgo) (err error) {
	var data []byte
	if data, err = e.gelf(compression); err != nil {
		return err
	}
	defer conn.Close()
	switch proto {
	case GelfTCP, GelfTLS:
		n, err := conn.Write(data)
		if err != nil {
			return err
		}
		if n != len(data) {
			return fmt.Errorf("entry not completely sent %d/%d", n, len(data))
		}
	case GelfUDP:
		if len(data) < UDPChunkMaxSize {
			n, err := conn.Write(data)
			if err != nil {
				return err
			}
			if n != len(data) {
				return fmt.Errorf("entry not completely sent %d/%d", n, len(data))
			}
		} else {
			// chunk buffer
			chunkBuf := bytes.NewBuffer(nil)
			// data buffer
			dataBuf := bytes.NewBuffer(data)

			// nb chunck
			nbChunks := int(math.Ceil(float64(len(data)/UDPChunkMaxDataSize))) + 1

			// MSG ID
			msgID := make([]byte, 8)
			n, err := io.ReadFull(rand.Reader, msgID)
			if err != nil || n != 8 {
				return fmt.Errorf("unable to generate msgID, %v", err)
			}

			for i := 0; i < nbChunks; i++ {
				chunkBuf.Write(GelfChunkMagicBytes)
				chunkBuf.Write(msgID)
				chunkBuf.WriteByte(byte(i))
				chunkBuf.WriteByte(byte(nbChunks))
				for j := 0; j < UDPChunkMaxDataSize; j++ {
					b, err := dataBuf.ReadByte()
					if err != nil {
						if err == io.EOF {
							//log.Println("EOF", dataBuf.Bytes())
							break
						}
						return fmt.Errorf("unable to read from dataBuff, %v", err)
					}
					err = chunkBuf.WriteByte(b)
					if err != nil {
						return fmt.Errorf("unable to write to chunk buffer %v", err)
					}
				}
				// write data
				n, err := conn.Write(chunkBuf.Bytes())
				if err != nil {
					return err
				}
				if n != len(chunkBuf.Bytes()) {
					return fmt.Errorf("entry not completely sent %d/%d", n, len(chunkBuf.Bytes()))
				}

				// reset chunk buffer
				chunkBuf.Reset()
			}
		}
	}
	return nil
}

// Serialize entry for Gelf Proto
func (e Entry) gelf(compression CompressAlgo) (out []byte, err error) {
	out, err = json.Marshal(e)
	if err != nil {
		return []byte{}, fmt.Errorf("ailed to marshal gelfEntry to JSON, %v", err)
	}
	// Compress ?
	if compression != CompressNone {
		var b bytes.Buffer
		switch compression {
		case CompressGzip:
			w := gzip.NewWriter(&b)
			w.Write(out)
			w.Close()
		case CompressZlib:
			w := zlib.NewWriter(&b)
			w.Write(out)
			w.Close()
		default:
			return []byte{}, fmt.Errorf("%v compression not supported", compression)
		}
		out = b.Bytes()
	}

	return out, nil
}
