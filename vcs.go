package makeaddon

import (
	"github.com/Masterminds/vcs"
	"log"
	"strings"
)

// checkout clones the remote repository to a cache directory.
func checkout(url, tag string) (string, error) {
	log.Printf("Checking out %s", url)
	dir, fresh := cache.Dir(url, tag)

	var repo vcs.Repo
	var err error
	if strings.HasPrefix(url, "https://repos.wowace.com/wow/") {
		repo, err = vcs.NewSvnRepo(url, dir)
	} else {
		repo, err = vcs.NewRepo(url, dir)
	}
	if err != nil {
		return "", err
	}

	if fresh {
		if err = repo.Get(); err != nil {
			return "", err
		}
		if len(tag) > 0 {
			if err = repo.UpdateVersion(tag); err != nil {
				return "", err
			}
		}
	}

	if tag == "latest" {
		// TODO: Find latest tag and use that version instead
	} else if tag == "" {
		if err = repo.Update(); err != nil {
			return "", err
		}
	}

	return dir, nil
}
