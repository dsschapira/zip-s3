package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	svc        *s3.S3
	s3session  *session.Session
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
)

var region = os.Getenv("S3_REGION")
var accessKey = os.Getenv("S3_ACCESS_KEY")
var secretKey = os.Getenv("S3_SECRET_KEY")

func init() {
	s3session, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		panic(err)
	}

	svc = s3.New(s3session)
	uploader = s3manager.NewUploader(s3session)
	downloader = s3manager.NewDownloader(s3session)
}

func listBuckets() (resp *s3.ListBucketsOutput) {
	resp, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		panic(err)
	}

	return resp
}

func listObjects(bucketname string) (resp *s3.ListObjectsV2Output) {
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketname),
	})

	if err != nil {
		panic(err)
	}
	return resp
}

func main() {
	buckets := listBuckets()
	objects := listObjects(*buckets.Buckets[0].Name)
	fmt.Println(objects.Contents)
}
