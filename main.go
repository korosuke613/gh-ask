package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cli/go-gh/pkg/browser"
	"github.com/cli/go-gh/pkg/tableprinter"
	"os"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/cli/go-gh/pkg/term"
)

func main() {
	if err := cli(); err != nil {
		fmt.Fprintf(os.Stderr, "gh-ask failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func cli() error {
	lucky := flag.Bool("lucky", false, "Open the first matching result in a web browser")
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
		return fmt.Errorf("could not determine what repo to use: %w", err)
	}

	if len(flag.Args()) < 1 {
		return errors.New("search term required")
	}
	search := strings.Join(flag.Args(), " ")

	client, err := gh.GQLClient(nil)
	if err != nil {
		return fmt.Errorf("could not create a graphql client: %w", err)
	}

	query := fmt.Sprintf(`{
		repository(owner: "%s", name: "%s") {
			hasDiscussionsEnabled
			discussions(first: 100) {
				edges { node {
					title
					body
					url
		}}}}}`, repo.Owner(), repo.Name())

	type Discussion struct {
		Title string
		URL   string `json:"url"`
		Body  string
	}

	response := struct {
		Repository struct {
			Discussions struct {
				Edges []struct {
					Node Discussion
				}
			}
			HasDiscussionsEnabled bool
		}
	}{}

	err = client.Do(query, nil, &response)
	if err != nil {
		return fmt.Errorf("failed to talk to the GitHub API: %w", err)
	}

	if !response.Repository.HasDiscussionsEnabled {
		return fmt.Errorf("%s/%s does not have discussions enabled", repo.Owner(), repo.Name())
	}

	var matches []Discussion

	for _, edge := range response.Repository.Discussions.Edges {
		if strings.Contains(edge.Node.Body+edge.Node.Title, search) {
			matches = append(matches, edge.Node)
		}
	}

	if len(matches) == 0 {
		fmt.Fprintln(os.Stderr, "No matching discussion threads found :(")
		return nil
	}

	if *lucky {
		b := browser.New("", os.Stdout, os.Stderr)
		err := b.Browse(matches[0].URL)
		if err != nil {
			return fmt.Errorf("failed open browser: %w", err)
		}
		return nil
	}

	isTerminal := term.IsTerminal(os.Stdout)
	tp := tableprinter.New(os.Stdout, isTerminal, 100)

	if isTerminal {
		fmt.Printf(
			"Searching discussions in '%s/%s' for '%s'\n",
			repo.Owner(), repo.Name(), search)
		fmt.Println()

		for _, d := range matches {
			tp.AddField(d.Title)
			tp.AddField(d.URL)
			tp.EndRow()
		}

		err = tp.Render()
		if err != nil {
			return fmt.Errorf("could not render data: %w", err)
		}
	} else {
		for _, d := range matches {
			fmt.Printf("%s %s\n", d.Title, d.URL)
		}
	}

	return nil
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
