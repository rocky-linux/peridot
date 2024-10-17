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

package directives

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/rocky-linux/srpmproc/pkg/data"
)

// returns right if not empty, else left
func eitherString(left string, right string) string {
	if right != "" {
		return right
	}

	return left
}

func add(cfg *srpmprocpb.Cfg, pd *data.ProcessData, md *data.ModeData, patchTree *git.Worktree, pushTree *git.Worktree) error {
	for _, add := range cfg.Add {
		var replacingBytes []byte
		var filePath string

		switch addType := add.Source.(type) {
		case *srpmprocpb.Add_File:
			filePath = checkAddPrefix(eitherString(filepath.Base(addType.File), add.Name))

			fPatch, err := patchTree.Filesystem.OpenFile(addType.File, os.O_RDONLY, 0o644)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_OPEN_FROM:%s", addType.File))
			}

			replacingBytes, err = io.ReadAll(fPatch)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_READ_FROM:%s", addType.File))
			}
			break
		case *srpmprocpb.Add_Lookaside:
			filePath = checkAddPrefix(eitherString(filepath.Base(addType.Lookaside), add.Name))
			var err error
			replacingBytes, err = pd.BlobStorage.Read(addType.Lookaside)
			if err != nil {
				return err
			}

			hashFunction := pd.CompareHash(replacingBytes, addType.Lookaside)
			if hashFunction == nil {
				return errors.New(fmt.Sprintf("LOOKASIDE_HASH_DOES_NOT_MATCH:%s", addType.Lookaside))
			}

			md.SourcesToIgnore = append(md.SourcesToIgnore, &data.IgnoredSource{
				Name:         filePath,
				HashFunction: hashFunction,
			})
			break
		}

		f, err := pushTree.Filesystem.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return errors.New(fmt.Sprintf("COULD_NOT_OPEN_DESTINATION:%s", filePath))
		}

		_, err = f.Write(replacingBytes)
		if err != nil {
			return errors.New(fmt.Sprintf("COULD_NOT_WRITE_DESTIONATION:%s", filePath))
		}
	}

	return nil
}
