package ovhlogs

import (
	"crypto/tls"
	"fmt"
	"net"
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

// Constants usefull for UDP
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
	endpoint    string
	async       bool
	token       string
	proto       Protocol
	compression CompressAlgo
	conn        *net.Conn
}

// New return a new OvhLogs
func New(endpoint, ovhToken string, proto Protocol, compression CompressAlgo, isAsync bool) *OvhLogs {
	return &OvhLogs{
		endpoint:    endpoint,
		async:       isAsync,
		token:       ovhToken,
		proto:       proto,
		compression: compression,
	}
}

// Send a Entry to ovh logs
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

	// get conn
	conn, err := o.getConn()
	if err != nil {
		return
	}

	if o.async {
		go e.send(conn, o.proto, o.compression)
		return nil
	}
	return e.send(conn, o.proto, o.compression)
}

// Get connection
func (o *OvhLogs) getConn() (conn net.Conn, err error) {
	switch o.proto {
	case GelfTCP:
		conn, err = net.DialTimeout("tcp", o.endpoint+":2202", 5*time.Second)
	case GelfTLS:
		conf := &tls.Config{}
		conn, err = tls.Dial("tcp", o.endpoint+":12202", conf)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(10 * time.Second))
	case GelfUDP:
		conn, err = net.DialTimeout("udp", o.endpoint+":2202", 5*time.Second)
	default:
		err = fmt.Errorf("%v not implemented or not supported", o.proto)
	}
	return
}

// implementation of std lib log interface
// Warning thoses methods could return an error

// Printf for log.Printf interface
func (o *OvhLogs) Printf(format string, v ...interface{}) error {
	return o.Send(Entry{
		FullMessage: fmt.Sprintf(format, v...),
	})
}

// Print for log.Print interface
func (o *OvhLogs) Print(v ...interface{}) error {
	return o.Send(Entry{
		FullMessage: fmt.Sprint(v...),
	})
}

// Println for log.Println interface
func (o *OvhLogs) Println(v ...interface{}) error {
	return o.Print(v)
}

// Fatal for log.Fatal interface
// Warning: error are dropped
func (o *OvhLogs) Fatal(v ...interface{}) {
	o.Send(Entry{
		FullMessage: fmt.Sprint(v...),
	})
	os.Exit(1)
}

// Fatalln for log.Fatalln interface
// Warning: error are dropped
func (o *OvhLogs) Fatalln(v ...interface{}) {
	o.Fatal(v)
}

// Fatalf for log.Fatalf interface
// Warning: error are dropped
func (o *OvhLogs) Fatalf(format string, v ...interface{}) {
	o.Send(Entry{
		FullMessage: fmt.Sprintf(format, v...),
	})
	os.Exit(1)
}

// Panic for log.Panic
func (o *OvhLogs) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	o.Send(Entry{
		FullMessage: s,
	})
	panic(s)
}

// Panicln for log.Panicln
func (o *OvhLogs) Panicln(v ...interface{}) {
	o.Panic(v)
}

// Panicf for log.Panicf
func (o *OvhLogs) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	o.Send(Entry{
		FullMessage: s,
	})
	panic(s)
}

// helpers

// Levels used are syslog levels
// 6 -> Info
// 3 -> Error

// Info send log message @info level
func (o *OvhLogs) Info(v ...interface{}) error {
	o.Send(Entry{
		Level:       6,
		FullMessage: fmt.Sprint(v),
	})
	return nil
}

func (o *OvhLogs) Error(v ...interface{}) error {
	o.Send(Entry{
		Level:       3,
		FullMessage: fmt.Sprint(v),
	})
	return nil
}
