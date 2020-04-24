package config

import (
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/utrack/pbtree/app"
	"github.com/utrack/pbtree/fetcher"
	"gopkg.in/yaml.v3"
)

// Config is a model for .pbtree.yaml.
type Config struct {
	// Replace <import1> with <import2>
	Replace map[string]string `yaml:"replace"`

	VendoredForeigns []string `yaml:"vendor"`

	// Paths to local protofiles or their directories
	// that should be added to the tree
	Paths []string `yaml:"paths"`

	// Output controls where to put the resulting tree.
	Output string `yaml:"output"`

	// RepoModuleName is current repo's name.
	RepoModuleName string `yaml:"moduleName"`

	// RepoToBranch maps repositories to desired branches.
	RepoToBranch map[string]string `yaml:"branches"`

	Fetchers Fetchers `yaml:"fetchers"`
}

type Fetchers struct {
	HTTP FetcherHTTP `yaml:"http"`
}

type FetcherHTTP struct {
	ModuleToAddr map[string]string `yaml:"repoToAddress"`
}

func FromFile(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("path to config file is empty")
	}
	r, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "opening config")
	}
	var c Config
	err = yaml.Unmarshal(r, &c)
	return &c, errors.Wrap(err, "reading config")
}

func Default(repoName string) Config {
	return Config{
		RepoModuleName: repoName,
		Output:         "vendor.pbtree",
		Replace: map[string]string{
			"google/api/*":      "github.com/googleapis/googleapis!/google/api/*",
			"google/type/*":     "github.com/googleapis/googleapis!/google/type/*",
			"google/rpc/*":      "github.com/googleapis/googleapis!/google/rpc/*",
			"google/protobuf/*": "github.com/google/protobuf!/src/google/protobuf/*",
		},
		Fetchers: Fetchers{
			HTTP: FetcherHTTP{
				ModuleToAddr: map[string]string{
					"github.com/googleapis/googleapis": "https://github.com/googleapis/googleapis/blob/{branch}/",
					"github.com/google/protobuf":       "https://github.com/google/protobuf/blob/{branch}/",
					"github.com/gogo/*":                "https://github.com/gogo/*blob/{branch}/",
				},
			},
		},
	}
}

func dedupeAndSort(ss []string) []string {
	lm := map[string]struct{}{}
	for i := range ss {
		lm[ss[i]] = struct{}{}
	}
	ss = make([]string, 0, len(lm))
	for k := range lm {
		ss = append(ss, k)
	}
	sort.Strings(ss)
	return ss
}

func ToFile(c Config, path string) error {
	if path == "" {
		return errors.New("path is empty")
	}

	c.VendoredForeigns = dedupeAndSort(c.VendoredForeigns)
	c.Paths = dedupeAndSort(c.Paths)
	buf, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return errors.Wrapf(ioutil.WriteFile(path, buf, 0644), "when writing '%v'", path)
}

func ToAppConfig(
	c Config,
	localRepoPath string,
	pathToGitCache string,
) (*app.Config, error) {
	var err error
	if !filepath.IsAbs(localRepoPath) {
		lrp := localRepoPath
		localRepoPath, err = filepath.Abs(localRepoPath)
		if err != nil {
			return nil, errors.Wrapf(err, "creating absolute path to '%v'", lrp)
		}
	}
	if !filepath.IsAbs(pathToGitCache) {
		lrp := pathToGitCache
		pathToGitCache, err = filepath.Abs(pathToGitCache)
		if err != nil {
			return nil, errors.Wrapf(err, "creating absolute path to '%v'", lrp)
		}
	}

	return &app.Config{
		ImportReplaces:   c.Replace,
		ForeignFileFQDNs: c.VendoredForeigns,
		Paths:            c.Paths,
		AbsTreeDest:      c.Output,
		ModuleName:       c.RepoModuleName,
		ModuleAbsPath:    localRepoPath,
		Fetchers: app.FetcherConfig{
			Git: fetcher.GitConfig{
				AbsPathToCache:  pathToGitCache,
				ReposToBranches: c.RepoToBranch,
			},
			HTTP: fetcher.HTTPConfig{
				PatternsToHTTPPrefix: c.Fetchers.HTTP.ModuleToAddr,
				ReposToBranches:      c.RepoToBranch,
			},
		},
	}, nil

}
