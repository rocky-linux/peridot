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
	"github.com/go-git/go-billy/v5"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"peridot.resf.org/publisher/updateinfo"
	"peridot.resf.org/secparse/db"
	secparsepb "peridot.resf.org/secparse/proto/v1"
	"peridot.resf.org/secparse/rpmutils"
	"strconv"
	"strings"
	"time"
)

type Scanner struct {
	DB db.Access
	FS billy.Filesystem
}

type internalAdvisory struct {
	Pb *secparsepb.Advisory
	Db *db.Advisory
}

type rpm struct {
	Name     string
	Src      string
	Sha256   string
	Epoch    string
	Advisory *internalAdvisory
}

func (s *Scanner) recursiveRPMScan(rootDir string) (map[string][]*rpm, error) {
	infos, err := s.FS.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	ret := map[string][]*rpm{}

	for _, fi := range infos {
		if fi.IsDir() {
			nRpms, err := s.recursiveRPMScan(filepath.Join(rootDir, fi.Name()))
			if err != nil {
				// Ignore paths we can't access
				continue
			}

			for k, v := range nRpms {
				if ret[k] == nil {
					ret[k] = []*rpm{}
				}

				ret[k] = append(ret[k], v...)
			}
		} else {
			if strings.HasSuffix(fi.Name(), ".rpm") {
				k := filepath.Join(rootDir, "..")
				if ret[k] == nil {
					ret[k] = []*rpm{}
				}

				ret[k] = append(ret[k], &rpm{
					Name: fi.Name(),
				})
			}
		}
	}

	return ret, nil
}

func (s *Scanner) findRepoData(rootDir string) (string, error) {
	if rootDir == "." {
		return "", errors.New("could not find repodata")
	}

	repoDataPath := filepath.Join(rootDir, "repodata")
	stat, err := s.FS.Stat(repoDataPath)
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

func (s *Scanner) ScanAndPublish(from string, composeName string, productName string, productShort string, republish bool) error {
	_, err := s.FS.Stat(composeName)
	if err != nil {
		return err
	}

	rpms, err := s.recursiveRPMScan(composeName)
	if err != nil {
		return err
	}

	published := map[string][]*rpm{}

	beginTx, err := s.DB.Begin()
	if err != nil {
		logrus.Errorf("Could not initiate tx: %v", err)
	}
	tx := s.DB.UseTransaction(beginTx)
	rollback := false

	advisories, err := tx.GetAllAdvisories(false)
	if err != nil {
		return err
	}
	for _, advisory := range advisories {
		// Skip already published advisories if republish is disabled
		if advisory.PublishedAt.Valid && !republish {
			continue
		}
		advisoryPb := db.DTOAdvisoryToPB(advisory)

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
						hasher := sha256.New()
						f, err := s.FS.Open(filepath.Join(repo, "Packages", repoRpm.Name))
						if err != nil {
							// If not found, then try sorted directory
							f, err = s.FS.Open(filepath.Join(repo, "Packages", strings.ToLower(string(repoRpm.Name[0])), repoRpm.Name))
							if err != nil {
								logrus.Errorf("Could not open affected package: %v", err)
								rollback = true
								break
							}
						}
						_, err = io.Copy(hasher, f)
						_ = f.Close()
						if err != nil {
							logrus.Errorf("Could not hash affected package: %v", err)
							rollback = true
							break
						}

						logrus.Infof("Advisory %s affects %s", advisoryPb.Name, artifact)
						err = tx.AddAdvisoryRPM(advisory.ID, artifact)
						if err != nil {
							logrus.Errorf("Could not add advisory RPM: %v", err)
							rollback = true
							break
						}
						touchedOnce = true
						repoRpm.Sha256 = hex.EncodeToString(hasher.Sum(nil))
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
				logrus.Errorf("Could not update advisory %s: %v", advisoryPb.Name, err)
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

		f, err := s.FS.Open(repoMdPath)
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

		var updateInfo *updateinfo.UpdatesRoot
		if olderUpdateInfo == "" {
			updateInfo = &updateinfo.UpdatesRoot{
				Updates: []*updateinfo.Update{},
			}
		} else {
			if republish {
				updateInfo = &updateinfo.UpdatesRoot{
					Updates: []*updateinfo.Update{},
				}
			} else {
				olderF, err := s.FS.Open(filepath.Join(repoDataDir, "..", olderUpdateInfo))
				if err != nil {
					logrus.Errorf("Could not open older updateinfo: %v", err)
					rollback = true
					break
				}

				var decoded bytes.Buffer
				r, err := gzip.NewReader(olderF)
				if err != nil {
					logrus.Errorf("Could not create new gzip reader: %v", err)
					rollback = true
					break
				}
				if _, err := io.Copy(&decoded, r); err != nil {
					logrus.Errorf("Could not copy gzip data: %v", err)
					rollback = true
					break
				}
				_ = r.Close()

				err = xml.NewDecoder(&decoded).Decode(&updateInfo)
				if err != nil {
					logrus.Errorf("Could not decode older updateinfo: %v", err)
					rollback = true
					break
				}
				_ = olderF.Close()

				if updateInfo.Updates == nil {
					updateInfo.Updates = []*updateinfo.Update{}
				}
			}
		}

		for advisoryName, publishedRpms := range advisories {
			advisory := advisoryByName[advisoryName]

			updateType := "enhancement"
			switch advisory.Pb.Type {
			case secparsepb.Advisory_BugFix:
				updateType = "bugfix"
				break
			case secparsepb.Advisory_Security:
				updateType = "security"
				break
			}

			severity := advisory.Pb.Severity.String()
			if advisory.Pb.Severity == secparsepb.Advisory_UnknownSeverity {
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
				Rights:      "Copyright (C) 2021 Rocky Enterprise Software Foundation",
				Release:     productName,
				PushCount:   "1",
				Severity:    severity,
				Summary:     advisory.Pb.Topic,
				Description: fmt.Sprintf("For more information visit https://errata.rockylinux.org/%s", advisory.Pb.Name),
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
				sourceCve := strings.Split(cve, ":::")
				sourceBy := sourceCve[0]
				sourceLink := sourceCve[1]
				id := sourceCve[2]

				referenceType := "erratum"
				if strings.HasPrefix(id, "CVE") {
					referenceType = "cve"
				}

				reference := &updateinfo.UpdateReference{
					Href:  sourceLink,
					ID:    id,
					Type:  referenceType,
					Title: fmt.Sprintf("Update information for %s is retrieved from %s", id, sourceBy),
				}

				update.References.References = append(update.References.References, reference)
			}

			for _, publishedRpm := range publishedRpms {
				nvr := rpmutils.NVR().FindStringSubmatch(publishedRpm.Name)

				update.PkgList.Collections[0].Packages = append(update.PkgList.Collections[0].Packages, &updateinfo.UpdatePackage{
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
				})
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

		uif, err := s.FS.OpenFile(updateInfoPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
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

		updateF, err := s.FS.OpenFile(repoMdPath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
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
			_ = s.FS.Remove(filepath.Join(repo, olderUpdateInfo))
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
