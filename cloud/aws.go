package cloud

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
)

var S3Uploader *s3manager.Uploader
var Config *aws.Config

func ConfigureS3(cfg *aws.Config) {
	session := session.Must(session.NewSession(cfg))
	S3Uploader = s3manager.NewUploader(session)
}

func Uploader() *s3manager.Uploader {
	return S3Uploader
}