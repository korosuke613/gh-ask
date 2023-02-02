package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/repository"
)

func main() {
	if err := cli(); err != nil {
		fmt.Fprintf(os.Stderr, "gh-ask failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func cli() error {
	repoOverride := flag.String(
		"repo", "", "Specify a repository. If omitted, uses current repository")
	flag.Parse()

	var repo repository.Repository
	var err error

	if *repoOverride == "" {
		repo, err = gh.CurrentRepository()
	} else {
		repo, err = repository.Parse(*repoOverride)
	}
	if err != nil {
		return fmt.Errorf("could not determine what repo to use: %v", err.Error())
	}

	if len(flag.Args()) < 1 {
		return errors.New("search term required")
	}
	search := strings.Join(flag.Args(), " ")

	fmt.Printf("Going to search discussions in '%s/%s' for '%s'\n", repo.Owner(), repo.Name(), search)

	return nil
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
