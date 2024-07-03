package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "bestPictures"
	PathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "6a26bbb2b68d49c54ca853dac179b07b73db0ab8"
	fmt.Println("pathName : ", expectedOriginalKey)
	expectedPathName := "6a26b/bb2b6/8d49c/54ca8/53dac/179b0/7b73d/b0ab8"
	if PathKey.PathName != expectedPathName {
		t.Errorf("pathName is not correct wanted %s", expectedPathName)
	}
	if PathKey.FileName != expectedOriginalKey {
		t.Errorf("original is not correct wanted %s", expectedOriginalKey)
	}
}

// func TestStoreDeleteKey(t *testing.T) {
// 	opts := StoreOpts{
// 		PathTransformFunc: CASPathTransformFunc,
// 	}
// 	s := NewStore(opts)
// 	key := "bestPictures"
// 	data := []byte("some files")

// 	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
// 		t.Error(err)
// 	}

// 	if err := s.Delete(key); err != nil {
// 		t.Error(err)
// 	}
// }

func TestStore(t *testing.T) {
	s := newStore()
	defer tearDown(t, s)

	for i := 0; i < 50; i++ {

		key := fmt.Sprintf("key-%d", i)
		data := []byte("some file")
		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key %s :", key)
		}

		r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}
		b, _ := io.ReadAll(r)
		fmt.Println("Read the data : ", string(b))

		if string(b) != string(data) {
			t.Errorf("error in reading the data")
		}

		if err := s.Delete(key); err != nil {
			t.Errorf("expected to Not have the key %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func tearDown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
