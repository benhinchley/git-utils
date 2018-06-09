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
	monorepoPath := filepath.Join(wd, monorepoName)

	for _, item := range work {
		remote, err := repo.CreateRemote(&config.RemoteConfig{
			Name: item.Name,
			URLs: []string{item.Remote},
		})
		if err != nil {
			return fmt.Errorf("could not create remote %q: %v", item.Name, err)
		}

		if err := remote.Fetch(&git.FetchOptions{
			RemoteName: item.Name,
		}); err != nil {
			return fmt.Errorf("could not fetch remote: %v", err)
		}

		for _, branch := range item.Branches {
			if err := checkoutBranch(branch, monorepoPath); err != nil {
				if err := createBranch(branch, monorepoPath); err != nil {
					return err
				}
			}

			remoteBranch := fmt.Sprintf("%s/%s", item.Name, branch)
			cmd := exec.Command("git", "merge", remoteBranch, "--allow-unrelated-histories", "-m", fmt.Sprintf("Migrating %[1]s/%[3]s into %[2]s/%[3]s", item.Name, monorepoName, branch))
			cmd.Dir = monorepoPath
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("could not merge %q into %q: %v", remoteBranch, monorepoName, err)
			}

		}

		cmd := exec.Command("git", "fetch", item.Name, fmt.Sprintf("refs/tags/*:refs/tags/%s/*", item.Name), "--no-tags")
		cmd.Dir = monorepoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not merge %q tags into %q", item.Name, monorepoName)
		}

		cmd = exec.Command("git", "remote", "rm", item.Name)
		cmd.Dir = monorepoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("could not remove remote %q: %v", item.Name, err)
		}

		if err := os.RemoveAll(item.Remote); err != nil {
			return fmt.Errorf("could not remove %q: %v", item.Remote, err)
		}
	}

	return checkoutBranch("master", monorepoPath)
}

func createBranch(branch, path string) error {
	cmd := exec.Command("git", "checkout", "--orphan", branch)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run \"git checkout --orphan %s\"", branch)
	}

	cmd = exec.Command("git", "rm", "-rf", "--ignore-unmatch", ".")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run \"git rm -rf --ignore-unmatch .\"")
	}

	cmd = exec.Command("git", "commit", "--allow-empty", "-m", fmt.Sprintf("Root commit for %s branch", branch))
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run \"git commit --allow-empty -m %q\"", fmt.Sprintf("Root commit for %s branch", branch))
	}

	return nil
}
