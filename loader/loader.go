package loader

import (
	"fmt"
	"io/ioutil"
	"os"
)

var data = make(map[string][]byte)

// Load will load the provided path and prevent it from being reloaded, by caching the read content.
func Load(path string) ([]byte, error) {
	if val, ok := data[path]; ok {
		fmt.Printf("Using cached version of: %v", path)
		return val, nil
	}
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Loading file had an error: %v, error: %v", path, err)
		return nil, fmt.Errorf("file \"%s\" not found on disk: %v", path, err)
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("Failed to read file reader: %v", err)
		return nil, err
	}
	fmt.Printf("read %v from file system", path)
	data[path] = b
	return b, nil
}
