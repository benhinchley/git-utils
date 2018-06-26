package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func CreateRepo(name, wd string) (*git.Repository, error) {
	if err := os.Mkdir(name, 0777); err != nil {
		return nil, fmt.Errorf("could not mkdir %q: %v", name, err)
	}
	repo, err := git.PlainInit(name, false)
	if err != nil {
		return nil, fmt.Errorf("could not init repo %q: %v", name, err)
	}

	cmd := exec.Command("git", "commit", "-m", fmt.Sprintf("Root commit for %s", name), "--allow-empty")
	cmd.Dir = filepath.Join(wd, name)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("could not create inital commit: %v", err)
	}

	return repo, nil
}

func ListRemoteBranches(path string) ([]string, error) {
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

func ListBranches(path string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-v")
	cmd.Dir = path
	buf, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed running \"git branch -v\"")
	}

	branches := make([]string, 0, 5)
	for _, branch := range strings.Split(string(buf), "\n") {
		parts := strings.Fields(branch)
		if len(parts) == 0 {
			continue
		}
		branch = parts[0]
		branch = strings.TrimSpace(branch)
		if strings.Contains(branch, "HEAD") || branch == "" || branch == "*" {
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

func CurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run \"git rev-parse --abbrev-ref HEAD\"")
	}
	buf, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
