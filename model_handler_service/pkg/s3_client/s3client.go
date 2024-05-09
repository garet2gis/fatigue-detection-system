package s3_client

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
)

type ConfigS3 struct {
	Region            string
	S3Host            string
	PartitionID       string
	HostnameImmutable bool
	Bucket            string

	AccessKeyID     string
	SecretAccessKey string
}

type S3Client struct {
	s3Service  *s3.Client
	configS3   ConfigS3
	bucketName string
}

func (s *S3Client) GetPureS3Client() *s3.Client {
	return s.s3Service
}

func NewS3Client(ctx context.Context, cfg ConfigS3) (*S3Client, error) {
	op := "s3_client.NewS3Client"

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       cfg.PartitionID,
			URL:               cfg.S3Host,
			SigningRegion:     cfg.Region,
			HostnameImmutable: cfg.HostnameImmutable,
		}, nil
	})

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(cfg.Region),
		awsConfig.WithEndpointResolverWithOptions(resolver),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	s3Client := s3.NewFromConfig(awsCfg)
	return &S3Client{
		configS3:   cfg,
		s3Service:  s3Client,
		bucketName: cfg.Bucket,
	}, nil
}

func (s *S3Client) DeleteFile(ctx context.Context, fileName string) error {
	op := "s3_client.S3Client.DeleteFile"
	_, err := s.s3Service.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (s *S3Client) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	op := "s3_client.S3Client.ListObjects"
	output, err := s.s3Service.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	var keys []string

	for _, val := range output.Contents {
		if val.Key != nil {
			keys = append(keys, *val.Key)
		}
	}

	return keys, nil
}

func (s *S3Client) DeleteObjects(ctx context.Context, keys []string) error {
	op := "s3_client.S3Client.DeleteObjects"
	objects := make([]types.ObjectIdentifier, 0, len(keys))
	for _, key := range keys {
		objects = append(objects, types.ObjectIdentifier{
			Key: aws.String(key),
		})
	}

	_, err := s.s3Service.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucketName),
		Delete: &types.Delete{Objects: objects},
	})
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return err
}

func (s *S3Client) DeleteFolder(ctx context.Context, folder string) error {
	op := "s3_client.S3Client.DeleteFolder"
	objects, err := s.ListObjects(ctx, folder)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	err = s.DeleteObjects(ctx, objects)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}

func (s *S3Client) GetFile(ctx context.Context, fileName string) (io.ReadCloser, error) {
	op := "s3_client.S3Client.GetFile"
	output, err := s.s3Service.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileName),
	})

	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return output.Body, nil
}

func (s *S3Client) SaveFile(ctx context.Context, key string, file io.Reader) error {
	op := "s3_client.S3Client.SaveFile"
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   file,
	}

	_, err := s.s3Service.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (s *S3Client) GenerateS3DownloadLink(key string) (string, error) {
	//op := "s3_client.S3Client.GenerateS3DownloadLink"
	//result, err := url.JoinPath(s.configS3.Endpoint, s.configS3.Bucket, key)
	//if err != nil {
	//	return "", fmt.Errorf("%s: %w", op, err)
	//}

	// TODO
	return "presigned url", nil
}
