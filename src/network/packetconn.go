package network

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type PacketConn struct {
	net.PacketConn
	MTU int

	Out SyncWriter
}

// Must support silent drops for certain message sizes

func (c *PacketConn) WriteTo(b []byte, address net.Addr) (int, error) {
	fmt.Fprintf(c.Out, "%4d\n", len(b))
	if len(b) <= c.MTU {
		return c.PacketConn.WriteTo(b, address)
	} else {
		// do not deliver
		return len(b), nil
	}
}

func Listen(address string, mtu int, out io.Writer) (*PacketConn, error) {
	conn, err := net.ListenPacket("udp4", address)
	if err != nil {
		return nil, err
	}
	return &PacketConn{PacketConn: conn, MTU: mtu, Out: out}, nil
}

// NewP2P creates two UDP conn objects that drop payloads larger than size bytes.
func NewMTUConnPair(clientAddr, serverAddr string, mtu int, clientOut, serverOut io.Writer) (client, server net.PacketConn, err error) {
	client, err = Listen(clientAddr, mtu, clientOut)
	if err != nil {
		return nil, nil, fmt.Errorf("client dial error, err = %v", err)
	}

	server, err = Listen(serverAddr, mtu, serverOut)
	if err != nil {
		return nil, nil, fmt.Errorf("server dial error, err = %v", err)
	}
	return client, server, nil
}

// SyncWriter is a concurrency-safe variant of io.Writer.
type SyncWriter interface {
	io.Writer
}

type SyncWrapper struct {
	m sync.Mutex
	w io.Writer
}

func (w *SyncWrapper) Write(p []byte) (int, error) {
	w.m.Lock()
	defer w.m.Unlock()
	return w.w.Write(p)
}

func NewSyncWrapper(w io.Writer) *SyncWrapper {
	return &SyncWrapper{w: w}
}
