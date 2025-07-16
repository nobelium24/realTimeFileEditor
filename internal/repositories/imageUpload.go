package repositories

import (
	"fmt"
	"io"
	"realTimeEditor/config"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

func CloudinaryUploader(base64Data string, resourceType ResourceType) (UploadedMedia, error) {
	if !strings.HasPrefix(base64Data, "data:image/") {
		return UploadedMedia{}, fmt.Errorf("invalid base64 image format")
	}

	cld, ctx, err := config.CloudinaryCredentials()
	if err != nil {
		return UploadedMedia{}, fmt.Errorf("cloudinary initialization failed: %w", err)
	}

	uploadResult, err := cld.Upload.Upload(ctx, base64Data, uploader.UploadParams{
		AllowedFormats: []string{"png", "jpg", "jpeg", "gif", "webp"},
		ResourceType:   string(resourceType),
	})

	if err != nil {
		return UploadedMedia{}, fmt.Errorf("upload failed: %w", err)
	}

	return UploadedMedia{
		URL:       uploadResult.URL,
		PublicID:  uploadResult.PublicID,
		SecureURL: uploadResult.SecureURL,
		Width:     uploadResult.Width,
		Height:    uploadResult.Height,
		Format:    uploadResult.Format,
	}, nil
}

func PtrBool(b bool) *bool {
	return &b
}

func CloudinaryUploaderStream(stream io.Reader, fileName string, resourceType ResourceType) (UploadedMedia, error) {
	cld, ctx, err := config.CloudinaryCredentials()
	if err != nil {
		return UploadedMedia{}, fmt.Errorf("cloudinary initialization failed: %w", err)
	}

	// Generate UUID for the public ID
	timestamp := time.Now().UTC().Unix() // Unix timestamp in seconds
	uuidStr := strings.ReplaceAll(uuid.New().String(), "-", "")
	publicID := fmt.Sprintf("file_%d_%s", timestamp, uuidStr)

	uploadResult, err := cld.Upload.Upload(
		ctx,
		stream,
		uploader.UploadParams{
			PublicID:       publicID,
			UseFilename:    PtrBool(true),  // Still preserve original filename in metadata
			UniqueFilename: PtrBool(false), // But we control uniqueness via UUID
			ResourceType:   string(resourceType),
			AllowedFormats: []string{"png", "jpg", "jpeg", "gif", "webp"},
		},
	)

	if err != nil {
		return UploadedMedia{}, fmt.Errorf("upload failed: %w", err)
	}

	return UploadedMedia{
		URL:       uploadResult.URL,
		PublicID:  uploadResult.PublicID, // This will include our UUID prefix
		SecureURL: uploadResult.SecureURL,
		Width:     uploadResult.Width,
		Height:    uploadResult.Height,
		Format:    uploadResult.Format,
	}, nil
}
