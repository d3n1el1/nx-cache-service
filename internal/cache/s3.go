package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"nxCacheService/internal/env"
)

var ErrMissingS3Bucket = errors.New("cache: S3_BUCKET_NAME must be set")

type S3 struct {
	client   *s3.Client
	uploader *transfermanager.Client
	bucket   string
}

var _ Store = (*S3)(nil)

func NewS3(ctx context.Context) (*S3, error) {
	region := env.S3Region.GetValue()
	accessKeyId := env.S3AccessKeyId.GetValue()
	secretAccessKey := env.S3SecretAccessKey.GetValue()
	bucketName := env.S3BucketName.GetValue()
	endpointUrl := env.S3EndpointUrl.GetValue()

	if len(bucketName) == 0 {
		return nil, ErrMissingS3Bucket
	}

	options := []func(*config.LoadOptions) error{}

	if len(region) > 0 {
		options = append(options, config.WithRegion(region))
	}

	if len(accessKeyId) > 0 && len(secretAccessKey) > 0 {
		options = append(options, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""),
		))
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, options...)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if len(endpointUrl) > 0 {
			o.BaseEndpoint = aws.String(endpointUrl)
			o.UsePathStyle = true
		}
	})

	return &S3{
		client:   client,
		uploader: transfermanager.New(client),
		bucket:   bucketName,
	}, nil
}

func (s *S3) Bucket() string { return s.bucket }

func (s *S3) Get(ctx context.Context, project string, hash string) (io.ReadCloser, int64, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey(project, hash)),
	})

	if err != nil {
		if isNotFound(err) {
			return nil, 0, ErrNotFound
		}

		return nil, 0, err
	}

	return out.Body, aws.ToInt64(out.ContentLength), nil
}

func (s *S3) Put(ctx context.Context, project string, hash string, r io.Reader) error {
	_, err := s.uploader.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey(project, hash)),
		Body:        r,
		IfNoneMatch: aws.String("*"),
	})

	if err != nil {
		if isPreconditionFailed(err) {
			return ErrExists
		}

		return err
	}

	return nil
}

func (s *S3) Flush(ctx context.Context, project string) error {
	pages := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(project + "/"),
	})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return err
		}

		if len(page.Contents) == 0 {
			continue
		}

		objects := make([]types.ObjectIdentifier, 0, len(page.Contents))

		for _, object := range page.Contents {
			objects = append(objects, types.ObjectIdentifier{Key: object.Key})
		}

		out, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &types.Delete{Objects: objects, Quiet: aws.Bool(true)},
		})

		if err != nil {
			return err
		}

		if len(out.Errors) > 0 {
			return fmt.Errorf("cache: delete %s: %s",
				aws.ToString(out.Errors[0].Key),
				aws.ToString(out.Errors[0].Message))
		}
	}

	return nil
}

func objectKey(project string, hash string) string {
	return path.Join(project, hash)
}

func isNotFound(err error) bool {
	var noSuchKey *types.NoSuchKey

	if errors.As(err, &noSuchKey) {
		return true
	}

	return hasStatusCode(err, http.StatusNotFound)
}

func isPreconditionFailed(err error) bool {
	var apiErr smithy.APIError

	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "PreconditionFailed" {
		return true
	}

	return hasStatusCode(err, http.StatusPreconditionFailed, http.StatusConflict)
}

func hasStatusCode(err error, codes ...int) bool {
	var respErr *awshttp.ResponseError

	if !errors.As(err, &respErr) {
		return false
	}

	for _, code := range codes {
		if respErr.HTTPStatusCode() == code {
			return true
		}
	}

	return false
}
