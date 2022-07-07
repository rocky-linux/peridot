// Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
// Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
// Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package s3

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	_ "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-git/go-billy/v5"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"peridot.resf.org/peridot/lookaside"
)

type Storage struct {
	bucket   string
	uploader *s3manager.Uploader
	fs       billy.Filesystem
}

func New(fs billy.Filesystem) (*Storage, error) {
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

	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, err
	}
	uploader := s3manager.NewUploader(sess)

	return &Storage{
		bucket:   viper.GetString("s3-bucket"),
		uploader: uploader,
		fs:       fs,
	}, nil
}

func (s *Storage) DownloadObject(objectName string, path string) error {
	obj, err := s.uploader.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return err
	}

	f, err := s.fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, obj.Body)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) ReadObject(objectName string) ([]byte, error) {
	obj, err := s.uploader.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *Storage) PutObject(objectName string, filePath string) (*lookaside.UploadInfo, error) {
	f, err := s.fs.Open(filePath)
	if err != nil {
		return nil, err
	}

	info, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
		Body:   f,
	})
	if err != nil {
		return nil, err
	}

	return &lookaside.UploadInfo{
		Location:  info.Location,
		VersionID: info.VersionID,
	}, nil
}

func (s *Storage) PutObjectBytes(objectName string, content []byte) (*lookaside.UploadInfo, error) {
	buf := bytes.NewBuffer(content)

	info, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
		Body:   buf,
	})
	if err != nil {
		return nil, err
	}

	return &lookaside.UploadInfo{
		Location:  info.Location,
		VersionID: info.VersionID,
	}, nil
}

func (s *Storage) DeleteObject(objectName string) error {
	_, err := s.uploader.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	})

	return err
}

func (s *Storage) Write(path string, content []byte) error {
	_, err := s.PutObjectBytes(path, content)
	return err
}

func (s *Storage) Read(path string) ([]byte, error) {
	return s.ReadObject(path)
}

func (s *Storage) Exists(path string) (bool, error) {
	_, err := s.uploader.S3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NoSuchKey" {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}
