package main

import (
	"fmt"
	"os"
	"os/exec"
)

// git filter-branch -f --index-filter "git rm --cached --ignore-unmatch -r <dir>" --prune-empty
func main() {
	if len(os.Args) < 2 {
		fmt.Println("a directory to remove is required")
		os.Exit(1)
	}

	dir := os.Args[1]
	args := []string{"filter-branch", "-f", "--index-filter", fmt.Sprintf("\"git rm --cached --ignore-unmatch -r %s\"", dir), "--prune-empty"}
	cmd := exec.Command("git", args...)

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
