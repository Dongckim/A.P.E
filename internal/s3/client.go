package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client defines the interface for S3 operations.
type S3Client interface {
	ListBuckets(ctx context.Context) ([]BucketInfo, error)
	ListObjects(ctx context.Context, bucket, prefix string) (*ObjectList, error)
}

// Client wraps the AWS S3 service client.
type Client struct {
	s3Client *s3.Client
	region   string
}

// New creates an S3 client using the default AWS credential chain (~/.aws/credentials, env vars, etc).
func New(ctx context.Context, region string) (*Client, error) {
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		s3Client: s3.NewFromConfig(cfg),
		region:   region,
	}, nil
}

// Region returns the configured AWS region.
func (c *Client) Region() string {
	return c.region
}
