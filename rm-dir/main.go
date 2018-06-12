// Command rm-dir removes directories from a git repository's history
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/benhinchley/git-utils/internal/git"
)

var (
	directories string
)

func main() {
	flag.StringVar(&directories, "dirs", "", "comma seperated list of directories to remove")
	flag.Parse()

	dirs := strings.Split(directories, ",")
	if len(dirs) == 1 && dirs[0] == "" {
		fmt.Println("at least 1 directory is required")
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("could not get working directory: %v\n", err)
		os.Exit(1)
	}

	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		branches, err := git.ListBranches(wd)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, branch := range branches {
			if err := git.CheckoutBranch(branch, wd); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("removing %q on branch %q\n", dir, branch)
			if err := removeDirectory(dir); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

	}
}

func removeDirectory(dir string) error {
	script, err := createRemoveDirScript(dir)
	if err != nil {
		return fmt.Errorf("could not create filter script for %q: %v", dir, err)
	}
	var buf bytes.Buffer
	cmd := exec.Command("bash", script)
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove %q from git history", dir)
	}

	if err := os.Remove(script); err != nil {
		return fmt.Errorf("could not remove %q: %v", script, err)
	}

	return nil
}

const removeDirFilterScript = `
#!/usr/bin/env bash
git filter-branch -f --index-filter 'git rm --cached --ignore-unmatch -r %s' --prune-empty
`

func createRemoveDirScript(dir string) (string, error) {
	f, err := ioutil.TempFile("", dir)
	if err != nil {
		return "", err
	}

	if _, err := f.WriteString(fmt.Sprintf(removeDirFilterScript, dir)); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	return f.Name(), nil
}
