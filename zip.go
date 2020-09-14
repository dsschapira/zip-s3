package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	svc        *s3.S3
	s3Session  *session.Session
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
)

var region = os.Getenv("S3_REGION")
var accessKey = os.Getenv("S3_ACCESS_KEY")
var secretKey = os.Getenv("S3_SECRET_KEY")

func init() {
	s3Session, _ = session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})

	svc = s3.New(s3Session)
	uploader = s3manager.NewUploader(s3Session)
	downloader = s3manager.NewDownloader(s3Session)
}

func main() {

}
