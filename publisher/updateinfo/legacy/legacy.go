// Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
// Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
// Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	apollodb "peridot.resf.org/apollo/db"
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/publisher/updateinfo"
	"peridot.resf.org/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Scanner struct {
	DB apollodb.Access
}

type internalAdvisory struct {
	Pb *apollopb.Advisory
	Db *apollodb.Advisory
}

type rpm struct {
	Name     string
	Src      string
	Sha256   string
	Epoch    string
	Repo     string
	Err      error
	Advisory *internalAdvisory
}

func (s *Scanner) recursiveRPMScan(rootDir string, cache map[string]string) (<-chan rpm, <-chan error) {
	res := make(chan rpm)
	errc := make(chan error, 1)

	go func() {
		var wg sync.WaitGroup
		err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".rpm") {
				return nil
			}
			if strings.Contains(path, "kickstart/Packages") {
				return nil
			}

			wg.Add(1)
			go func() {
				k, err := s.findRepoData(filepath.Join(path, ".."))
				if err != nil {
					logrus.Errorf("could not find repodata for %s: %s", path, err)
					k = filepath.Join(path, "..")
				}
				k = filepath.Join(k, "..")

				var sum string
				if s := cache[d.Name()]; s != "" {
					sum = s
				} else {
					f, _ := os.Open(path)
					defer f.Close()
					hasher := sha256.New()
					_, err = io.Copy(hasher, f)
					sum = hex.EncodeToString(hasher.Sum(nil))
				}

				select {
				case res <- rpm{
					Name:   d.Name(),
					Sha256: sum,
					Repo:   k,
					Err:    err,
				}:
				}

				wg.Done()
			}()

			select {
			default:
				return nil
			}
		})
		go func() {
			wg.Wait()
			close(res)
		}()
		errc <- err
	}()

	return res, errc
}

func (s *Scanner) findRepoData(rootDir string) (string, error) {
	if rootDir == "." {
		return "", errors.New("could not find repodata")
	}

	repoDataPath := filepath.Join(rootDir, "repodata")
	stat, err := os.Stat(repoDataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.findRepoData(filepath.Join(rootDir, ".."))
		} else {
			return "", err
		}
	}

	if stat.IsDir() {
		return repoDataPath, nil
	} else {
		return s.findRepoData(filepath.Join(rootDir, ".."))
	}
}

