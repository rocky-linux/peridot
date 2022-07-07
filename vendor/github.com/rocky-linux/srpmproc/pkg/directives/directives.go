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
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/rocky-linux/srpmproc/pkg/data"
)

func checkAddPrefix(file string) string {
	if strings.HasPrefix(file, "SOURCES/") ||
		strings.HasPrefix(file, "SPECS/") {
		return file
	}

	return filepath.Join("SOURCES", file)
}

func Apply(cfg *srpmprocpb.Cfg, pd *data.ProcessData, md *data.ModeData, patchTree *git.Worktree, pushTree *git.Worktree) []error {
	var errs []error

	directives := []func(*srpmprocpb.Cfg, *data.ProcessData, *data.ModeData, *git.Worktree, *git.Worktree) error{
		replace,
		del,
		add,
		patch,
		lookaside,
		specChange,
	}

	for _, directive := range directives {
		err := directive(cfg, pd, md, patchTree, pushTree)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
