package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3 represents an S3-compatible storage client
type S3 struct {
	Client     *s3.Client
	BucketName string
}

// Config holds the S3 configuration
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

// New creates a new S3 client
func New(ctx context.Context, config Config) (*S3, error) {
	// Create a custom resolver that routes all requests to the specified endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               config.Endpoint,
			HostnameImmutable: true,
			SigningRegion:     "us-east-1", // MinIO doesn't care about the region
		}, nil
	})

	// Create a custom AWS config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.AccessKey,
			config.SecretKey,
			"",
		)),
		config.WithRegion("us-east-1"), // MinIO doesn't care about the region
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // MinIO requires path-style addressing
	})

	s3Client := &S3{
		Client:     client,
		BucketName: config.Bucket,
	}

	// Ensure the bucket exists
	if err := s3Client.EnsureBucketExists(ctx); err != nil {
		return nil, err
	}

	return s3Client, nil
}

// EnsureBucketExists ensures that the configured bucket exists
func (s *S3) EnsureBucketExists(ctx context.Context) error {
	// Check if the bucket exists
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.BucketName),
	})
	if err == nil {
		// Bucket exists
		return nil
	}

	// Create the bucket
	_, err = s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.BucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

// UploadObject uploads an object to S3
func (s *S3) UploadObject(ctx context.Context, key string, data io.Reader, contentType string) error {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.BucketName),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

// GetObject retrieves an object from S3
func (s *S3) GetObject(ctx context.Context, key string) (*s3.GetObjectOutput, error) {
	result, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if ok := errors.As(err, &nsk); ok {
			return nil, fmt.Errorf("object not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return result, nil
}
