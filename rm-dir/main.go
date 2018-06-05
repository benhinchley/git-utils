// Command rm-dir removes directories from a git repository's history
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	directories string
)

// git filter-branch -f --index-filter "git rm --cached --ignore-unmatch -r <dir>" --prune-empty
func main() {
	flag.StringVar(&directories, "dirs", "", "comma seperated list of directories to remove")
	flag.Parse()

	dirs := strings.Split(directories, ",")
	if len(dirs) == 1 && dirs[0] == "" {
		fmt.Println("at least 1 directory is required")
		os.Exit(1)
	}

	for _, dir := range dirs {
		args := []string{"filter-branch", "-f", "--index-filter", fmt.Sprintf("\"git rm --cached --ignore-unmatch -r %s\"", dir), "--prune-empty"}
		cmd := exec.Command("git", args...)

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
