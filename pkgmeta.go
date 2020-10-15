package makeaddon

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// MetaData represents the content of an addon's pkgmeta file.
type MetaData struct {
	PackageAs       string              `yaml:"package-as"`
	ManualChangelog *Changelog          `yaml:"manual-changelog"`
	Externals       map[string]External `yaml:"externals"`
	MoveFolders     map[string]string   `yaml:"move-folders"`
	Ignore          []string            `yaml:"ignore"`
	LicenseOutput   string              `yaml:"license-output"`
	NoLibCreation   bool                `yaml:"enable-nolib-creation"`
}

// External represents a configured external resource - a VCS url and optionally a tag.
type External struct {
	Url string `yaml:"url"`
	Tag string `yaml:"tag"`
}

func (e *External) UnmarshalYAML(unmarshal func(interface{}) error) error {
	str := ""
	if err := unmarshal(&str); err == nil {
		e.Url = str
		return nil
	}

	// Cast to a type that doesn't implement yaml.Unmarshaler and carry on.
	type bare External
	return unmarshal((*bare)(e))
}

// Changelog represents the configuration for a manual changelog.
type Changelog struct {
	Filename   string `yaml:"filename"`
	MarkupType string `yaml:"markup-type"`
}

func (m *Changelog) UnmarshalYAML(unmarshal func(interface{}) error) error {
	str := ""
	if err := unmarshal(&str); err == nil {
		m.Filename = str
		m.MarkupType = "plain"
		return nil
	}

	// Cast to a type that doesn't implement yaml.Unmarshaler and carry on.
	type bare Changelog
	return unmarshal((*bare)(m))
}

// ReadMetaData reads package metadata from the given reader.
func ReadMetaData(reader io.Reader) (*MetaData, error) {
	data := MetaData{
		NoLibCreation: true,
	}
	return &data, yaml.NewDecoder(reader).Decode(&data)
}

// MetaDataFromDirectory scans the given directory to find a relevant metadata file and returns a
// processed representation.
func MetaDataFromDirectory(dir string) (*MetaData, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for i := range files {
		if !files[i].IsDir() && isFile(files[i].Name()) {
			f, err := os.Open(filepath.Join(dir, files[i].Name()))
			if err != nil {
				return nil, err
			}

			//goland:noinspection GoDeferInLoop
			defer f.Close()
			return ReadMetaData(f)
		}
	}

	return nil, fmt.Errorf("pkgmeta file not found in path %s", dir)
}

func isFile(name string) bool {
	return strings.EqualFold(name, ".pkgmeta") ||
		strings.EqualFold(name, "pkgmeta.yaml") ||
		strings.EqualFold(name, "pkgmeta.yml")
}
