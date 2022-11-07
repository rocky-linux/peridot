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
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/rocky-linux/srpmproc/pkg/data"
)

func replace(cfg *srpmprocpb.Cfg, pd *data.ProcessData, _ *data.ModeData, patchTree *git.Worktree, pushTree *git.Worktree) error {
	for _, replace := range cfg.Replace {
		filePath := checkAddPrefix(replace.File)
		stat, err := pushTree.Filesystem.Stat(filePath)
		if replace.File == "" || err != nil {
			return errors.New(fmt.Sprintf("INVALID_FILE:%s", filePath))
		}

		err = pushTree.Filesystem.Remove(filePath)
		if err != nil {
			return errors.New(fmt.Sprintf("COULD_NOT_REMOVE_OLD_FILE:%s", filePath))
		}

		f, err := pushTree.Filesystem.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, stat.Mode())
		if err != nil {
			return errors.New(fmt.Sprintf("COULD_NOT_OPEN_REPLACEMENT:%s", filePath))
		}

		switch replacing := replace.Replacing.(type) {
		case *srpmprocpb.Replace_WithFile:
			fPatch, err := patchTree.Filesystem.OpenFile(replacing.WithFile, os.O_RDONLY, 0o644)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_OPEN_REPLACING:%s", replacing.WithFile))
			}

			replacingBytes, err := ioutil.ReadAll(fPatch)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_READ_REPLACING:%s", replacing.WithFile))
			}

			_, err = f.Write(replacingBytes)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_WRITE_REPLACING:%s", replacing.WithFile))
			}
			break
		case *srpmprocpb.Replace_WithInline:
			_, err := f.Write([]byte(replacing.WithInline))
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_WRITE_INLINE:%s", filePath))
			}
			break
		case *srpmprocpb.Replace_WithLookaside:
			bts, err := pd.BlobStorage.Read(replacing.WithLookaside)
			if err != nil {
				return err
			}
			hasher := pd.CompareHash(bts, replacing.WithLookaside)
			if hasher == nil {
				return errors.New("LOOKASIDE_FILE_AND_HASH_NOT_MATCHING")
			}

			_, err = f.Write(bts)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_WRITE_LOOKASIDE:%s", filePath))
			}
			break
		}
	}

	return nil
}
