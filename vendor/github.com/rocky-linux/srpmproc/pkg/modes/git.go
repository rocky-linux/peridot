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

package modes

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rocky-linux/srpmproc/pkg/misc"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/rocky-linux/srpmproc/pkg/data"
)

type remoteTarget struct {
	remote string
	when   time.Time
}

type remoteTargetSlice []remoteTarget

func (p remoteTargetSlice) Len() int {
	return len(p)
}

func (p remoteTargetSlice) Less(i, j int) bool {
	return p[i].when.Before(p[j].when)
}

func (p remoteTargetSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type GitMode struct{}

func (g *GitMode) RetrieveSource(pd *data.ProcessData) (*data.ModeData, error) {
	repo, err := git.Init(memory.NewStorage(), memfs.New())
	if err != nil {
		return nil, fmt.Errorf("could not init git Repo: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("could not get Worktree: %v", err)
	}

	refspec := config.RefSpec("+refs/heads/*:refs/remotes/*")
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name:  "upstream",
		URLs:  []string{fmt.Sprintf("%s.git", pd.RpmLocation)},
		Fetch: []config.RefSpec{refspec},
	})
	if err != nil {
		return nil, fmt.Errorf("could not create remote: %v", err)
	}

	fetchOpts := &git.FetchOptions{
		Auth:     pd.Authenticator,
		RefSpecs: []config.RefSpec{refspec},
		Tags:     git.AllTags,
		Force:    true,
	}

	err = remote.Fetch(fetchOpts)
	if err != nil {
		if err == transport.ErrInvalidAuthMethod || err == transport.ErrAuthenticationRequired {
			fetchOpts.Auth = nil
			err = remote.Fetch(fetchOpts)
			if err != nil {
				return nil, fmt.Errorf("could not fetch upstream: %v", err)
			}
		} else {
			return nil, fmt.Errorf("could not fetch upstream: %v", err)
		}
	}

	var branches remoteTargetSlice

	latestTags := map[string]*remoteTarget{}

	tagAdd := func(tag *object.Tag) error {
		if strings.HasPrefix(tag.Name, fmt.Sprintf("imports/%s%d", pd.ImportBranchPrefix, pd.Version)) {
			refSpec := fmt.Sprintf("refs/tags/%s", tag.Name)
			if misc.GetTagImportRegex(pd).MatchString(refSpec) {
				match := misc.GetTagImportRegex(pd).FindStringSubmatch(refSpec)

				exists := latestTags[match[2]]
				if exists != nil && exists.when.After(tag.Tagger.When) {
					return nil
				}

				latestTags[match[2]] = &remoteTarget{
					remote: refSpec,
					when:   tag.Tagger.When,
				}
			}
		}
		return nil
	}

	tagIter, err := repo.TagObjects()
	if err != nil {
		return nil, fmt.Errorf("could not get tag objects: %v", err)
	}
	_ = tagIter.ForEach(tagAdd)

	listOpts := &git.ListOptions{
		Auth: pd.Authenticator,
	}
	list, err := remote.List(listOpts)
	if err != nil {
		if err == transport.ErrInvalidAuthMethod || err == transport.ErrAuthenticationRequired {
			listOpts.Auth = nil
			list, err = remote.List(listOpts)
			if err != nil {
				return nil, fmt.Errorf("could not list upstream: %v", err)
			}
		} else {
			return nil, fmt.Errorf("could not list upstream: %v", err)
		}
	}

	for _, ref := range list {
		if ref.Hash().IsZero() {
			continue
		}

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			continue
		}
		_ = tagAdd(&object.Tag{
			Name:   strings.TrimPrefix(string(ref.Name()), "refs/tags/"),
			Tagger: commit.Committer,
		})
	}

	for _, branch := range latestTags {
		pd.Log.Printf("tag: %s", strings.TrimPrefix(branch.remote, "refs/tags/"))
		branches = append(branches, *branch)
	}

	sort.Sort(branches)

	var sortedBranches []string
	for _, branch := range branches {
		sortedBranches = append(sortedBranches, branch.remote)
	}

	return &data.ModeData{
		Name:       filepath.Base(pd.RpmLocation),
		Repo:       repo,
		Worktree:   w,
		FileWrites: nil,
		Branches:   sortedBranches,
	}, nil
}

