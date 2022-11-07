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
	"bytes"
	"errors"
	"fmt"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/go-git/go-git/v5"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/rocky-linux/srpmproc/pkg/data"
)

func patch(cfg *srpmprocpb.Cfg, pd *data.ProcessData, _ *data.ModeData, patchTree *git.Worktree, pushTree *git.Worktree) error {
	for _, patch := range cfg.Patch {
		patchFile, err := patchTree.Filesystem.Open(patch.File)
		if err != nil {
			return errors.New(fmt.Sprintf("COULD_NOT_OPEN_PATCH_FILE:%s", patch.File))
		}
		files, _, err := gitdiff.Parse(patchFile)
		if err != nil {
			pd.Log.Printf("could not parse patch file: %v", err)
			return errors.New(fmt.Sprintf("COULD_NOT_PARSE_PATCH_FILE:%s", patch.File))
		}

		for _, patchedFile := range files {
			srcPath := patchedFile.NewName
			if !patch.Strict {
				srcPath = checkAddPrefix(patchedFile.NewName)
			}
			var output bytes.Buffer
			if !patchedFile.IsDelete && !patchedFile.IsNew {
				patchSubjectFile, err := pushTree.Filesystem.Open(srcPath)
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_OPEN_PATCH_SUBJECT:%s", srcPath))
				}

				err = gitdiff.NewApplier(patchSubjectFile).ApplyFile(&output, patchedFile)
				if err != nil {
					pd.Log.Printf("could not apply patch: %v", err)
					return errors.New(fmt.Sprintf("COULD_NOT_APPLY_PATCH_WITH_SUBJECT:%s", srcPath))
				}
			}

			oldName := patchedFile.OldName
			if !patch.Strict {
				oldName = checkAddPrefix(patchedFile.OldName)
			}
			_ = pushTree.Filesystem.Remove(oldName)
			_ = pushTree.Filesystem.Remove(srcPath)

			if patchedFile.IsNew {
				newFile, err := pushTree.Filesystem.Create(srcPath)
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_CREATE_NEW_FILE:%s", srcPath))
				}
				err = gitdiff.NewApplier(newFile).ApplyFile(&output, patchedFile)
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_APPLY_PATCH_TO_NEW_FILE:%s", srcPath))
				}
				_, err = newFile.Write(output.Bytes())
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_WRITE_TO_NEW_FILE:%s", srcPath))
				}
				_, err = pushTree.Add(srcPath)
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_ADD_NEW_FILE_TO_GIT:%s", srcPath))
				}
			} else if !patchedFile.IsDelete {
				newFile, err := pushTree.Filesystem.Create(srcPath)
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_CREATE_POST_PATCH_FILE:%s", srcPath))
				}
				_, err = newFile.Write(output.Bytes())
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_WRITE_POST_PATCH_FILE:%s", srcPath))
				}
				_, err = pushTree.Add(srcPath)
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_ADD_POST_PATCH_FILE_TO_GIT:%s", srcPath))
				}
			} else {
				_, err = pushTree.Remove(oldName)
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_REMOVE_FILE_FROM_GIT:%s", oldName))
				}
			}
		}
	}

	return nil
}
