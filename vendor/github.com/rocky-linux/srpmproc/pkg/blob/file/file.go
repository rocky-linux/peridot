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

package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type File struct {
	path string
}

func New(path string) *File {
	return &File{
		path: path,
	}
}

func (f *File) Write(path string, content []byte) error {
	w, err := os.OpenFile(filepath.Join(f.path, path), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}

	_, err = w.Write(content)
	if err != nil {
		return fmt.Errorf("could not write file to file: %v", err)
	}

	// Close, just like writing a file.
	if err := w.Close(); err != nil {
		return fmt.Errorf("could not close file writer to source: %v", err)
	}

	return nil
}

func (f *File) Read(path string) ([]byte, error) {
	r, err := os.OpenFile(filepath.Join(f.path, path), os.O_RDONLY, 0o644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (f *File) Exists(path string) (bool, error) {
	_, err := os.Stat(filepath.Join(f.path, path))
	if !os.IsNotExist(err) {
		if !os.IsExist(err) {
			return false, err
		}
		return true, nil
	}

	return false, nil
}
