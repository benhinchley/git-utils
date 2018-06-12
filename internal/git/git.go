package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func ListBranches(path string) ([]string, error) {
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

func CheckoutBranch(branch, path string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run \"git checkout %s\"", branch)
	}
	return nil
}

func CloneRepo(name, folder, remote string) error {
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

func createSSHKeyAuth() (*ssh.PublicKeys, error) {
	s := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	keys, err := ssh.NewPublicKeysFromFile("git", s, "")
	if err != nil {
		return nil, err
	}
	return keys, nil
}
