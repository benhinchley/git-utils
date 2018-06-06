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

	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		script, err := createRemoveDirScript(dir)
		if err != nil {
			fmt.Printf("could not create filter script for %q: %v\n", dir, err)
			os.Exit(1)
		}
		var buf bytes.Buffer
		cmd := exec.Command("bash", script)
		cmd.Stderr = &buf

		fmt.Printf("removing %q\n", dir)
		if err := cmd.Run(); err != nil {
			fmt.Println(buf.String())
			os.Exit(1)
		}

		if err := os.Remove(script); err != nil {
			fmt.Printf("could not remove %q: %v\n", script, err)
			os.Exit(1)
		}
	}
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
