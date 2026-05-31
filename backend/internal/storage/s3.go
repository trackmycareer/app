package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps an S3-compatible object storage backend.
type Client struct {
	minio     *minio.Client
	bucket    string
	publicURL string
}

// New creates a new S3 storage client. The endpoint should be a host:port
// without scheme (e.g. "localhost:3900"). Set useSSL to true for HTTPS.
func New(endpoint, accessKeyID, secretAccessKey, bucket, publicURL, region string, useSSL bool) (*Client, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("creating S3 client: %w", err)
	}

	return &Client{
		minio:     mc,
		bucket:    bucket,
		publicURL: publicURL,
	}, nil
}

// Upload stores an object in the configured bucket and returns its public URL.
func (c *Client) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := c.minio.PutObject(ctx, c.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: "public, max-age=31536000, immutable",
	})
	if err != nil {
		return "", fmt.Errorf("uploading object %s: %w", key, err)
	}

	return c.publicURL + "/" + key, nil
}

// Delete removes an object from the configured bucket.
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.minio.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
}
