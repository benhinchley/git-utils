package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	g "github.com/benhinchley/git-utils/internal/git"
)

func moveWorker(wd string, work []string) (*mergeItem, error) {
	remote, name, folder := work[0], work[1], work[1]
	if len(work) > 2 {
		folder = work[2]
	}
	repoPath := filepath.Join(wd, folder)

	fmt.Printf("rewriting history for %q\n", name)

	if err := g.CloneRepo(name, folder, remote); err != nil {
		return nil, err
	}

	branches, err := g.ListRemoteBranches(repoPath)
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		if err := g.CheckoutBranch(branch, repoPath); err != nil {
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
