package ovhlogs

import (
	"os"
	"time"
)

// Endpoint is OVH log Endpoint
const Endpoint = "gra1.logs.ovh.com"

// Protocol used to push logs to OVH PAAS
type Protocol uint8

const (
	// GelfUDP for Gelf + UDP
	GelfUDP Protocol = 1 + iota
	// GelfTCP for Gelf + TCP
	GelfTCP
	// GelfTLS for Gelf + TLS
	GelfTLS
	// CapnProtoUDP for Cap'n proto + UDP
	CapnProtoUDP
	// CapnProtoTCP for Cap'n proto + TCP
	CapnProtoTCP
	// CapnProtoTLS for Cap'n proto + TLS
	CapnProtoTLS
)

// reverse map
func (p Protocol) String() string {
	switch p {
	case GelfTCP:
		return "GELFTCP"
	case GelfUDP:
		return "GelfUDP"
	case GelfTLS:
		return "GelfTLS"
	default:
		return "UNKNOW"
	}
}

// CompressAlgo is the  compression algorithm used
type CompressAlgo uint8

const (
	// CompressNone No compression
	CompressNone = 1 + iota
	// CompressGzip GZIP compression for GELF
	CompressGzip
	// CompressZlib ZLIB compression for GELF
	CompressZlib
	// CompressDeflate DEFLATE for GELF
	CompressDeflate
)

func (c CompressAlgo) String() string {
	switch c {
	case CompressNone:
		return "no compression"
	case CompressGzip:
		return "gzip"
	case CompressZlib:
		return "CompressZlib"
	case CompressDeflate:
		return "CompressDeflate"
	default:
		return "unknow"
	}

}

const (
	// UDPChunkMaxSizeFrag max chunk size (fragmented)
	UDPChunkMaxSizeFrag = 8192
	// UDPChunkMaxSize chunk max size
	UDPChunkMaxSize = 1420
	// UDPChunkMaxDataSize chunk data max size
	UDPChunkMaxDataSize = 1348 // UDP_CHUNK_MAX_SIZE - ( 2 + 8 + 1 + 1)
)

var (
	// GelfChunkMagicBytes "magic bytes" for GELF chunk headers
	GelfChunkMagicBytes = []byte{0x1e, 0x0f}
)

// OvhLogs represents a OVH logs PAAS wrapper
type OvhLogs struct {
	async       bool
	token       string
	proto       Protocol
	compression CompressAlgo
}

// New return a new OvhLogs
func New(ovhToken string, proto Protocol, compression CompressAlgo, isAsync bool) *OvhLogs {
	return &OvhLogs{
		async:       isAsync,
		token:       ovhToken,
		proto:       proto,
		compression: compression,
	}

}

// Send data to ovh logs
func (o *OvhLogs) Send(e Entry) (err error) {
	// validate entry / set default value

	// OVH token
	e.OvhToken = o.token

	// Version
	e.Version = "1.1"

	// Host
	if e.Host == "" {
		e.Host, err = os.Hostname()
		if err != nil {
			e.Host = "undefined"
		}
	}

	// Timestamp
	if e.Timestamp == 0.0 {
		e.Timestamp = float64(time.Now().UnixNano()/1000000) / 1000.
	}

	// Short message
	if e.ShortMessage == "" {
		if len(e.FullMessage) > 80 {
			e.ShortMessage = e.FullMessage[0:80] + "..."
		} else {
			e.ShortMessage = e.FullMessage
		}
	}

	if o.async {
		go e.send(o.proto, o.compression)
		return nil
	}
	return e.send(o.proto, o.compression)
}
