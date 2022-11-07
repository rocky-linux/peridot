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
	"log"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rocky-linux/srpmproc/pkg/blob"
)

type FsCreatorFunc func(branch string) (billy.Filesystem, error)

type ProcessData struct {
	RpmLocation          string
	UpstreamPrefix       string
	Version              int
	GitCommitterName     string
	GitCommitterEmail    string
	Mode                 int
	ModulePrefix         string
	ImportBranchPrefix   string
	BranchPrefix         string
	SingleTag            string
	Authenticator        transport.AuthMethod
	Importer             ImportMode
	BlobStorage          blob.Storage
	NoDupMode            bool
	ModuleMode           bool
	TmpFsMode            string
	NoStorageDownload    bool
	NoStorageUpload      bool
	ManualCommits        []string
	ModuleFallbackStream string
	BranchSuffix         string
	StrictBranchMode     bool
	FsCreator            FsCreatorFunc
	CdnUrl               string
	Log                  *log.Logger
	PackageVersion       string
	PackageRelease       string
	TaglessMode          bool
	AltLookAside         bool
}
