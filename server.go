package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/akhilk2802/distributed-file-storage/p2p"
)

type FileServerOpts struct {
	storageRoot       string
	pathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootStrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := &StoreOpts{
		Root:              opts.storageRoot,
		PathTransformFunc: opts.pathTransformFunc,
	}
	fs := &FileServer{
		store:          NewStore(*storeOpts),
		FileServerOpts: opts,
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
	return fs
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		return s.store.Read(key)
	}
	fmt.Printf("don't have (%s) file locally, fetching from the network", key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 3)

	for _, peer := range s.peers {
		fmt.Println("Receiving stream from the peer : ", peer.RemoteAddr())
		fileBuf := new(bytes.Buffer)
		n, err := io.CopyN(fileBuf, peer, 30)
		if err != nil {
			return nil, err
		}
		fmt.Println("Recieved bytes over the network : ", n)
		fmt.Println(fileBuf.String())
	}

	select {}

	return nil, errors.New("file not found")
}

// type DataMessage struct {
// 	Key  string
// 	Data []byte
// }

func (s *FileServer) strem(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(&msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
	// 1. Store this file to disk
	// 2. Broadcast this file to all known peers in the network

	fileBuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuffer)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: int(size),
		},
	}
	// msgBuf := new(bytes.Buffer)
	// if err := gob.NewEncoder(msgBuf).Encode(&msg); err != nil {
	// 	return err
	// }

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	for _, peer := range s.peers {
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}
		fmt.Printf("Recieved and written %d to disk", n)
	}
	return nil
}

func (s *FileServer) stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Println("Connected with remote peer : ", p.RemoteAddr().String())
	return nil
}

func (s *FileServer) loop() {

	defer func() {
		log.Println("file server stopped due to error or user quit the action ")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Fatal(err)
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("Error in handling message : ", err)
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
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {

	if !s.store.Has(msg.Key) {
		return fmt.Errorf("cannot read file (%s) from the disk it doesn't exist", msg.Key)
	}

	r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("Peer %s not in map", from)
	}
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("Written %d bytes over the network to  %s \n ", n, from)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Printf(" recv store file msg : %+v\n", msg)
	peer, ok := s.peers[from]
	if !ok {
		return errors.New("peer not found in peers map")
	}

	tcpPeer, ok := peer.(*p2p.TCPPeer)
	if !ok {
		return errors.New("peer is not of type *p2p.TCPPeer")
	}

	if tcpPeer.Wg == nil {
		return errors.New("wg is nil in *p2p.TCPPeer")
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, int64(msg.Size)))
	if err != nil {
		return fmt.Errorf("failed to write to store: %w", err)
	}

	fmt.Printf("written %d bytes to disk \n", n)

	return nil
}

func (s *FileServer) bootStrapNetwork() error {
	for _, addr := range s.BootStrapNodes {
		if len(addr) == 0 {
			continue
		}
		fmt.Println("Attempting to connect with remote : ", addr)
		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				// panic(err)
				log.Println("dial bootstrap node failed ", err)
			}
		}(addr)
		fmt.Println("after the panic")
	}
	return nil
}

func (s *FileServer) start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootStrapNetwork()

	s.loop()

	return nil

}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
