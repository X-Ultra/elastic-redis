package cluster

import (
	"net"
	"time"

	"github.com/hashicorp/raft"
)

// Listener is the interface Raft-compatible network layers
// should implement.
type Listener interface {
	net.Listener
	Dial(address string, timeout time.Duration) (net.Conn, error)
}

// Transport is the network service provided to Raft, and wraps a Listener.
type TransportDelegate struct {
	ln Listener
}

// NewTransport returns an initialized Transport.
func NewTransportDelegate(ln Listener) *TransportDelegate {
	return &TransportDelegate{
		ln: ln,
	}
}

// Dial creates a new network connection.
func (t *TransportDelegate) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	return t.ln.Dial(string(addr), timeout)
}

// Accept waits for the next connection.
func (t *TransportDelegate) Accept() (net.Conn, error) {
	return t.ln.Accept()
}

// Close closes the transport
func (t *TransportDelegate) Close() error {
	return t.ln.Close()
}

// Addr returns the binding address of the transport.
func (t *TransportDelegate) Addr() net.Addr {
	return t.ln.Addr()
}
