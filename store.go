package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
)

const defaultRootFolderName = "forever-store"

// CASPathTransformFunc can help with transforming a file path
// hashing the key.
// Content Addressable Storage
func CASPathTransformFunc(root, key string) PathKey {
	hash := sha1.Sum([]byte(key)) // [20]byte -> []byte -> Use [:]
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 8
	sliceLen := len(hashStr) / blockSize

	paths := make([]string, sliceLen)
	paths[0] = root

	for i := 1; i < sliceLen; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashStr[from:to]
	}
	return PathKey{filepath.Join(paths...), hashStr}
}

type PathTransformFunc func(string, string) PathKey

type PathKey struct {
	Pathname string
	Filename string
}

func (p PathKey) FirstParentName() string {
	return filepath.Base(p.Pathname)
}

func (p PathKey) FullPath() string {
	return filepath.Join(p.Pathname, p.Filename)
}

type StoreOpts struct {
	// Root is the folder name of the root containing all the files/folders of the system.
	Root              string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(root, key string) PathKey {
	return PathKey{
		Pathname: root + key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if opts.Root == "" {
		opts.Root = defaultRootFolderName
	}
	return &Store{
		opts,
	}
}

func (s *Store) Has(key string) bool {
	pathkey := s.PathTransformFunc(s.Root, key)

	_, err := os.Stat(pathkey.FullPath())
	return err == nil
}

func (s *Store) Delete(key string) error {
	defer func() {
		log.Printf("deleted [%s] from disk\n", s.Root)
	}()

	return os.RemoveAll(s.Root) // To get First parent directory name
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
	pathkey := s.PathTransformFunc(s.Root, key)
	return os.Open(pathkey.FullPath())
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathkey := s.PathTransformFunc(s.Root, key)
	if err := os.MkdirAll(pathkey.Pathname, os.ModePerm); err != nil {
		return err
	}

	pathAndFileName := pathkey.FullPath()

	f, err := os.Create(pathAndFileName)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk: %s\n", n, pathAndFileName)

	return nil
}
