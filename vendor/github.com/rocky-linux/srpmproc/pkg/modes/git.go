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
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rocky-linux/srpmproc/pkg/misc"

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

// Struct to define the possible template values ( {{.Value}} in CDN URL strings:
type Lookaside struct {
	Name     string
	Branch   string
	Hash     string
	Hashtype string
	Filename string
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

	// In case of "tagless mode", we need to get the head ref of the branch instead
	// This is a kind of alternative implementation of the above tagAdd assignment
	refAdd := func(tag *object.Tag) error {
		if misc.TaglessRefOk(tag.Name, pd) {
			pd.Log.Printf("Tagless mode:  Identified tagless commit for import: %s\n", tag.Name)
			refSpec := fmt.Sprintf(tag.Name)

			// We split the string by "/", the branch name we're looking for to pass to latestTags is always last
			// (ex: "refs/heads/c9s" ---> we want latestTags[c9s]
			tmpRef := strings.Split(refSpec, "/")
			tmpBranchName := tmpRef[(len(tmpRef) - 1)]

			latestTags[tmpBranchName] = &remoteTarget{
				remote: refSpec,
				when:   tag.Tagger.When,
			}
		}
		return nil
	}

	tagIter, err := repo.TagObjects()
	if err != nil {
		return nil, fmt.Errorf("could not get tag objects: %v", err)
	}

	// tagless mode means we use "refAdd" (add commit by reference)
	// normal mode means we can rely on "tagAdd" (the tag should be present for us in the source repo)
	if pd.TaglessMode {
		_ = tagIter.ForEach(refAdd)
	} else {
		_ = tagIter.ForEach(tagAdd)
	}

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

		// Call refAdd instead of tagAdd in the case of TaglessMode enabled
		if pd.TaglessMode {
			_ = refAdd(&object.Tag{
				Name:   string(ref.Name()),
				Tagger: commit.Committer,
			})
		} else {
			_ = tagAdd(&object.Tag{
				Name:   strings.TrimPrefix(string(ref.Name()), "refs/tags/"),
				Tagger: commit.Committer,
			})
		}

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

	if err != nil && !pd.TaglessMode {
		return fmt.Errorf("could not get upstream remote: %v", err)
	}

	var refspec config.RefSpec
	var branchName string

	// In the case of tagless mode, we already have the transformed repo sitting in the worktree,
	// and don't need to perform any checkout or fetch operations
	if !pd.TaglessMode {
		if strings.HasPrefix(md.TagBranch, "refs/heads") {
			refspec = config.RefSpec(fmt.Sprintf("+%s:%s", md.TagBranch, md.TagBranch))
			branchName = strings.TrimPrefix(md.TagBranch, "refs/heads/")
		} else {
			match := misc.GetTagImportRegex(pd).FindStringSubmatch(md.TagBranch)
			branchName = match[2]
			refspec = config.RefSpec(fmt.Sprintf("+refs/heads/%s:%s", branchName, md.TagBranch))
			fmt.Println("Found branchname that does not start w/ refs/heads :: ", branchName)
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
	}

	if pd.TaglessMode {
		branchName = fmt.Sprintf("%s%d%s", pd.ImportBranchPrefix, pd.Version, pd.BranchSuffix)
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

	fileBytes, err := io.ReadAll(metadataFile)
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

				url := ""

				// We need to figure out the hashtype for templating purposes:
				hashType := "sha512"
				switch len(hash) {
				case 128:
					hashType = "sha512"
				case 64:
					hashType = "sha256"
				case 40:
					hashType = "sha1"
				case 32:
					hashType = "md5"
				}

				// need the name of the file without "SOURCES/":
				fileName := strings.Split(path, "/")[1]

				// Feed our template info to ProcessUrl and transform to the real values: ( {{.Name}}, {{.Branch}}, {{.Hash}}, {{.Hashtype}}, {{.Filename}} )
				url, hasTemplate := ProcessUrl(pd.CdnUrl, md.Name, branchName, hash, hashType, fileName)

				var req *http.Request
				var resp *http.Response

				// Download the --cdn-url given, but *only* if it contains template strings ( {{.Name}} , {{.Hash}} , etc. )
				// Otherwise we need to fall back to the traditional cdn-url patterns
				if hasTemplate {
					pd.Log.Printf("downloading %s", url)

					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						return fmt.Errorf("could not create new http request: %v", err)
					}
					req.Header.Set("Accept-Encoding", "*")

					resp, err = client.Do(req)
					if err != nil {
						return fmt.Errorf("could not download dist-git file: %v", err)
					}
				}

				// Default cdn-url:  If we don't have a templated download string, try the default <SITE>/<PKG>/<BRANCH>/<HASH> pattern:
				if resp == nil || resp.StatusCode != http.StatusOK {
					url = fmt.Sprintf("%s/%s/%s/%s", pd.CdnUrl, md.Name, branchName, hash)
					pd.Log.Printf("Attempting default URL: %s", url)
					req, err = http.NewRequest("GET", url, nil)
					if err != nil {
						return fmt.Errorf("could not create new http request: %v", err)
					}
					req.Header.Set("Accept-Encoding", "*")
					resp, err = client.Do(req)
					if err != nil {
						return fmt.Errorf("could not download dist-git file: %v", err)
					}
				}

				// If the default URL fails, we have one more pattern to try.  The simple <SITE>/<HASH> pattern
				// If this one fails, we are truly lost, and have to bail out w/ an error:
				if resp == nil || resp.StatusCode != http.StatusOK {
					url = fmt.Sprintf("%s/%s", pd.CdnUrl, hash)
					pd.Log.Printf("Attempting 2nd fallback URL: %s", url)
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

				body, err = io.ReadAll(resp.Body)
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

// Given a cdnUrl string as input, return same string, but with substituted
// template values ( {{.Name}} , {{.Hash}}, {{.Filename}}, etc. )
func ProcessUrl(cdnUrl string, name string, branch string, hash string, hashtype string, filename string) (string, bool) {
	tmpUrl := Lookaside{name, branch, hash, hashtype, filename}

	// Return cdnUrl as-is if we don't have any templates ("{{ .Variable }}") to process:
	if !(strings.Contains(cdnUrl, "{{") && strings.Contains(cdnUrl, "}}")) {
		return cdnUrl, false
	}

	// If we run into trouble with our template parsing, we'll just return the cdnUrl, exactly as we found it
	tmpl, err := template.New("").Parse(cdnUrl)
	if err != nil {
		return cdnUrl, false
	}

	var result bytes.Buffer
	err = tmpl.Execute(&result, tmpUrl)
	if err != nil {
		log.Fatalf("ERROR: Could not process CDN URL template(s) from URL string: %s\n", cdnUrl)
	}

	return result.String(), true

}
