package gelf

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"log"
	"math"
	"net"
	"os"
	"sync/atomic"
	"time"
)

const (
	defaultMaxChunkSize = 1420
)

var (
	hh = func() []byte {
		id := make([]byte, adler32.Size)
		hash := adler32.New()
		host, err := os.Hostname()
		if err != nil {
			rand.Read(id)
		} else {
			hash.Write([]byte(host))
			copy(id, hash.Sum(nil))
		}
		return id
	}()
)

type Config struct {
	GraylogAddr  string
	MaxChunkSize int
}

type Gelf struct {
	Config
	addr atomic.Value
}

func New(config Config) *Gelf {

	if config.GraylogAddr == "" {
		panic("nil graylog address")
	}
	if config.MaxChunkSize == 0 {
		config.MaxChunkSize = defaultMaxChunkSize
	}

	addr, err := net.ResolveUDPAddr("udp", config.GraylogAddr)
	if err != nil {
		panic(err)
	}

	_, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	g := &Gelf{
		Config: config,
	}

	g.addr.Store(addr)

	//go func() {
	//	t := time.NewTicker(5 * time.Minute)
	//	defer t.Stop()
	//	for range t.C {
	//		if addr, err := net.ResolveUDPAddr("udp", config.GraylogAddr); err == nil {
	//			g.addr.Store(addr)
	//		}
	//	}
	//}()

	return g
}

func (g *Gelf) Log(message []byte) {

	conn, err := net.DialUDP("udp", nil, g.addr.Load().(*net.UDPAddr))
	if err != nil {
		fmt.Printf("write udp failed: %v", err)
		return
	}

	defer conn.Close()

	compressed := g.Compress(message)
	chunksize := g.Config.MaxChunkSize
	length := compressed.Len()

	if length > chunksize {

		chunkCountInt := int(math.Ceil(float64(length) / float64(chunksize)))

		id := make([]byte, 8)
		binary.BigEndian.PutUint64(id, uint64(time.Now().UnixNano()))
		copy(id[:4], hh[:4])

		for i, index := 0, 0; i < length; i, index = i+chunksize, index+1 {
			packet := g.CreateChunkedMessage(index, chunkCountInt, id, compressed)
			g.Send(conn, packet.Bytes())
		}

	} else {
		g.Send(conn, compressed.Bytes())
	}
}

func (g *Gelf) CreateChunkedMessage(index int, chunkCountInt int, id []byte, compressed *bytes.Buffer) bytes.Buffer {
	var packet bytes.Buffer

	chunksize := g.GetChunksize()

	packet.Write(g.IntToBytes(30))
	packet.Write(g.IntToBytes(15))
	packet.Write(id)

	packet.Write(g.IntToBytes(index))
	packet.Write(g.IntToBytes(chunkCountInt))

	packet.Write(compressed.Next(chunksize))

	return packet
}

func (g *Gelf) GetChunksize() int {
	return g.Config.MaxChunkSize
}

func (g *Gelf) IntToBytes(i int) []byte {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, int8(i))
	if err != nil {
		log.Printf("Uh oh! %s", err)
	}
	return buf.Bytes()
}

func (g *Gelf) Compress(b []byte) *bytes.Buffer {

	buf := new(bytes.Buffer)
	comp := zlib.NewWriter(buf)

	comp.Write(b)
	comp.Close()

	return buf
}

func (g *Gelf) Send(conn net.Conn, b []byte) {

	_, err := conn.Write(b)
	if err != nil {
		fmt.Printf("write udp failed: %v", err)
	}
}
