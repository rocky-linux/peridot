// Copyright (c) 2021 The Srpmproc Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package s3

import (
	"bytes"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"
)

type S3 struct {
	bucket   string
	uploader *s3manager.Uploader
}

func New(name string) *S3 {
	awsCfg := &aws.Config{}

	if accessKey := viper.GetString("s3-access-key"); accessKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentials(accessKey, viper.GetString("s3-secret-key"), "")
	}

	if endpoint := viper.GetString("s3-endpoint"); endpoint != "" {
		awsCfg.Endpoint = aws.String(endpoint)
	}

	if region := viper.GetString("s3-region"); region != "" {
		awsCfg.Region = aws.String(region)
	}

	if disableSsl := viper.GetBool("s3-disable-ssl"); disableSsl {
		awsCfg.DisableSSL = aws.Bool(true)
	}

	if forcePathStyle := viper.GetBool("s3-force-path-style"); forcePathStyle {
		awsCfg.S3ForcePathStyle = aws.Bool(true)
	}

	sess := session.Must(session.NewSession(awsCfg))
	uploader := s3manager.NewUploader(sess)

	return &S3{
		bucket:   name,
		uploader: uploader,
	}
}

func (s *S3) Write(path string, content []byte) error {
	buf := bytes.NewBuffer(content)

	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   buf,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3) Read(path string) ([]byte, error) {
	obj, err := s.uploader.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		s3err, ok := err.(awserr.Error)
		if !ok || s3err.Code() != s3.ErrCodeNoSuchKey {
			return nil, err
		}

		return nil, nil
	}

	body, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *S3) Exists(path string) (bool, error) {
	_, err := s.uploader.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	return err == nil, nil
}
