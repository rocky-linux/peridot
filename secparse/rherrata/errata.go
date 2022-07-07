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

package rherrata

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	secparsepb "peridot.resf.org/secparse/proto/v1"
	"strings"
	"time"
)

type Architecture string

const (
	ArchX8664   Architecture = "x86_64"
	ArchAArch64 Architecture = "aarch64"
	ArchPPC64   Architecture = "ppc64le"
	ArchS390X   Architecture = "s390x"
	ArchNoArch  Architecture = "noarch"
)

type Fix struct {
	BugzillaID  string
	Description string
}

type UpdatedPackages struct {
	SRPMs    []string
	Packages map[Architecture][]string
}

type Errata struct {
	Synopsis         string
	Type             secparsepb.Advisory_Type
	Severity         secparsepb.Advisory_Severity
	Topic            []string
	Description      []string
	Solution         []string
	AffectedProducts map[string]*UpdatedPackages
	Fixes            []*Fix
	CVEs             []string
	References       []string
	IssuedAt         time.Time
}

func (a *API) GetErrata(advisory string) (*Errata, error) {
	var err error
	var errata Errata
	c := colly.NewCollector(colly.UserAgent(a.userAgent))

	// Do not fix this typo. It is like this on Red Hat's website
	c.OnHTML("div#synpopsis", func(element *colly.HTMLElement) {
		errata.Synopsis = element.DOM.Find("p").Text()
	})
	c.OnHTML("div#topic > p", func(element *colly.HTMLElement) {
		errata.Topic = append(errata.Topic, element.Text)
	})
	c.OnHTML("div#solution > p", func(element *colly.HTMLElement) {
		errata.Solution = append(errata.Solution, element.Text)
	})
	c.OnHTML("div#fixes > ul > li", func(element *colly.HTMLElement) {
		fixComponents := strings.SplitN(element.Text, "-", 3)
		if len(fixComponents) != 3 {
			return
		}

		for i, comp := range fixComponents {
			fixComponents[i] = strings.TrimSpace(comp)
		}

		fix := &Fix{
			BugzillaID:  fixComponents[1],
			Description: fixComponents[2],
		}
		errata.Fixes = append(errata.Fixes, fix)
	})
	c.OnHTML("div#cves > ul > li", func(element *colly.HTMLElement) {
		errata.CVEs = append(errata.CVEs, strings.TrimSpace(element.Text))
	})
	c.OnHTML("div#references > ul > li", func(element *colly.HTMLElement) {
		errata.References = append(errata.References, strings.TrimSpace(element.Text))
	})
	c.OnHTML("dl.details", func(element *colly.HTMLElement) {
		issuedAt, err := time.Parse("2006-01-02", element.DOM.Find("dd").First().Text())
		if err == nil {
			errata.IssuedAt = issuedAt
		}
	})
	c.OnHTML("div#packages", func(element *colly.HTMLElement) {
		productIndex := map[int]string{}
		products := map[string]*UpdatedPackages{}
		element.DOM.Find("h2").Each(func(i int, selection *goquery.Selection) {
			productIndex[i] = selection.Text()
			products[selection.Text()] = &UpdatedPackages{}
		})

		element.DOM.Find("table.files").Each(func(i int, selection *goquery.Selection) {
			productUpdate := products[productIndex[i]]
			if productUpdate.Packages == nil {
				productUpdate.Packages = map[Architecture][]string{}
			}

			selection.Find("td.name").Each(func(_ int, selection *goquery.Selection) {
				name := strings.TrimSpace(selection.Text())
				isRpm := strings.HasSuffix(name, ".rpm")
				isSrcRpm := strings.HasSuffix(name, ".src.rpm")
				if isRpm {
					if isSrcRpm {
						productUpdate.SRPMs = append(productUpdate.SRPMs, name)
					} else {
						var arch Architecture
						if strings.Contains(name, ".x86_64") || strings.Contains(name, ".i686") {
							arch = ArchX8664
						} else if strings.Contains(name, ".aarch64") {
							arch = ArchAArch64
						} else if strings.Contains(name, ".ppc64le") {
							arch = ArchPPC64
						} else if strings.Contains(name, ".s390x") {
							arch = ArchS390X
						} else if strings.Contains(name, ".noarch") {
							arch = ArchNoArch
						}

						if productUpdate.Packages[arch] == nil {
							productUpdate.Packages[arch] = []string{}
						}
						productUpdate.Packages[arch] = append(productUpdate.Packages[arch], name)
					}
				}
			})

			errata.AffectedProducts = products
		})
	})
	c.OnHTML("div#description > p", func(element *colly.HTMLElement) {
		htmlText, err := element.DOM.Html()
		if err != nil {
			return
		}
		htmlText = strings.TrimSuffix(htmlText, "<br/>")

		if element.Text == "Security Fix(es):" || element.Text == "Bug Fix(es) and Enhancement(s):" || element.Text == "Bug Fix(es):" || element.Text == "Enhancement(s):" {
			return
		}
		errata.Description = append(errata.Description, strings.Split(htmlText, "<br/>")...)
	})
	c.OnHTML("div#type-severity", func(element *colly.HTMLElement) {
		typeSeverity := strings.Split(element.DOM.Find("p").Text(), ":")
		if typeSeverity[0] == "Product Enhancement Advisory" {
			errata.Type = secparsepb.Advisory_Enhancement
		} else if typeSeverity[0] == "Bug Fix Advisory" {
			errata.Type = secparsepb.Advisory_BugFix
		} else {
			if len(typeSeverity) != 2 {
				err = errors.New("invalid type/severity")
				return
			}

			typeSplit := strings.Split(typeSeverity[0], " ")
			if len(typeSplit) != 2 {
				err = errors.New("invalid type")
				return
			}

			switch strings.TrimSpace(typeSplit[0]) {
			case "Security":
				errata.Type = secparsepb.Advisory_Security
				break
			case "BugFix":
				errata.Type = secparsepb.Advisory_BugFix
				break
			case "Enhancement":
				errata.Type = secparsepb.Advisory_Enhancement
				break
			}

			switch strings.TrimSpace(typeSeverity[1]) {
			case "Low":
				errata.Severity = secparsepb.Advisory_Low
				break
			case "Moderate":
				errata.Severity = secparsepb.Advisory_Moderate
				break
			case "Important":
				errata.Severity = secparsepb.Advisory_Important
				break
			case "Critical":
				errata.Severity = secparsepb.Advisory_Critical
				break
			}
		}
	})

	errC := c.Visit(fmt.Sprintf("%s/%s", a.baseURLErrata, advisory))
	if errC != nil {
		return nil, errC
	}

	c.Wait()

	if err != nil {
		return nil, err
	}
	return &errata, nil
}
