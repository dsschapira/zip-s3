package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

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

func getObject(bucketname string, key string) (resp *s3.GetObjectOutput) {
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
	})

	if err != nil {
		panic(err)
	}
	return resp
}

func getFileList(objects *s3.ListObjectsV2Output) []string {
	var fileList []string
	for _, obj := range objects.Contents {
		fileList = append(fileList, *obj.Key)
	}
	return fileList
}

func main() {
	buckets := listBuckets()
	bucketname := *buckets.Buckets[0].Name
	objects := listObjects(bucketname)
	files := getFileList(objects)

	f, fileErr := os.Create("downloaded/zipped_txt_file.zip")
	if fileErr != nil {
		panic(fileErr)
	}
	zipWriter := zip.NewWriter(f)

	var chans []chan []byte
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			zipWriter.Close()
			f.Close()
			wg.Done()
		}()
		chans = make([]chan []byte, len(files))
		for i := 0; i < len(files); i++ {
			ch := make(chan []byte)
			chans[i] = ch
			go handleFileDownload(ch, &wg, bucketname, files[i])
		}

		cases := make([]reflect.SelectCase, len(chans))
		for j, ch := range chans {
			cases[j] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		}
		remaining := len(cases)
		for remaining > 0 {
			chosen, value, ok := reflect.Select(cases)
			if !ok {
				// The chosen channel has been closed, so zero out the chanel to disable the case
				cases[chosen].Chan = reflect.ValueOf(nil)
				remaining -= 1
				continue
			}
			handleZipAdd(value, zipWriter, &wg, files[chosen])
		}
	}()

	wg.Wait()
}

func handleZipAdd(zc reflect.Value, zw *zip.Writer, wg *sync.WaitGroup, filename string) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()
	fmt.Println("Creating file: ", filename)
	fw, zipErr := zw.Create(filename)
	if zipErr != nil {
		panic(zipErr)
	}
	fw.Write(zc.Bytes())
}

func handleFileDownload(dc chan []byte, wg *sync.WaitGroup, bucketname string, key string) {
	wg.Add(1)
	defer func() {
		close(dc)
		wg.Done()
	}()
	buf := bytes.NewBuffer(make([]byte, 0, 50001000))
	fmt.Println("Downloading... ", key)
	obj := getObject(bucketname, key)
	io.Copy(buf, obj.Body)
	fmt.Printf("Write to channel %#v\n", dc)
	dc <- buf.Bytes()
}
