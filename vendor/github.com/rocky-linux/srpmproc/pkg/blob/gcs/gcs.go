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

package gcs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type GCS struct {
	bucket *storage.BucketHandle
}

func New(name string) (*GCS, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create gcloud client: %v", err)
	}

	return &GCS{
		bucket: client.Bucket(name),
	}, nil
}

func (g *GCS) Write(path string, content []byte) error {
	ctx := context.Background()
	obj := g.bucket.Object(path)
	w := obj.NewWriter(ctx)

	_, err := w.Write(content)
	if err != nil {
		return fmt.Errorf("could not write file to gcs: %v", err)
	}

	// Close, just like writing a file.
	if err := w.Close(); err != nil {
		return fmt.Errorf("could not close gcs writer to source: %v", err)
	}

	return nil
}

func (g *GCS) Read(path string) ([]byte, error) {
	ctx := context.Background()
	obj := g.bucket.Object(path)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (g *GCS) Exists(path string) (bool, error) {
	ctx := context.Background()
	obj := g.bucket.Object(path)
	_, err := obj.NewReader(ctx)
	return err == nil, nil
}
