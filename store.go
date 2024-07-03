package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	blocksize := 5
	sliceLen := len(hashStr) / blocksize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	FileName string
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s%s", p.PathName, p.FileName)
}

type StoreOpts struct {
	//Root it the folder name of the root, containing all the file/folders of the system
	Root              string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(Key string) PathKey {
	return PathKey{
		PathName: Key,
		FileName: Key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}

	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {
	PathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, PathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	PathKey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("Deleted [%s] from the disk", PathKey.FileName)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, PathKey.FirstPathName())

	// return os.RemoveAll(PathKey.FullPath())
	return os.RemoveAll(firstPathNameWithRoot)

}

func (s *Store) Write(Key string, r io.Reader) error {
	return s.writeStream(Key, r)
}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	PathKey := s.PathTransformFunc(key)
	FullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, PathKey.FullPath())
	f, err := os.Open(FullPathWithRoot)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Store) writeStream(Key string, r io.Reader) error {
	PathKey := s.PathTransformFunc(Key)
	PathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, PathKey.PathName)
	if err := os.MkdirAll(PathNameWithRoot, os.ModePerm); err != nil {
		return err
	}

	FullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, PathKey.FullPath())
	file, err := os.Create(FullPathWithRoot)
	if err != nil {
		return err
	}
	n, err := io.Copy(file, r)
	if err != nil {
		return err
	}
	log.Printf("Written (%d) bytes to disk: %s ", n, FullPathWithRoot)
	return nil
}
