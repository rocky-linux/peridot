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

	"github.com/go-git/go-git/v5"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/rocky-linux/srpmproc/pkg/data"
)

func del(cfg *srpmprocpb.Cfg, _ *data.ProcessData, _ *data.ModeData, _ *git.Worktree, pushTree *git.Worktree) error {
	for _, del := range cfg.Delete {
		filePath := del.File
		_, err := pushTree.Filesystem.Stat(filePath)
		if err != nil {
			return errors.New(fmt.Sprintf("FILE_DOES_NOT_EXIST:%s", filePath))
		}

		err = pushTree.Filesystem.Remove(filePath)
		if err != nil {
			return errors.New(fmt.Sprintf("COULD_NOT_DELETE_FILE:%s", filePath))
		}
	}

	return nil
}
