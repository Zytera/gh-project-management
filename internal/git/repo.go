package git

import (
	"errors"

	gitm "github.com/gogs/git-module"
	giturl "github.com/kubescape/go-git-url"
)

func GetRepo(dir string) (*gitm.Repository, error) {
	repo, err := gitm.Open(dir)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func GetOriginUrl(dir string) (string, error) {
	repo, err := GetRepo(dir)
	if err != nil {
		return "", err
	}
	remotes, err := repo.Remotes()
	if err != nil {
		return "", err
	}

	for _, remote := range remotes {
		if remote != "origin" {
			continue
		}

		urls, err := gitm.RemoteGetURL(dir, remote)
		if err != nil || len(urls) == 0 {
			return "", err
		}
		return urls[0], nil
	}

	return "", errors.New("no origin remote found")
}

func GetRepoName(dir string) (string, error) {
	originUrl, err := GetOriginUrl(dir)

	if err != nil {
		return "", err
	}

	gitURL, err := giturl.NewGitURL(originUrl)
	if err != nil {
		return "", err
	}

	return gitURL.GetRepoName(), nil
}
