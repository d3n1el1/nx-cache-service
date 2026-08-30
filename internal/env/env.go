package env

import "os"

type EnvKey string

func (key EnvKey) GetValue() string {
	return os.Getenv(string(key))
}

const (
	CiToken       EnvKey = "CI_TOKEN"
	ReadOnlyToken EnvKey = "READ_ONLY_TOKEN"

	S3Region          EnvKey = "S3_REGION"
	S3AccessKeyId     EnvKey = "S3_ACCESS_KEY_ID"
	S3SecretAccessKey EnvKey = "S3_SECRET_ACCESS_KEY"
	S3BucketName      EnvKey = "S3_BUCKET_NAME"
	S3EndpointUrl     EnvKey = "S3_ENDPOINT_URL"
)
