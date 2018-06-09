package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func moveWorker(wd string, work []string) (*mergeItem, error) {
	remote, name, folder := work[0], work[1], work[1]
	if len(work) > 2 {
		folder = work[2]
	}
	repoPath := filepath.Join(wd, folder)

	fmt.Printf("rewriting history for %q\n", name)

	if err := cloneRepo(name, folder, remote); err != nil {
		return nil, err
	}

	branches, err := listBranches(remote, repoPath)
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		if err := checkoutBranch(branch, repoPath); err != nil {
			return nil, err
		}

		if err := rewriteHistory(name, folder, repoPath); err != nil {
			return nil, err
		}
	}

	abs, _ := filepath.Abs(filepath.Join(wd, folder))
	return &mergeItem{
		Remote:   abs,
		Name:     name,
		Branches: branches,
	}, nil
}

func cloneRepo(name, folder, remote string) error {
	if err := os.Mkdir(folder, 0777); err != nil {
		return fmt.Errorf("could not create tmp dir to clone %q into: %v", name, err)
	}

	auth, err := createSSHKeyAuth()
	if err != nil {
		return err
	}

	_, err = git.PlainClone(folder, false, &git.CloneOptions{
		URL:  remote,
		Auth: auth,
	})
	if err != nil {
		return fmt.Errorf("could not clone %q: %v", remote, err)
	}

	return nil
}

func listBranches(remote, path string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")
	cmd.Dir = path
	buf, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed running \"git branch -r\"")
	}

	branches := make([]string, 0, 5)
	for _, branch := range strings.Split(string(buf), "\n") {
		branch = strings.TrimSpace(branch)
		branch = strings.Replace(branch, "origin/", "", -1)
		if strings.HasPrefix(branch, "HEAD") || branch == "" {
			continue
		}
		branches = append(branches, branch)
	}

	return branches, nil
}

func checkoutBranch(branch, path string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run \"git checkout %s\"", branch)
	}
	return nil
}

func rewriteHistory(name, folder, path string) error {
	s, err := createMoveScript(folder)
	if err != nil {
		return fmt.Errorf("could not write move script: %v", err)
	}

	cmd := exec.Command("bash", s)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not rewrite history for %q: %v", name, err)
	}

	if err := os.Remove(s); err != nil {
		return fmt.Errorf("could not remove %q: %v", s, err)
	}

	return nil
}

const moveFilterScript = `
#!/usr/bin/env bash
git filter-branch -f --index-filter 'git ls-files -s | sed "s-	\"*-&%s/-" | GIT_INDEX_FILE=$GIT_INDEX_FILE.new git update-index --index-info && mv "$GIT_INDEX_FILE.new" "$GIT_INDEX_FILE"' HEAD
`

func createMoveScript(dir string) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	dir = strings.Replace(dir, ".", "\\.", -1)
	dir = strings.Replace(dir, "-", "\\-", -1)

	if _, err := f.WriteString(fmt.Sprintf(moveFilterScript, dir)); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func createSSHKeyAuth() (*ssh.PublicKeys, error) {
	s := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	keys, err := ssh.NewPublicKeysFromFile("git", s, "")
	if err != nil {
		return nil, err
	}
	return keys, nil
}
