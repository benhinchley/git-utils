package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	g "github.com/benhinchley/git-utils/internal/git"
)

var (
	monorepoName string
	inputFile    string
)

func main() {
	flag.StringVar(&monorepoName, "name", "go", "name of the monorepo to be created")
	flag.StringVar(&inputFile, "input", "", "file containing repos to be merged")
	flag.Parse()

	if inputFile == "" {
		fmt.Println("-input flag is required")
		flag.Usage()
		os.Exit(2)
	}

	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("could not open %q: %v", inputFile, err)
		os.Exit(1)
	}
	defer f.Close()

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("could not get working directory: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(f)
	moveWork := make(chan []string, 4)
	go func() {
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 2 {
				fmt.Println("malformed input file")
				os.Exit(1)
			}

			moveWork <- fields
		}
		close(moveWork)
	}()

	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	mergeWork := make([]*mergeItem, 0, 10)
	done := make(chan struct{})
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

		FOR_LOOP:
			for {
				select {
				case item, ok := <-moveWork:
					if !ok {
						break FOR_LOOP
					}

					mItem, err := moveWorker(wd, item)
					if err != nil {
						select {
						case errChan <- err:
							close(done)

						default:
						}
						continue FOR_LOOP
					}
					mergeWork = append(mergeWork, mItem)

				case <-done:
					break FOR_LOOP
				}
			}

		}()
	}
	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	repo, err := g.CreateRepo(monorepoName, wd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("merging repositories into %q\n", monorepoName)
	if err := mergeRepos(repo, wd, mergeWork); err != nil {
		fmt.Printf("could not merge repositories together: %v\n", err)
		os.Exit(1)
	}
}

type mergeItem struct {
	// Remote is the path to the tmp clone dir
	Remote string
	// Name is the name of the remote
	Name     string
	Branches []string
}
