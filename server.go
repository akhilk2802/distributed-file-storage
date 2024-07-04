package main

import (
	"fmt"
	"log"
	"sync"

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
		log.Println("file server stopped due to user quit the action ")
		s.Transport.Close()
	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
		case <-s.quitch:
			return
		}
	}
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

	// if len(s.BootStrapNodes) != 0 {
	// 	s.bootStrapNetwork()
	// }

	s.bootStrapNetwork()

	s.loop()

	return nil

}
