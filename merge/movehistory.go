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

	fmt.Printf("rewriting history for %q\n", name)

	if err := os.Mkdir(folder, 0777); err != nil {
		return nil, fmt.Errorf("could not create tmp dir to clone %q into: %v", name, err)
	}

	auth, err := createSSHKeyAuth()
	if err != nil {
		return nil, err
	}

	_, err = git.PlainClone(folder, false, &git.CloneOptions{
		URL:  remote,
		Auth: auth,
	})
	if err != nil {
		return nil, fmt.Errorf("could not clone %q: %v", remote, err)
	}
	cmd := exec.Command("git", "fetch")
	cmd.Dir = filepath.Join(wd, folder)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("could not fetch %q: %v", name, err)
	}

	s, err := createMoveScript(folder)
	if err != nil {
		return nil, fmt.Errorf("could not write move script: %v", err)
	}

	cmd = exec.Command("bash", s)
	cmd.Dir = filepath.Join(wd, folder)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("could not rewrite history for %q: %v", name, err)
	}
	if err := os.Remove(s); err != nil {
		return nil, fmt.Errorf("could not remove %q: %v", s, err)
	}

	abs, _ := filepath.Abs(filepath.Join(wd, folder))
	return &mergeItem{
		Remote: abs,
		Name:   name,
	}, nil
}

const moveFilterScript = `
#!/usr/bin/env bash
git filter-branch -f --index-filter 'git ls-files -s | sed "s-	\"*-&%s/-" | GIT_INDEX_FILE=$GIT_INDEX_FILE.new git update-index --index-info && mv "$GIT_INDEX_FILE.new" "$GIT_INDEX_FILE"' HEAD
`

func createMoveScript(dir string) (string, error) {
	f, err := ioutil.TempFile("", dir)
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
