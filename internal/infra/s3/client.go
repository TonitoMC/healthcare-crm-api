package s3

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3     *s3.Client
	bucket string
}

// NewClient creates an S3 client that works with AWS or MinIO depending on env vars.
// For MinIO, set S3_ENDPOINT=http://minio:9000 and S3_FORCE_PATH_STYLE=true
func NewClient(bucket string, endpoint string, accessKey string, secretKey string, region string, forcePathStyle bool) (*Client, error) {
	var cfg aws.Config
	var err error

	if endpoint != "" {
		// Local MinIO-style endpoint
		cfg, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load custom S3 config: %w", err)
		}

		client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = forcePathStyle
		})

		return &Client{s3: client, bucket: bucket}, nil
	}

	// Default: AWS environment / IAM role
	cfg, err = config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	return &Client{s3: client, bucket: bucket}, nil
}

func (c *Client) Upload(file multipart.File, key string, contentType string) (string, error) {
	_, err := c.s3.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return fmt.Sprintf("%s/%s", c.bucket, key), nil
}

func (c *Client) Delete(key string) error {
	_, err := c.s3.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (c *Client) GetPublicURL(endpoint string, key string) string {
	if endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", endpoint, c.bucket, key)
	}
	// Default AWS-style public URL
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", c.bucket, key)
}

func (c *Client) Download(key string) (io.ReadCloser, error) {
	out, err := c.s3.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	return out.Body, nil
}
