package modulemd

type V3 struct {
	Document string  `yaml:"document,omitempty"`
	Version  int     `yaml:"version,omitempty"`
	Data     *V3Data `yaml:"data,omitempty"`
}

type Configurations struct {
	Context       string              `yaml:"context,omitempty"`
	Platform      string              `yaml:"platform,omitempty"`
	BuildRequires map[string][]string `yaml:"buildrequires,omitempty"`
	Requires      map[string][]string `yaml:"requires,omitempty"`
	BuildOpts     *BuildOpts          `yaml:"buildopts,omitempty"`
}

type V3Data struct {
	Name           string                       `yaml:"name,omitempty"`
	Stream         string                       `yaml:"stream,omitempty"`
	Summary        string                       `yaml:"summary,omitempty"`
	Description    string                       `yaml:"description,omitempty"`
	License        []string                     `yaml:"license,omitempty"`
	Xmd            map[string]map[string]string `yaml:"xmd,omitempty"`
	Configurations []*Configurations            `yaml:"configurations,omitempty"`
	References     *References                  `yaml:"references,omitempty"`
	Profiles       map[string]*Profile          `yaml:"profiles,omitempty"`
	Profile        map[string]*Profile          `yaml:"profile,omitempty"`
	API            *API                         `yaml:"api,omitempty"`
	Filter         *API                         `yaml:"filter,omitempty"`
	Demodularized  *API                         `yaml:"demodularized,omitempty"`
	Components     *Components                  `yaml:"components,omitempty"`
}
