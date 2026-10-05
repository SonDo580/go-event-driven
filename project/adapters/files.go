package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

type FilesApiClient struct {
	clients *clients.Clients
}

func NewFilesApiClient(clients *clients.Clients) *FilesApiClient {
	if clients == nil {
		panic("NewFilesApiClient: clients is nil")
	}

	return &FilesApiClient{clients: clients}
}

func (c FilesApiClient) UploadFile(ctx context.Context, fileID string, fileContent string) error {
	resp, err := c.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(ctx, fileID, fileContent)
	if err != nil {
		return fmt.Errorf("failed to upload file %s: %w", fileID, err)
	}

	switch resp.StatusCode() {
	case http.StatusConflict:
		log.FromContext(ctx).With("file", fileID).Info("file already exists")
		return nil
	case http.StatusCreated:
		return nil
	default:
		return fmt.Errorf("unexpected status code while uploading file %s: %d", fileID, resp.StatusCode())
	}
}
