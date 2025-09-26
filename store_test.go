package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathkey := CASPathTransformFunc(defaultRootFolderName, key)
	expectedoriginalKey := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathName := defaultRootFolderName + "/74181a63/c50c3d81/d733a12f/14a353ff"

	if pathkey.Pathname != expectedPathName {
		t.Errorf("have %s want this %s", pathkey.Pathname, expectedPathName)
	}
	if pathkey.Filename != expectedoriginalKey {
		t.Errorf("have %s want this %s", pathkey.Filename, expectedoriginalKey)
	}
}

func TestDelete(t *testing.T) {
	opts := StoreOpts{
		Root:              defaultRootFolderName,
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "momsspecials"
	data := []byte("some jpg bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		Root:              defaultRootFolderName,
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "momsspecials"
	data := []byte("some jpg bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if ok := s.Has(key); !ok {
		t.Errorf("Expected to have key %s", key)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	fmt.Println(string(b))

	if string(b) != string(data) {
		t.Errorf("want %s have %s", data, b)
	}

	s.Delete(key)
}