func (g *GitMode) WriteSource(pd *data.ProcessData, md *data.ModeData) error {
	remote, err := md.Repo.Remote("upstream")
	if err != nil {
		return fmt.Errorf("could not get upstream remote: %v", err)
	}

	var refspec config.RefSpec
	var branchName string

	if strings.HasPrefix(md.TagBranch, "refs/heads") {
		refspec = config.RefSpec(fmt.Sprintf("+%s:%s", md.TagBranch, md.TagBranch))
		branchName = strings.TrimPrefix(md.TagBranch, "refs/heads/")
	} else {
		match := misc.GetTagImportRegex(pd).FindStringSubmatch(md.TagBranch)
		branchName = match[2]
		refspec = config.RefSpec(fmt.Sprintf("+refs/heads/%s:%s", branchName, md.TagBranch))
	}
	pd.Log.Printf("checking out upstream refspec %s", refspec)
	fetchOpts := &git.FetchOptions{
		Auth:       pd.Authenticator,
		RemoteName: "upstream",
		RefSpecs:   []config.RefSpec{refspec},
		Tags:       git.AllTags,
		Force:      true,
	}
	err = remote.Fetch(fetchOpts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		if err == transport.ErrInvalidAuthMethod || err == transport.ErrAuthenticationRequired {
			fetchOpts.Auth = nil
			err = remote.Fetch(fetchOpts)
			if err != nil && err != git.NoErrAlreadyUpToDate {
				return fmt.Errorf("could not fetch upstream: %v", err)
			}
		} else {
			return fmt.Errorf("could not fetch upstream: %v", err)
		}
	}

	err = md.Worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(md.TagBranch),
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("could not checkout source from git: %v", err)
	}

	_, err = md.Worktree.Add(".")
	if err != nil {
		return fmt.Errorf("could not add Worktree: %v", err)
	}

	metadataPath := ""
	ls, err := md.Worktree.Filesystem.ReadDir(".")
	if err != nil {
		return fmt.Errorf("could not read directory: %v", err)
	}
	for _, f := range ls {
		if strings.HasSuffix(f.Name(), ".metadata") {
			if metadataPath != "" {
				return fmt.Errorf("multiple metadata files found")
			}
			metadataPath = f.Name()
		}
	}
	if metadataPath == "" {
		metadataPath = fmt.Sprintf(".%s.metadata", md.Name)
	}

	metadataFile, err := md.Worktree.Filesystem.Open(metadataPath)
	if err != nil {
		pd.Log.Printf("warn: could not open metadata file, so skipping: %v", err)
		return nil
	}

	fileBytes, err := ioutil.ReadAll(metadataFile)
	if err != nil {
		return fmt.Errorf("could not read metadata file: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}
	fileContent := strings.Split(string(fileBytes), "\n")
	for _, line := range fileContent {
		if strings.TrimSpace(line) == "" {
			continue
		}

		lineInfo := strings.SplitN(line, " ", 2)
		hash := strings.TrimSpace(lineInfo[0])
		path := strings.TrimSpace(lineInfo[1])

		var body []byte

		if md.BlobCache[hash] != nil {
			body = md.BlobCache[hash]
			pd.Log.Printf("retrieving %s from cache", hash)
		} else {
			fromBlobStorage, err := pd.BlobStorage.Read(hash)
			if err != nil {
				return err
			}
			if fromBlobStorage != nil && !pd.NoStorageDownload {
				body = fromBlobStorage
				pd.Log.Printf("downloading %s from blob storage", hash)
			} else {
				url := fmt.Sprintf("%s/%s/%s/%s", pd.CdnUrl, md.Name, branchName, hash)
				pd.Log.Printf("downloading %s", url)

				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					return fmt.Errorf("could not create new http request: %v", err)
				}
				req.Header.Set("Accept-Encoding", "*")

				resp, err := client.Do(req)
				if err != nil {
					return fmt.Errorf("could not download dist-git file: %v", err)
				}
				if resp.StatusCode != http.StatusOK {
					url = fmt.Sprintf("%s/%s", pd.CdnUrl, hash)
					req, err = http.NewRequest("GET", url, nil)
					if err != nil {
						return fmt.Errorf("could not create new http request: %v", err)
					}
					req.Header.Set("Accept-Encoding", "*")
					resp, err = client.Do(req)
					if err != nil {
						return fmt.Errorf("could not download dist-git file: %v", err)
					}
					if resp.StatusCode != http.StatusOK {
						return fmt.Errorf("could not download dist-git file (status code %d): %v", resp.StatusCode, err)
					}
				}

				body, err = ioutil.ReadAll(resp.Body)
				if err != nil {
					return fmt.Errorf("could not read the whole dist-git file: %v", err)
				}
				err = resp.Body.Close()
				if err != nil {
					return fmt.Errorf("could not close body handle: %v", err)
				}
			}

			md.BlobCache[hash] = body
		}

		f, err := md.Worktree.Filesystem.Create(path)
		if err != nil {
			return fmt.Errorf("could not open file pointer: %v", err)
		}

		hasher := pd.CompareHash(body, hash)
		if hasher == nil {
			return fmt.Errorf("checksum in metadata does not match dist-git file")
		}

		md.SourcesToIgnore = append(md.SourcesToIgnore, &data.IgnoredSource{
			Name:         path,
			HashFunction: hasher,
		})

		_, err = f.Write(body)
		if err != nil {
			return fmt.Errorf("could not copy dist-git file to in-tree: %v", err)
		}
		_ = f.Close()
	}

	return nil
}

func (g *GitMode) PostProcess(md *data.ModeData) error {
	for _, source := range md.SourcesToIgnore {
		_, err := md.Worktree.Filesystem.Stat(source.Name)
		if err == nil {
			err := md.Worktree.Filesystem.Remove(source.Name)
			if err != nil {
				return fmt.Errorf("could not remove dist-git file: %v", err)
			}
		}
	}

	_, err := md.Worktree.Add(".")
	if err != nil {
		return fmt.Errorf("could not add git sources: %v", err)
	}

	return nil
}

func (g *GitMode) ImportName(pd *data.ProcessData, md *data.ModeData) string {
	if misc.GetTagImportRegex(pd).MatchString(md.TagBranch) {
		match := misc.GetTagImportRegex(pd).FindStringSubmatch(md.TagBranch)
		return match[3]
	}

	return strings.Replace(strings.TrimPrefix(md.TagBranch, "refs/heads/"), "%", "_", -1)
}
