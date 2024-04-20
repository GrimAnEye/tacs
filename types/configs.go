package types

import (
	l "log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		// Directory to search for templates
		// The search is also performed in subdirectories, but without following symbolic links
		// Required file extension: *.tmpl
		TemplateDir string `yaml:"templateDir"`

		Local *Local `yaml:"local"`
		Ldap  *Ldap  `yaml:"ldap"`
	}
	// Block for local processing of user logins.
	// Executes first because it doesn't require external requests.
	// If there is a match, subsequent stages are not checked
	Local struct {
		Default Params            `yaml:"default"`
		List    map[string]Params `yaml:"list"`
	}
	Params struct {
		// Name of the GO template, with insert fields -
		// `{{define "templateName"}}`
		Template string `yaml:"template"`

		// A key-value map (hash table) that is used to fill the template.
		// Overridden at user level. If the field is not present at the user level, it will be set to the specified value.
		// The key is a variable declared in the template - {{ .variableName }}
		// Value - what will be substituted for the key
		Fields map[string]string `yaml:"fields"`
	}

	Ldap struct {
		Uid                string       `yaml:"uid"`
		Filter             string       `yaml:"filter"`
		UsersSearchBase    string       `yaml:"usersSearchBase"`
		GroupsSearchBase   string       `yaml:"groupsSearchBase"`
		Subgroups          bool         `yaml:"subgroups"`
		AllowWithoutGroups bool         `yaml:"allowWithoutGroups"`
		Default            Params       `yaml:"default"`
		List               []LdapParams `yaml:"list"`
	}
	LdapParams struct {
		Group  string `yaml:"group"`
		Params `yaml:",inline"`
	}
)

// Contains a scheme for using templates from the configuration file
var Scheme = Config{}

// Merge - takes parameters from the source and
// copies them to the target if there are no corresponding fields.
//
// Used to supplement empty parameters with default values
func (to Params) Merge(from Params) Params {
	// temp_to := to

	if to.Template == "" && from.Template != "" {
		to.Template = from.Template
	}

	if to.Fields == nil {
		to.Fields = make(map[string]string)
	}

	for key := range from.Fields {
		if _, ok := to.Fields[key]; !ok {
			to.Fields[key] = from.Fields[key]
		}
	}
	return to
}

// Values -returns the values of the `Params.Fields` map.
//
// Used when compiling an array of fields to query in LDAP
// Ignores fields prefixed with `raw:`
func (p *Params) Values() []string {
	var out []string = make([]string, 0)
	for _, v := range p.Fields {
		if strings.HasPrefix(v, "raw:") {
			continue
		}
		out = append(out, v)
	}
	return out
}

// Values -returns the values of the `Params.Fields` map.
//
// Used when compiling an array of fields to query in LDAP
// Ignores fields prefixed without `raw:`
func (p *Params) RawValues() []string {
	var out []string = make([]string, 0)
	for _, v := range p.Fields {
		if !strings.HasPrefix(v, "raw:") {
			continue
		}
		out = append(out, strings.TrimPrefix(v, "raw:"))
	}
	return out
}

// Set - sets a `p.Fields` property
func (p *Params) Set(key, value string, isRaw bool) {
	if p.Fields == nil {
		p.Fields = make(map[string]string)
	}
	if isRaw {
		key = "raw:" + key
	}
	for k, v := range p.Fields {
		if v == key {
			p.Fields[k] = value
			break
		}
	}
}

// Load - loads the scheme file to the specified path.
//
// The presence of `TACS_SCHEME` in environment variables is required.
func (c *Config) Load() error {
	path := os.Getenv("TACS_SCHEME")

	if path != "" {
		f, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			l.Error("schema opening error",
				l.String("TACS_SCHEME", path),
				l.Any("err", err))
			return err
		}
		if err := yaml.Unmarshal(f, c); err != nil {
			l.Error("scheme unmarshal error",
				l.String("TACS_SCHEME", path),
				l.Any("err", err))
			return err
		}
	}
	return nil
}
