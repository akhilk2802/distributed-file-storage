package main

import "github.com/akhilk2802/distributed-file-storage/p2p"

type FileServerOpts struct {
	storageRoot       string
	pathTransformFunc PathTransformFunc
	Transport         p2p.Transport
}

type FileServer struct {
	FileServerOpts
	store *Store
}

func NewFileServer(opts FileServerOpts) *FileServer {

	storeOpts := &StoreOpts{
		Root:              opts.storageRoot,
		PathTransformFunc: opts.pathTransformFunc,
	}
	fs := &FileServer{
		store:          NewStore(*storeOpts),
		FileServerOpts: opts,
	}
	return fs
}

func (s *FileServerOpts) start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	return nil
}
