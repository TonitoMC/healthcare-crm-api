package adapters

import (
	"mime/multipart"

	infra "github.com/tonitomc/healthcare-crm-api/internal/infra/s3"
)

// S3Config holds configuration for S3 or MinIO storage.
type S3Config struct {
	Bucket         string
	Region         string
	Endpoint       string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
}

// S3Adapter acts as a bridge between the domain layer and the S3 infrastructure client.
type S3Adapter struct {
	client   *infra.Client
	endpoint string
}

// NewS3Adapter initializes the infra client from config and wraps it in an adapter.
func NewS3Adapter(cfg S3Config) (*S3Adapter, error) {
	client, err := infra.NewClient(
		cfg.Bucket,
		cfg.Endpoint,
		cfg.AccessKey,
		cfg.SecretKey,
		cfg.Region,
		cfg.ForcePathStyle,
	)
	if err != nil {
		return nil, err
	}

	return &S3Adapter{
		client:   client,
		endpoint: cfg.Endpoint,
	}, nil
}

// Upload uploads a file and returns its public URL.
func (a *S3Adapter) Upload(file multipart.File, key, contentType string) (string, error) {
	_, err := a.client.Upload(file, key, contentType)
	if err != nil {
		return "", err
	}
	return a.client.GetPublicURL(a.endpoint, key), nil
}

// Delete removes a file from the bucket.
func (a *S3Adapter) Delete(key string) error {
	return a.client.Delete(key)
}
