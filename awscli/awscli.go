// Package awscli contains aws utils used by the bedrock-api
package awscli

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	AccessKeyEnvVar = "AWS_ACCESS_KEY"
	SecretKeyEnvVar = "AWS_SECRET_KEY"
	RegionEnvVar    = "AWS_REGION"
)

// S3Object represents an AWS S3 object
type S3Object struct {
	BucketName string
	Key        string
}

// NewS3Object returns a S3Object
func NewS3Object(bucket, key string) *S3Object {

	return &S3Object{
		BucketName: bucket,
		Key:        key,
	}
}

// PreSignedURL returns a pre-signed aws s3 url that expires in h hours
func (s3o *S3Object) PreSignedURL(sess *session.Session, hours int) (string, error) {

	svc := s3.New(sess)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s3o.BucketName),
		Key:    aws.String(s3o.Key),
	})

	urlStr, err := req.Presign(time.Duration(hours) * time.Hour)
	if err != nil {
		return "", fmt.Errorf("url file %s", err.Error())
	}
	return urlStr, nil
}

// Session returns a new AWS session using envVar
func Session() (*session.Session, error) {

	accessKeyID := os.Getenv(AccessKeyEnvVar)
	secretAccessKey := os.Getenv(SecretKeyEnvVar)
	if accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("empty %v or %v env vars", AccessKeyEnvVar, SecretKeyEnvVar)
	}

	value := credentials.Value{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}

	region := os.Getenv(RegionEnvVar)
	if region == "" {
		region = "us-east-1"
	}

	creds := credentials.NewStaticCredentialsFromCreds(value)
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})
	if err != nil {
		return nil, err
	}
	return sess, nil
}
