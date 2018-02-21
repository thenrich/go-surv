package aws

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws"
	"log"
	"github.com/thenrich/go-surv/config"
	"github.com/thenrich/go-surv/cloud"
	"io"
	"github.com/pkg/errors"
)

type S3Storage struct {
	S3     *s3manager.Uploader
	Bucket string
}

func (s3 *S3Storage) UploadFile(r io.ReadCloser, key string) error {
	result, err := s3.S3.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3.Bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return errors.Wrap(err, "error uploading file")
	}
	log.Println(result.Location)
	return nil
}

func NewS3Storage(cfg config.AWSConfig, bucket string) *S3Storage {
	return &S3Storage{S3: cloud.Uploader(), Bucket: bucket}
}
