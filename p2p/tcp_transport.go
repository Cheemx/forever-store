package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP established connection.
type TCPPeer struct {
	// conn is the underlying connection of the peer.
	net.Conn

	// If we dial and retrieve a conn => true
	// If we accept and retrieve a conn => false
	outbound bool

	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcCh    chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcCh:            make(chan RPC),
	}
}

// Consume implements the Transport interface
// Which will return *read-only channel*
// Received from another Peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcCh
}

// Send implements the Peer interface
// accepts slice of  bytes and writes it via it's
// underlying connection.
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Write(b)
	return err
}

// Close implements the Transporter interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the Transporter interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()
	fmt.Printf("TCP transport listening on port: %s\n", t.ListenAddr)

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP accept error: %v\n", err)
		}
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("Dropping peer connection: %v\n", err)
		conn.Close()
	}()
	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read Loop
	rpc := &RPC{}
	for {
		err = t.Decoder.Decode(conn, rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr()

		peer.wg.Add(1)
		fmt.Println("waiting till stream is done")
		t.rpcCh <- *rpc
		peer.wg.Wait()
		fmt.Println("stream done continuing normal read loop")
	}
}
