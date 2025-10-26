package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/Cheemx/forever-store/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transporter
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

type Message struct {
	From    string
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// Store this file to disk
	// Broadcast this file to all known peers in the network

	buf := new(bytes.Buffer)

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: 15,
		},
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("Connected with remote Peer %v", p.RemoteAddr())

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		s.Transport.Close()
		log.Println("File Server Stopped due to User Quit Action")
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println(err)
			}

			if err := s.handleMessage(rpc.From.String(), &msg); err != nil {
				log.Println(err)
				return
			}

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	if err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size)); err != nil {
		return err
	}

	peer.(*p2p.TCPPeer).Wg.Done()

	return nil
}

func (s *FileServer) bootsrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if addr == "" {
			continue
		}
		fmt.Println("Attempting to connect with remote: ", addr)
		go func() {
			if err := s.Transport.Dial(addr); err != nil {
				log.Printf("Dial Error: %v\n", err)
			}
		}()
	}
	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootsrapNetwork()

	s.loop()

	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
}
