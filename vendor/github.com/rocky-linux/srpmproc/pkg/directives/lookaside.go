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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/rocky-linux/srpmproc/pkg/data"
)

func lookaside(cfg *srpmprocpb.Cfg, _ *data.ProcessData, md *data.ModeData, patchTree *git.Worktree, pushTree *git.Worktree) error {
	for _, directive := range cfg.Lookaside {
		var buf bytes.Buffer
		writer := tar.NewWriter(&buf)
		w := pushTree
		if directive.FromPatchTree {
			w = patchTree
		}

		for _, file := range directive.File {
			if directive.Tar && directive.ArchiveName == "" {
				return errors.New("TAR_NO_ARCHIVE_NAME")
			}

			path := filepath.Join("SOURCES", file)
			if directive.FromPatchTree {
				path = file
			}

			stat, err := w.Filesystem.Stat(path)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_STAT_FILE:%s", path))
			}

			f, err := w.Filesystem.Open(path)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_OPEN_FILE:%s", path))
			}

			bts, err := ioutil.ReadAll(f)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_READ_FILE:%s", path))
			}

			if directive.Tar {
				hdr := &tar.Header{
					Name: file,
					Mode: int64(stat.Mode()),
					Size: stat.Size(),
				}

				err = writer.WriteHeader(hdr)
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_WRITE_TAR_HEADER:%s", file))
				}

				_, err = writer.Write(bts)
				if err != nil {
					return errors.New(fmt.Sprintf("COULD_NOT_WRITE_TAR_FILE:%s", file))
				}
			} else {
				if directive.FromPatchTree {
					pushF, err := pushTree.Filesystem.OpenFile(filepath.Join("SOURCES", filepath.Base(file)), os.O_CREATE|os.O_TRUNC|os.O_RDWR, stat.Mode())
					if err != nil {
						return errors.New(fmt.Sprintf("COULD_NOT_CREATE_FILE_IN_PUSH_TREE:%s", file))
					}

					_, err = pushF.Write(bts)
					if err != nil {
						return errors.New(fmt.Sprintf("COULD_NOT_WRITE_FILE_IN_PUSH_TREE:%s", file))
					}
				}

				md.SourcesToIgnore = append(md.SourcesToIgnore, &data.IgnoredSource{
					Name:         filepath.Join("SOURCES", file),
					HashFunction: sha256.New(),
				})
			}
		}

		if directive.Tar {
			err := writer.Close()
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_CLOSE_TAR:%s", directive.ArchiveName))
			}

			var gbuf bytes.Buffer
			gw := gzip.NewWriter(&gbuf)
			gw.Name = fmt.Sprintf("%s.tar.gz", directive.ArchiveName)
			gw.ModTime = time.Now()

			_, err = gw.Write(buf.Bytes())
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_WRITE_GZIP:%s", directive.ArchiveName))
			}
			err = gw.Close()
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_CLOSE_GZIP:%s", directive.ArchiveName))
			}

			path := filepath.Join("SOURCES", fmt.Sprintf("%s.tar.gz", directive.ArchiveName))
			pushF, err := pushTree.Filesystem.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o644)
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_CREATE_TAR_FILE:%s", path))
			}

			_, err = pushF.Write(gbuf.Bytes())
			if err != nil {
				return errors.New(fmt.Sprintf("COULD_NOT_WRITE_TAR_FILE:%s", path))
			}

			md.SourcesToIgnore = append(md.SourcesToIgnore, &data.IgnoredSource{
				Name:         path,
				HashFunction: sha256.New(),
			})
		}
	}
	return nil
}
