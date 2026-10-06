package adapters

import (
	"context"
	"sync"
)

type FileApiStub struct {
	lock  sync.Mutex
	files map[string]string
}

func (c *FileApiStub) UploadFile(ctx context.Context, fileID string, fileContent string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.files == nil {
		c.files = make(map[string]string)
	}

	c.files[fileID] = fileContent

	return nil
}
