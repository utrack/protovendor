package fetcher

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

// Git fetches remote repos via git/https, saving them to
// local cache directory.
type Git struct {
	absPathToCache string
	repoToBranch   map[string]string
}

func NewGit(absPathToCacheDir string, branches map[string]string) *Git {
	return &Git{
		absPathToCache: absPathToCacheDir,
		repoToBranch:   branches,
	}
}
func (c *Git) FetchRepo(ctx context.Context, module string) (string, error) {
	// TODO retrieve git address via ?go-get=1 or similar
	dst := filepath.Join(c.absPathToCache, module)

	// TODO allow https/git selection
	repo := "https://" + module

	cmd := exec.Command("git", "fetch")
	if _, err := os.Stat(filepath.Dir(dst)); os.IsNotExist(err) {
		cmd = exec.Command("git", "clone", "--depth", "1", repo, dst)
	} else {
		cmd.Dir = dst
	}
	log.Println("protovendor: fetching ", repo)
	err := cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "when running "+cmd.String())
	}

	branch := "master"
	if v, ok := c.repoToBranch[module]; ok {
		branch = v
	}

	cmd = exec.Command("git", "checkout", "origin/"+branch)
	cmd.Dir = dst
	err = cmd.Run()
	return dst, errors.Wrapf(err, "when checking out '%v' branch", branch)
}