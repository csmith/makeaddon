package makeaddon

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
)

var (
	cacheDir = flag.String("cache-dir", "", "Directory to cache dependencies in")
)

// Cache provides persistent directories for checking out dependencies.
type Cache struct {
	dir     string
	content map[string]string
}

// NewCache creates a new cache, loading the previously saved index if it exists.
func NewCache() *Cache {
	cache := &Cache{
		dir:     findCacheDir(),
		content: map[string]string{},
	}
	cache.loadIndex()
	return cache
}

// Dir provides a directory to use for the given url/tag combo. The second return parameter indicates whether the
// directory was newly created (true), or has previously been cached (false).
func (c *Cache) Dir(url, tag string) (string, bool) {
	key := fmt.Sprintf("%s %s", url, tag)
	if existing, ok := c.content[key]; ok {
		return filepath.Join(c.dir, existing), false
	}

	for {
		dir := randomDirName()
		fullDir := filepath.Join(c.dir, dir)
		_, err := os.Stat(fullDir)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(fullDir, os.FileMode(0755)); err != nil {
				log.Panicf("Unable to create cache dir: %v", err)
			}
			c.content[key] = dir
			c.saveIndex()
			return filepath.Join(c.dir, dir), true
		}
	}
}

func (c *Cache) loadIndex() {
	index := filepath.Join(c.dir, "index.json")
	if _, err := os.Stat(index); err == nil {
		b, err := ioutil.ReadFile(index)
		if err != nil {
			log.Printf("Unable to read cache index file: %v", err)
			return
		}
		if err := json.Unmarshal(b, &c.content); err != nil {
			log.Printf("Unable to deserialise cache index file: %v", err)
			return
		}
	}
}

func (c *Cache) saveIndex() {
	index := filepath.Join(c.dir, "index.json")
	b, err := json.Marshal(c.content)
	if err != nil {
		log.Printf("Unable to serialise cache index file: %v", err)
		return
	}

	if err := ioutil.WriteFile(index, b, os.FileMode(0755)); err != nil {
		log.Printf("Unable to write cache index file: %v", err)
	}
}

func findCacheDir() string {
	if *cacheDir != "" {
		return *cacheDir
	}

	dir := ""
	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("LocalAppData")
	case "darwin":
		dir = filepath.Join(os.Getenv("home"), "lib", "cache")
	default:
		dir = os.Getenv("XDG_CACHE_HOME")
		if dir == "" {
			dir = filepath.Join(os.Getenv("HOME"), ".cache")
		}
	}

	if dir == "" {
		dir = os.TempDir()
	}

	return filepath.Join(dir, "makeaddon")
}

func randomDirName() string {
	const (
		chars  = "abcdefghijklmnopqrstuvwxyz0123456789"
		length = 10
	)

	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
