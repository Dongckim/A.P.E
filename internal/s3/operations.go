package s3

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// BucketInfo represents an S3 bucket.
type BucketInfo struct {
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
}

// ObjectInfo represents an S3 object or common prefix (folder).
type ObjectInfo struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	IsDir        bool   `json:"is_dir"`
	LastModified string `json:"last_modified,omitempty"`
	StorageClass string `json:"storage_class,omitempty"`
}

// ObjectList holds objects and pagination info for a ListObjects call.
type ObjectList struct {
	Prefix      string       `json:"prefix"`
	Objects     []ObjectInfo `json:"objects"`
	IsTruncated bool         `json:"is_truncated"`
}

// ListBuckets returns all S3 buckets accessible with the current credentials.
func (c *Client) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	out, err := c.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	buckets := make([]BucketInfo, 0, len(out.Buckets))
	for _, b := range out.Buckets {
		info := BucketInfo{Name: aws.ToString(b.Name)}
		if b.CreationDate != nil {
			info.CreationDate = b.CreationDate.UTC().Format("2006-01-02T15:04:05Z")
		}
		buckets = append(buckets, info)
	}
	return buckets, nil
}

// ListObjects returns objects in a bucket under the given prefix, with folder-like navigation using "/" as a delimiter.
func (c *Client) ListObjects(ctx context.Context, bucket, prefix string) (*ObjectList, error) {
	// Ensure prefix ends with "/" for folder navigation (unless empty/root)
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int32(1000),
	}

	out, err := c.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in %s/%s: %w", bucket, prefix, err)
	}

	objects := make([]ObjectInfo, 0, len(out.CommonPrefixes)+len(out.Contents))

	// Common prefixes are "folders"
	for _, cp := range out.CommonPrefixes {
		folderKey := aws.ToString(cp.Prefix)
		objects = append(objects, ObjectInfo{
			Key:   folderKey,
			IsDir: true,
		})
	}

	// Contents are "files"
	for _, obj := range out.Contents {
		key := aws.ToString(obj.Key)
		// Skip the prefix itself (S3 sometimes returns the folder as an object)
		if key == prefix {
			continue
		}
		info := ObjectInfo{
			Key:  key,
			Size: aws.ToInt64(obj.Size),
		}
		if obj.LastModified != nil {
			info.LastModified = obj.LastModified.UTC().Format("2006-01-02T15:04:05Z")
		}
		if obj.StorageClass != "" {
			info.StorageClass = string(obj.StorageClass)
		}
		objects = append(objects, info)
	}

	return &ObjectList{
		Prefix:      prefix,
		Objects:     objects,
		IsTruncated: aws.ToBool(out.IsTruncated),
	}, nil
}

// UploadObject uploads data from a reader to an S3 object.
func (c *Client) UploadObject(ctx context.Context, bucket, key string, reader io.Reader) error {
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s/%s: %w", bucket, key, err)
	}
	return nil
}

// DownloadObject streams an S3 object into the provided writer.
func (c *Client) DownloadObject(ctx context.Context, bucket, key string, w io.Writer) error {
	out, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download %s/%s: %w", bucket, key, err)
	}
	defer out.Body.Close()

	if _, err := io.Copy(w, out.Body); err != nil {
		return fmt.Errorf("failed to stream %s/%s: %w", bucket, key, err)
	}
	return nil
}

// DeleteObject removes an object from S3.
func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete %s/%s: %w", bucket, key, err)
	}
	return nil
}

// PresignDownload generates a presigned URL for downloading an S3 object.
func (c *Client) PresignDownload(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	req, err := c.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to presign %s/%s: %w", bucket, key, err)
	}
	return req.URL, nil
}
