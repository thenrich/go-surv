package cloud

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
)

var S3Uploader *s3manager.Uploader
var Config *aws.Config
var sess *session.Session

func ConfigureS3(cfg *aws.Config) {
	sess = session.Must(session.NewSession(cfg))
	S3Uploader = s3manager.NewUploader(sess)
}

func Uploader() *s3manager.Uploader {
	return S3Uploader
}