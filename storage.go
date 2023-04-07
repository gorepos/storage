package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Format string

type Options struct {
	Dir string
}

// Storage structure
type Storage struct {
	sync.RWMutex
	options Options
}

var gDefaultOptions = Options{
	Dir: "storage",
}

// Default storage instance
var gStorage Storage = Storage{
	// default options
	options: gDefaultOptions,
}

func NewStorage(options Options) *Storage {
	storage := new(Storage)
	storage.options = gDefaultOptions
	storage.SetOptions(options)
	return storage
}

// SetOptions - set storage directory path (relative or absolute)
func SetOptions(options Options) {
	gStorage.SetOptions(options)
}

// Put - write value to the storage
func Put(key string, value interface{}) error {
	return gStorage.Put(key, value)
}

// Get - get value from the storage
func Get(key string, ref interface{}) error {
	return gStorage.Get(key, ref)
}

// Move - rename key
func Move(srcKey, dstKey string) error {
	return gStorage.Move(srcKey, dstKey)
}

// Delete - remove value from the storage
func Delete(key string) error {
	return gStorage.Delete(key)
}

// Keys - gets all the existing keys in the storage started with 'prefix' string
// if empty string given it returns all the keys
func Keys(prefix string) []string {
	return gStorage.Keys(prefix)
}

// Put - write value to the storage
func (s *Storage) Put(key string, value interface{}) error {
	s.Lock()
	defer s.Unlock()

	jsonFilePath, err := s.keyToPath(key)
	if err != nil {
		return err
	}
	err = mkdirs(jsonFilePath)
	if err != nil {
		return err
	}

	var bytes []byte

	// serialize
	bytes, err = json.MarshalIndent(value, "", "  ")

	if err != nil {
		return err
	}
	err = os.WriteFile(jsonFilePath, bytes, 0666)
	if err != nil {
		return err
	}
	return nil
}

// Get - get value from the storage
func (s *Storage) Get(key string, ref interface{}) error {
	s.RLock()
	defer s.RUnlock()

	jsonFilePath, err := s.keyToPath(key)
	if err != nil {
		return err
	}
	bytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return err
	}

	// deserialize
	err = json.Unmarshal(bytes, ref)

	if err != nil {
		return err
	}

	return nil
}

// Move - rename key
func (s *Storage) Move(oldKey, newKey string) error {
	s.Lock()
	defer s.Unlock()

	srcPath, err := s.keyToPath(oldKey)
	if err != nil {
		return err
	}
	dstPath, err := s.keyToPath(newKey)
	if err != nil {
		return err
	}
	err = mkdirs(dstPath)
	if err != nil {
		return err
	}

	err = os.Rename(srcPath, dstPath)
	if err != nil {
		return err
	}

	// recursively remove empty directories
	parts := strings.Split(oldKey, "/")
	for i := len(parts); i > 0; i-- {
		dirPath := s.options.Dir + "/" + strings.Join(parts[:i], "/")
		if empty, _ := isEmpty(dirPath); empty {
			err := os.Remove(dirPath)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

// Delete - delete key
func (s *Storage) Delete(key string) error {
	s.Lock()
	defer s.Unlock()

	jsonFilePath, err := s.keyToPath(key)
	if err != nil {
		return err
	}
	filepath.Dir(jsonFilePath)
	err = os.Remove(jsonFilePath)
	if err != nil {
		return err
	}

	// recursively remove empty directories
	parts := strings.Split(key, "/")
	for i := len(parts); i > 0; i-- {
		dirPath := s.options.Dir + "/" + strings.Join(parts[:i], "/")
		if empty, _ := isEmpty(dirPath); empty {
			err := os.Remove(dirPath)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return nil
}

// Keys - gets all the existing keys in the storage started with 'startsWith' string
// if empty string given it returns all the keys
func (s *Storage) Keys(prefix string) []string {
	s.RLock()
	defer s.RUnlock()

	var result []string
	err := filepath.Walk(s.options.Dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".json") {
				return nil
			}
			relpath, err := filepath.Rel(s.options.Dir, path)

			key := strings.TrimSuffix(relpath, ".json")
			if strings.HasPrefix(key, prefix) {
				result = append(result, key)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return result
}

// SetOptions - set storage options. Keeps empty options unchanged
func (s *Storage) SetOptions(options Options) {
	if options.Dir != "" {
		s.options.Dir = options.Dir
	}
}

// keyToPath - Build filename path for given key
func (s *Storage) keyToPath(key string) (string, error) {
	// check path traversal
	if strings.HasPrefix(key, "../") ||
		strings.HasPrefix(key, "./") ||
		strings.HasSuffix(key, "/..") ||
		strings.HasSuffix(key, "/.") ||
		strings.Contains(key, "/../") ||
		strings.Contains(key, "/./") {
		return "", fmt.Errorf("path traversal not allowed. Invalid key: '%s'", key)
	}
	thePath := s.options.Dir + "/" + key + ".json"

	// another check: file must be *inside* the storage directory
	absPath, err := filepath.Abs(thePath)
	if err != nil {
		return "", err
	}
	absStoragePath, err := filepath.Abs(s.options.Dir)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, absStoragePath) {
		return "", fmt.Errorf("path is outside storage dir. Invalid key: '%s'", key)
	}

	return thePath, nil
}

// mkdirs - Create a directory (with necessary parents) for filename
func mkdirs(filename string) error {
	directory := filepath.Dir(filename)
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// isEmpty - Check if directory is empty
func isEmpty(directory string) (bool, error) {
	file, err := os.Open(directory)
	if err != nil {
		return false, err
	}
	defer file.Close()

	_, err = file.Readdirnames(1) // Or file.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