func (s *Scanner) ScanAndPublish(from string, composeName string, productName string, productShort string, productID int64, scanAndStop bool) error {
	logrus.Infof("using %s as root directory", composeName)

	realPathCompose, err := filepath.EvalSymlinks(composeName)
	if err != nil {
		return err
	}

	logrus.Infof("real path is %s", realPathCompose)

	_, err = os.Stat(realPathCompose)
	if err != nil {
		return fmt.Errorf("could not find compose %s: %w", realPathCompose, err)
	}

	// Read cache file if exists, so we can skip hashing on known artifacts
	cacheFile := filepath.Join(realPathCompose, fmt.Sprintf("apollocache_%d", productID))
	cache := map[string]string{}
	if _, err := os.Stat(cacheFile); err == nil {
		cacheBts, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			return err
		}
		cacheLines := strings.Split(string(cacheBts), "\n")
		for _, line := range cacheLines {
			if line == "" {
				continue
			}
			parts := strings.Split(line, " ")
			cache[parts[0]] = parts[1]
		}
	}

	rpms := map[string][]*rpm{}
	rpmsChan, errChan := s.recursiveRPMScan(realPathCompose, cache)
	for r := range rpmsChan {
		rpmCopy := r
		if rpmCopy.Err != nil {
			return rpmCopy.Err
		}

		if rpms[rpmCopy.Repo] == nil {
			rpms[rpmCopy.Repo] = []*rpm{}
		}

		rpms[rpmCopy.Repo] = append(rpms[rpmCopy.Repo], &rpmCopy)
	}
	if err := <-errChan; err != nil {
		return err
	}

	if len(rpms) == 0 {
		return errors.New("no rpms found")
	}

	// Cache hashes in {REPO_DIR}/apollocache_{PRODUCT_ID}
	var newCacheEntries []string
	for _, v := range rpms {
		for _, rpm := range v {
			entry := fmt.Sprintf("%s %s", rpm.Name, rpm.Sha256)
			if !utils.StrContains(entry, newCacheEntries) {
				newCacheEntries = append(newCacheEntries, entry)
			}
		}
	}
	if err := ioutil.WriteFile(cacheFile, []byte(strings.Join(newCacheEntries, "\n")), 0644); err != nil {
		return err
	}

	if scanAndStop {
		for k := range rpms {
			logrus.Infof("repo %s", k)
		}
		return nil
	}

	published := map[string][]*rpm{}

	beginTx, err := s.DB.Begin()
	if err != nil {
		logrus.Errorf("Could not initiate tx: %v", err)
	}
	tx := s.DB.UseTransaction(beginTx)
	rollback := false

	advisories, err := tx.GetAllAdvisories(&apollopb.AdvisoryFilters{
		IncludeUnpublished: wrapperspb.Bool(true),
	}, 0, -1)
	if err != nil {
		return err
	}
	for _, advisory := range advisories {
		advisoryPb := apollodb.DTOAdvisoryToPB(advisory)

		touchedOnce := false
		for _, artifactWithSrpm := range advisory.BuildArtifacts {
			artifactSplit := strings.Split(artifactWithSrpm, ":::")
			artifact := artifactSplit[0]
			artifactSrc := rpmutils.Epoch().ReplaceAllString(artifactSplit[1], "")

			for repo, repoRpms := range rpms {
				if strings.HasSuffix(repo, "/Packages") {
					repo = strings.TrimSuffix(repo, "/Packages")
				}
				if published[repo] == nil {
					published[repo] = []*rpm{}
				}

				for _, repoRpm := range repoRpms {
					if repoRpm.Name == rpmutils.Epoch().ReplaceAllString(artifact, "") {
						logrus.Infof("Advisory %s affects %s", advisoryPb.Name, artifact)
						err = tx.AddAdvisoryRPM(advisory.ID, artifact, productID)
						if err != nil {
							logrus.Errorf("Could not add advisory RPM: %v", err)
							rollback = true
							break
						}
						touchedOnce = true
						repoRpm.Epoch = strings.TrimSuffix(rpmutils.Epoch().FindStringSubmatch(artifact)[0], ":")
						repoRpm.Advisory = &internalAdvisory{
							Pb: advisoryPb,
							Db: advisory,
						}
						repoRpm.Src = artifactSrc
						published[repo] = append(published[repo], repoRpm)
					}
				}
			}
		}
		if rollback {
			break
		}
		if !touchedOnce {
			continue
		}

		if !advisory.PublishedAt.Valid {
			advisory.PublishedAt = sql.NullTime{Valid: true, Time: time.Now()}
			_, err = tx.UpdateAdvisory(advisory)
			if err != nil {
				logrus.Errorf("could not update advisory %s: %v", advisoryPb.Name, err)
				rollback = true
				break
			}
		}
	}

	publishedMappedByAdvisory := map[string]map[string][]*rpm{}
	advisoryByName := map[string]*internalAdvisory{}

	for repo, publishedRpms := range published {
		if publishedMappedByAdvisory[repo] == nil {
			publishedMappedByAdvisory[repo] = map[string][]*rpm{}
		}

		for _, publishedRpm := range publishedRpms {
			if publishedMappedByAdvisory[repo][publishedRpm.Advisory.Pb.Name] == nil {
				publishedMappedByAdvisory[repo][publishedRpm.Advisory.Pb.Name] = []*rpm{}
			}
			if advisoryByName[publishedRpm.Advisory.Pb.Name] == nil {
				advisoryByName[publishedRpm.Advisory.Pb.Name] = publishedRpm.Advisory
			}
			publishedMappedByAdvisory[repo][publishedRpm.Advisory.Pb.Name] = append(publishedMappedByAdvisory[repo][publishedRpm.Advisory.Pb.Name], publishedRpm)
		}
	}

	for repo, advisories := range publishedMappedByAdvisory {
		repoDataDir, err := s.findRepoData(repo)
		if err != nil {
			logrus.Error(err)
			rollback = true
			break
		}
		repoMdPath := filepath.Join(repoDataDir, "repomd.xml")

		f, err := os.Open(repoMdPath)
		if err != nil {
			logrus.Errorf("Could not open repomd.xml: %v", err)
			rollback = true
			break
		}

		var repomd updateinfo.RepoMdRoot
		err = xml.NewDecoder(f).Decode(&repomd)
		if err != nil {
			logrus.Errorf("Could not decode repomd: %v", err)
			rollback = true
			break
		}

		_ = f.Close()

		var olderUpdateInfo string
		for _, e := range repomd.Data {
			if e.Type == "updateinfo" {
				olderUpdateInfo = e.Location.Href
			}
		}

		updateInfo := &updateinfo.UpdatesRoot{
			Updates: []*updateinfo.Update{},
		}

		for advisoryName, publishedRpms := range advisories {
			advisory := advisoryByName[advisoryName]

			updateType := "enhancement"
			switch advisory.Pb.Type {
			case apollopb.Advisory_TYPE_BUGFIX:
				updateType = "bugfix"
				break
			case apollopb.Advisory_TYPE_SECURITY:
				updateType = "security"
				break
			}

			severity := advisory.Pb.Severity.String()
			if advisory.Pb.Severity == apollopb.Advisory_SEVERITY_UNKNOWN {
				severity = "None"
			}

			update := &updateinfo.Update{
				From:    from,
				Status:  "final",
				Type:    updateType,
				Version: "2",
				ID:      advisory.Pb.Name,
				Title:   advisory.Pb.Synopsis,
				Issued: &updateinfo.UpdateDate{
					Date: advisory.Db.PublishedAt.Time.Format(updateinfo.TimeFormat),
				},
				Updated: &updateinfo.UpdateDate{
					Date: advisory.Db.RedHatIssuedAt.Time.Format(updateinfo.TimeFormat),
				},
				Rights:      "Copyright (C) 2022 Rocky Enterprise Software Foundation",
				Release:     productName,
				PushCount:   "1",
				Severity:    severity,
				Summary:     advisory.Pb.Topic,
				Description: advisory.Pb.Description,
				References: &updateinfo.UpdateReferenceRoot{
					References: []*updateinfo.UpdateReference{},
				},
				PkgList: &updateinfo.UpdateCollectionRoot{
					Collections: []*updateinfo.UpdateCollection{
						{
							Short:    productShort,
							Name:     productName,
							Packages: []*updateinfo.UpdatePackage{},
						},
					},
				},
			}

			for _, cve := range advisory.Pb.Cves {
				sourceBy := cve.SourceBy
				sourceLink := cve.SourceLink
				id := cve.Name

				referenceType := "erratum"
				if strings.HasPrefix(id, "CVE") {
					referenceType = "cve"
				}

				reference := &updateinfo.UpdateReference{
					Href:  sourceLink.Value,
					ID:    id,
					Type:  referenceType,
					Title: fmt.Sprintf("Update information for %s is retrieved from %s", id, sourceBy.Value),
				}

				update.References.References = append(update.References.References, reference)
			}

			for _, publishedRpm := range publishedRpms {
				nvr := rpmutils.NVR().FindStringSubmatch(publishedRpm.Name)

				updPkg := &updateinfo.UpdatePackage{
					Name:     nvr[1],
					Version:  nvr[2],
					Release:  nvr[3],
					Epoch:    publishedRpm.Epoch,
					Arch:     nvr[4],
					Src:      publishedRpm.Src,
					Filename: publishedRpm.Name,
					Sum: []*updateinfo.UpdatePackageSum{
						{
							Type:  "sha256",
							Value: publishedRpm.Sha256,
						},
					},
				}
				if advisory.Db.RebootSuggested {
					updPkg.RebootSuggested = "True"
				}
				update.PkgList.Collections[0].Packages = append(update.PkgList.Collections[0].Packages, updPkg)
			}
			if rollback {
				break
			}

			updateInfo.Updates = append(updateInfo.Updates, update)
		}
		if rollback {
			break
		}

		xmlBytes, err := xml.MarshalIndent(updateInfo, "", "  ")
		if err != nil {
			logrus.Errorf("Could not encode updateinfo xml: %v", err)
			rollback = true
			break
		}

		hasher := sha256.New()

		openSize := len(xmlBytes)
		_, err = hasher.Write(xmlBytes)
		if err != nil {
			logrus.Errorf("Could not hash updateinfo: %v", err)
			rollback = true
			break
		}
		openChecksum := hex.EncodeToString(hasher.Sum(nil))
		hasher.Reset()

		var gzippedBuf bytes.Buffer
		w := gzip.NewWriter(&gzippedBuf)
		_, err = w.Write(xmlBytes)
		if err != nil {
			logrus.Errorf("Could not gzip encode: %v", err)
			rollback = true
			break
		}
		_ = w.Close()

		closedSize := len(gzippedBuf.Bytes())
		_, err = hasher.Write(gzippedBuf.Bytes())
		if err != nil {
			logrus.Errorf("Could not hash gzipped: %v", err)
			rollback = true
			break
		}
		closedChecksum := hex.EncodeToString(hasher.Sum(nil))
		hasher.Reset()

		timestamp := strconv.FormatInt(time.Now().Unix(), 10)

		updateInfoPath := filepath.Join(repoDataDir, fmt.Sprintf("%s-updateinfo.xml.gz", closedChecksum))
		updateInfoEntry := &updateinfo.RepoMdData{
			Type: "updateinfo",
			Checksum: &updateinfo.RepoMdDataChecksum{
				Type:  "sha256",
				Value: closedChecksum,
			},
			OpenChecksum: &updateinfo.RepoMdDataChecksum{
				Type:  "sha256",
				Value: openChecksum,
			},
			Location: &updateinfo.RepoMdDataLocation{
				Href: strings.ReplaceAll(updateInfoPath, repo+"/", ""),
			},
			Timestamp: timestamp,
			Size:      strconv.Itoa(closedSize),
			OpenSize:  strconv.Itoa(openSize),
		}

		if olderUpdateInfo == "" {
			repomd.Data = append(repomd.Data, updateInfoEntry)
		} else {
			for i, e := range repomd.Data {
				if e.Type == "updateinfo" {
					repomd.Data[i] = updateInfoEntry
				}
			}
		}

		uif, err := os.OpenFile(updateInfoPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			logrus.Errorf("Could not open updateinfo file %s: %v", updateInfoPath, err)
			rollback = true
			break
		}
		_, err = uif.Write(gzippedBuf.Bytes())
		if err != nil {
			logrus.Errorf("Could not write gzipped updateinfo file: %v", err)
			rollback = true
			break
		}
		_ = uif.Close()

		if repomd.Rpm != "" && repomd.XmlnsRpm == "" {
			repomd.XmlnsRpm = repomd.Rpm
			repomd.Rpm = ""
		}

		updateF, err := os.OpenFile(repoMdPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			logrus.Errorf("Could not open repomd file for update: %v", err)
			rollback = true
			break
		}
		_, _ = updateF.Write([]byte(xml.Header))
		enc := xml.NewEncoder(updateF)
		enc.Indent("", "  ")
		err = enc.Encode(&repomd)
		if err != nil {
			logrus.Errorf("Could not encode updated repomd file: %v", err)
			rollback = true
			break
		}
		_ = updateF.Close()

		if olderUpdateInfo != "" {
			_ = os.Remove(filepath.Join(repo, olderUpdateInfo))
		}
	}

	if rollback {
		err := beginTx.Rollback()
		if err != nil {
			logrus.Errorf("Could not rollback: %v", err)
		}

		return errors.New("rolled back")
	}

	err = beginTx.Commit()
	if err != nil {
		logrus.Errorf("Could not commit transaction: %v", err)
	}

	return nil
}
