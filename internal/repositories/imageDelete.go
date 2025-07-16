package repositories

import (
	"fmt"
	"realTimeEditor/config"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func CloudinaryDelete(public_id string, resourceType ResourceType) (*uploader.DestroyResult, error) {
	cld, ctx, err := config.CloudinaryCredentials()
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	deleteResult, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     public_id,
		ResourceType: string(resourceType),
	})

	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	return deleteResult, nil
}
