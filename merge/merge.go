package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
)

func mergeRepos(repo *git.Repository, wd string, work []*mergeItem) error {
	for _, item := range work {
		remote, err := repo.CreateRemote(&config.RemoteConfig{
			Name: item.Name,
			URLs: []string{item.Remote},
		})
		if err != nil {
			return fmt.Errorf("could not create remote %q: %v", item.Name, err)
		}
		auth, err := createSSHKeyAuth()
		if err != nil {
			return fmt.Errorf("could not create auth: %v", err)
		}

		if err := remote.Fetch(&git.FetchOptions{
			RemoteName: item.Name,
			Auth:       auth,
		}); err != nil {
			return fmt.Errorf("could not fetch remote: %v", err)
		}

		cmd := exec.Command("git", "merge", fmt.Sprintf("%s/master", item.Name), "--allow-unrelated-histories", "-m", fmt.Sprintf("Migrating %s/master into %s/master", item.Name, monorepoName))
		cmd.Dir = filepath.Join(wd, monorepoName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not merge %q into %q: %v", item.Name, monorepoName, err)
		}

		cmd = exec.Command("git", "fetch", item.Name, fmt.Sprintf("refs/tags/*:refs/tags/%s/*", item.Name), "--no-tags")
		cmd.Dir = filepath.Join(wd, monorepoName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not merge %q tags into %q: %v", item.Name, monorepoName, err)
		}

		cmd = exec.Command("git", "remote", "rm", item.Name)
		cmd.Dir = filepath.Join(wd, monorepoName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not remove remote %q: %v", item.Name, err)
		}

		if err := os.RemoveAll(item.Remote); err != nil {
			return fmt.Errorf("could not remove %q: %v", item.Remote, err)
		}
	}

	return nil
}
