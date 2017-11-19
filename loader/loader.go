package loader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var data = make(map[string][]byte)

// Load will load the provided path and prevent it from being reloaded, by caching the read content.
func Load(path string) (io.Reader, error) {
	if val, ok := data[path]; ok {
		return bytes.NewReader(val), nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("file \"%s\" not found on disk: %v", path, err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	data[path] = b
	return bytes.NewReader(b), nil
}
