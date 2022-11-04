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

package modulemd

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	"gopkg.in/yaml.v3"
)

type ServiceLevelType string

const (
	ServiceLevelRawhide       ServiceLevelType = "rawhide"
	ServiceLevelStableAPI     ServiceLevelType = "stable_api"
	ServiceLevelBugFixes      ServiceLevelType = "bug_fixes"
	ServiceLevelSecurityFixes ServiceLevelType = "security_fixes"
)

type ServiceLevel struct {
	Eol string `yaml:"eol,omitempty"`
}

type License struct {
	Module  []string `yaml:"module,omitempty"`
	Content []string `yaml:"content,omitempty"`
}

type Dependencies struct {
	BuildRequires map[string][]string `yaml:"buildrequires,omitempty,omitempty"`
	Requires      map[string][]string `yaml:"requires,omitempty,omitempty"`
}

type References struct {
	Community     string `yaml:"community,omitempty"`
	Documentation string `yaml:"documentation,omitempty"`
	Tracker       string `yaml:"tracker,omitempty"`
}

type Profile struct {
	Description string   `yaml:"description,omitempty"`
	Rpms        []string `yaml:"rpms,omitempty"`
}

type API struct {
	Rpms []string `yaml:"rpms,omitempty"`
}

type BuildOptsRPM struct {
	Macros    string   `yaml:"macros,omitempty"`
	Whitelist []string `yaml:"whitelist,omitempty"`
}

type BuildOpts struct {
	Rpms   *BuildOptsRPM `yaml:"rpms,omitempty"`
	Arches []string      `yaml:"arches,omitempty"`
}

type ComponentRPM struct {
	Name          string   `yaml:"name,omitempty"`
	Rationale     string   `yaml:"rationale,omitempty"`
	Repository    string   `yaml:"repository,omitempty"`
	Cache         string   `yaml:"cache,omitempty"`
	Ref           string   `yaml:"ref,omitempty"`
	Buildonly     bool     `yaml:"buildonly,omitempty"`
	Buildroot     bool     `yaml:"buildroot,omitempty"`
	SrpmBuildroot bool     `yaml:"srpm-buildroot,omitempty"`
	Buildorder    int      `yaml:"buildorder,omitempty"`
	Arches        []string `yaml:"arches,omitempty"`
	Multilib      []string `yaml:"multilib,omitempty"`
}

type ComponentModule struct {
	Rationale  string `yaml:"rationale,omitempty"`
	Repository string `yaml:"repository,omitempty"`
	Ref        string `yaml:"ref,omitempty"`
	Buildorder int    `yaml:"buildorder,omitempty"`
}

type Components struct {
	Rpms    map[string]*ComponentRPM    `yaml:"rpms,omitempty"`
	Modules map[string]*ComponentModule `yaml:"modules,omitempty"`
}

type ArtifactsRPMMap struct {
	Name    string  `yaml:"name,omitempty"`
	Epoch   int     `yaml:"epoch,omitempty"`
	Version float64 `yaml:"version,omitempty"`
	Release string  `yaml:"release,omitempty"`
	Arch    string  `yaml:"arch,omitempty"`
	Nevra   string  `yaml:"nevra,omitempty"`
}

type Artifacts struct {
	Rpms   []string                               `yaml:"rpms,omitempty"`
	RpmMap map[string]map[string]*ArtifactsRPMMap `yaml:"rpm-map,omitempty"`
}

type Data struct {
	Name          string                             `yaml:"name,omitempty"`
	Stream        string                             `yaml:"stream,omitempty"`
	Version       string                             `yaml:"version,omitempty"`
	StaticContext bool                               `yaml:"static_context,omitempty"`
	Context       string                             `yaml:"context,omitempty"`
	Arch          string                             `yaml:"arch,omitempty"`
	Summary       string                             `yaml:"summary,omitempty"`
	Description   string                             `yaml:"description,omitempty"`
	ServiceLevels map[ServiceLevelType]*ServiceLevel `yaml:"servicelevels,omitempty"`
	License       *License                           `yaml:"license,omitempty"`
	Xmd           map[string]map[string]string       `yaml:"xmd,omitempty"`
	Dependencies  []*Dependencies                    `yaml:"dependencies,omitempty"`
	References    *References                        `yaml:"references,omitempty"`
	Profiles      map[string]*Profile                `yaml:"profiles,omitempty"`
	Profile       map[string]*Profile                `yaml:"profile,omitempty"`
	API           *API                               `yaml:"api,omitempty"`
	Filter        *API                               `yaml:"filter,omitempty"`
	BuildOpts     *BuildOpts                         `yaml:"buildopts,omitempty"`
	Components    *Components                        `yaml:"components,omitempty"`
	Artifacts     *Artifacts                         `yaml:"artifacts,omitempty"`
}

type ModuleMd struct {
	Document string `yaml:"document,omitempty"`
	Version  int    `yaml:"version,omitempty"`
	Data     *Data  `yaml:"data,omitempty"`
}

type DetectVersionDocument struct {
	Document string `yaml:"document,omitempty"`
	Version  int    `yaml:"version,omitempty"`
}

type DefaultsData struct {
	Module   string              `yaml:"module,omitempty"`
	Stream   string              `yaml:"stream,omitempty"`
	Profiles map[string][]string `yaml:"profiles,omitempty"`
}

type Defaults struct {
	Document string        `yaml:"document,omitempty"`
	Version  int           `yaml:"version,omitempty"`
	Data     *DefaultsData `yaml:"data,omitempty"`
}

func Parse(input []byte) (*ModuleMd, error) {
	var detect DetectVersionDocument
	err := yaml.Unmarshal(input, &detect)
	if err != nil {
		return nil, fmt.Errorf("error detecting document version: %s", err)
	}

	var ret ModuleMd

	if detect.Version == 2 {
		err = yaml.Unmarshal(input, &ret)
		if err != nil {
			return nil, fmt.Errorf("error parsing modulemd: %s", err)
		}
	} else if detect.Version == 3 {
		var v3 V3
		err = yaml.Unmarshal(input, &v3)
		if err != nil {
			return nil, fmt.Errorf("error parsing modulemd: %s", err)
		}

		ret = ModuleMd{
			Document: v3.Document,
			Version:  v3.Version,
			Data: &Data{
				Name:        v3.Data.Name,
				Stream:      v3.Data.Stream,
				Summary:     v3.Data.Summary,
				Description: v3.Data.Description,
				License: &License{
					Module: v3.Data.License,
				},
				Xmd:        v3.Data.Xmd,
				References: v3.Data.References,
				Profiles:   v3.Data.Profiles,
				Profile:    v3.Data.Profile,
				API:        v3.Data.API,
				Filter:     v3.Data.Filter,
				BuildOpts: &BuildOpts{
					Rpms:   v3.Data.Configurations[0].BuildOpts.Rpms,
					Arches: v3.Data.Configurations[0].BuildOpts.Arches,
				},
				Components: v3.Data.Components,
			},
		}
	}

	return &ret, nil
}

func (m *ModuleMd) Marshal(fs billy.Filesystem, path string) error {
	bts, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	_ = fs.Remove(path)
	f, err := fs.Create(path)
	if err != nil {
		return err
	}
	_, err = f.Write(bts)
	if err != nil {
		return err
	}
	_ = f.Close()

	return nil
}
