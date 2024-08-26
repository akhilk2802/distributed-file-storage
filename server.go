package main

import (
	"bytes"
	"encoding/binary"
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
		fmt.Printf("[%s] Serving file (%s) from local disk \n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(key)
		return r, err
	}
	fmt.Printf("[%s] don't have (%s) file locally, fetching from the network \n", s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers {
		// First read the file size so we can limit the amount of bytes that we need from the connection
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.Write(key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}

		fmt.Printf("[%s] Recieved (%d) bytes over the network from (%s): ", s.Transport.Addr(), n, peer.RemoteAddr())
		peer.CloseStream()
	}

	_, r, err := s.store.Read(key)
	return r, err
}

// type DataMessage struct {
// 	Key  string
// 	Data []byte
// }

func (s *FileServer) stream(msg *Message) error {
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
		peer.Send([]byte{p2p.IncomingMessage})
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

	time.Sleep(time.Second * 2)

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingStream})
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}
		fmt.Printf("Recieved and written %d bytes to disk \n", n)
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
	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)

	fileSize, r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("Peer %s not in map", from)
	}

	//First send the "incoming stream byte to the peer"
	// and then we can send the file size
	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] Written %d bytes over the network to  %s \n ", s.Transport.Addr(), n, from)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Printf(" recv store file msg : %+v\n", msg)
	peer, ok := s.peers[from]
	if !ok {
		return errors.New("peer not found in peers map")
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, int64(msg.Size)))
	if err != nil {
		return fmt.Errorf("failed to write to store: %w", err)
	}

	fmt.Printf("[%s] written %d bytes to disk \n", s.Transport.Addr(), n)

	// peer.(*p2p.TCPPeer).Wg.Done()
	peer.CloseStream()

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
		// fmt.Println("after the panic")
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
