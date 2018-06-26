package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/benhinchley/cmd"

	g "github.com/benhinchley/git-utils/internal/git"
)

type moveCmd struct {
	allBranches bool
}

var _ cmd.Command = (*moveCmd)(nil)

const moveCmdHelp = `
mv rewrites the history of the repository so that it appears
that it has always been under the provided directory.
`

func (moveCmd) Name() string { return "mv" }
func (moveCmd) Args() string { return "<dir>" }
func (moveCmd) Desc() string { return "TODO ..." }
func (moveCmd) Help() string { return strings.TrimSpace(moveCmdHelp) }

func (c *moveCmd) Register(fs *flag.FlagSet) {
	fs.BoolVar(&c.allBranches, "all", false, "run mv operation on all branches")
}

func (c *moveCmd) Run(ctx cmd.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments provided")
	}

	dir := args[0]
	wd := ctx.WorkingDir()

	cb, err := g.CurrentBranch(wd)
	if err != nil {
		return fmt.Errorf("could not get current branch: %v", err)
	}

	if c.allBranches {
		branches, err := g.ListRemoteBranches(wd)
		if err != nil {
			return err
		}

		for _, branch := range branches {
			cmd.Out.Printf("rewriting history for branch %s", branch)

			if err := g.CheckoutBranch(branch, wd); err != nil {
				return err
			}

			if err := rewriteHistory(dir, wd); err != nil {
				return err
			}
		}

		return nil
	}

	cmd.Out.Printf("rewriting history for branch %s", cb)
	return rewriteHistory(dir, wd)
}

func rewriteHistory(folder, path string) error {
	s, err := createMoveScript(folder)
	if err != nil {
		return fmt.Errorf("could not write move script: %v", err)
	}

	cmd := exec.Command("bash", s)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not rewrite history: %v", err)
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
