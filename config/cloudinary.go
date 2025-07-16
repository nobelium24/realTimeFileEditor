package config

import (
	"context"
	"fmt"
	"realTimeEditor/pkg/constants"

	"github.com/cloudinary/cloudinary-go/v2"
)

func CloudinaryCredentials() (*cloudinary.Cloudinary, context.Context, error) {
	envVars, err := constants.LoadEnv()
	if err != nil {
		return nil, nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	cld, err := cloudinary.NewFromParams(
		envVars.CLOUD_NAME,
		envVars.API_KEY,
		envVars.API_SECRET,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	cld.Config.URL.Secure = true
	ctx := context.Background()

	return cld, ctx, nil
}
