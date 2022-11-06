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

package data

import (
	"hash"

	"github.com/go-git/go-git/v5"
)

type ImportMode interface {
	RetrieveSource(pd *ProcessData) (*ModeData, error)
	WriteSource(pd *ProcessData, md *ModeData) error
	PostProcess(md *ModeData) error
	ImportName(pd *ProcessData, md *ModeData) string
}

type ModeData struct {
	Name            string
	Repo            *git.Repository
	Worktree        *git.Worktree
	FileWrites      map[string][]byte
	TagBranch       string
	PushBranch      string
	Branches        []string
	SourcesToIgnore []*IgnoredSource
	BlobCache       map[string][]byte
}

type IgnoredSource struct {
	Name         string
	HashFunction hash.Hash
	Expired      bool
}
