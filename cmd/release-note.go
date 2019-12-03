package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/google/go-github/v26/github"
	"golang.org/x/oauth2"
)

func Run() {

	user := flag.String("user", "", "Your github handle.")
	token := flag.String("token", "", "Your github token.")
	repoToInspect := flag.String("repository", "", "The github repository to inspect, in org/name format.")
	fromCommit := flag.String("from", "", "the older commit (excluded)")
	toCommit := flag.String("to", "HEAD", "the newer commit (included), defaults to HEAD")
	debug := flag.Bool("debug", false, "enable debug logs")

	log := func(msg string, args ...interface{}) {
		if *debug {
			fmt.Printf(msg, args...)
		}
	}

	flag.Parse()

	if user == nil || *user == "" {
		panic("missing user flag")
	}

	if token == nil || *token == "" {
		panic("missing token flag")
	}

	if repoToInspect == nil || *repoToInspect == "" {
		panic("missing repository flag")
	}

	if fromCommit == nil || *fromCommit == "" {
		panic("missing from flag")
	}

	// init github client and get repositories
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	repos, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		panic(err)
	}

	for _, currentRepo := range repos {
		match, owner, repo := matchesRepo(currentRepo, *repoToInspect)
		if match {
			log("repo: %v\n", *currentRepo.Name)

			// first find merge commits
			mergeCommits := make([]string, 0)

			branch, _, err := client.Repositories.GetBranch(ctx, owner, repo, "master")
			if err != nil {
				panic(err)
			}

			sha := *branch.Commit.SHA
			parents := branch.Commit.Parents

			inRange := false
			if *toCommit == "HEAD" {
				log("1st in range: %s\n", sha)
				inRange = true
			}

			for sha != "" {
				if !inRange && sha == *toCommit {
					inRange = true
					log("1st in range: %s\n", sha)
				}
				if sha == *fromCommit {
					log("found fromCommit: %s\n", sha)
					sha = ""
					break
				}
				if inRange {
					mergeCommits = append(mergeCommits, sha)
					log("adding: %s\n", sha)
				} else {
					log("skipping: %s\n", sha)
				}

				// make sure to leave if no parent was found
				sha = ""

				for _, parent := range parents {

					// get commit
					commit, _, err := client.Repositories.GetCommit(ctx, owner, repo, *parent.SHA)
					if err != nil {
						panic(err)
					}

					if strings.HasPrefix(*commit.Commit.Message, "Merge pull request") {
						sha = *commit.SHA
						parents = commit.Parents
						break
					}
				}
			}

			nextPage := 1
			for nextPage != 0 {
				prs, response, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
					State:     "all",
					Sort:      "updated",
					Direction: "desc",
					ListOptions: github.ListOptions{
						Page:    nextPage,
						PerPage: 20,
					},
				})
				if err != nil {
					panic(err)
				}

				nextPage = response.NextPage

				for _, pr := range prs {

					log("pr nr %v, state %v: %v\n", *pr.Number, *pr.State, *pr.Title)

					// find merged PRs
					if pr.MergedAt != nil {
						log("  is merged")
						match, mergeCommits = containsAndDelete(mergeCommits, *pr.MergeCommitSHA)
						if match {
							log("    in range")

							// try to extract release note
							fmt.Printf("- %s", findReleaseNote(*pr.Body))

						}

						if len(mergeCommits) == 0 {
							// all PRs found :)
							nextPage = 0
							break
						}
					}

				}  // end current prs
			} // end pr page

			break
		} // end matched repo
	} // end repos
}

func matchesRepo(current *github.Repository, repoToInspect string) (match bool, owner string, repo string)  {
	owner, repo = splitRepo(repoToInspect)
	match = *current.Owner.Login == owner && *current.Name == repo
	return
}

func splitRepo(repoToInspect string) (owner, repo string) {
	parts := strings.Split(repoToInspect, "/")
	if len(parts) != 2 {
		panic("malformed repository flag value")
	}
	owner = parts[0]
	repo = parts[1]
	return
}

func containsAndDelete(all []string, one string) (bool, []string) {
	for i, s := range all {
		if s == one {
			all = append(all[:i], all[i+1:]...)
			return true, all
		}
	}
	return false, all
}

func findReleaseNote(desc string) string {
	inRelNote := false
	shortDesc := ""
	relNote := ""
	for _, line := range strings.Split(desc, "\n") {

		line = strings.TrimSpace(line)
		// skip empty lines
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "```release-note") {
			inRelNote = true
			continue
		} else if strings.HasPrefix(line, "```") {
			// inRelNote = false
			continue
		} else {

			if inRelNote {
				if !strings.Contains(line, "NONE") {
					relNote += line + "\n"
				}
			} else {
				// skip boilerplate
				skip := false
				for _, prefix := range strings.Split(prTemplate, "\n") {
					if prefix != "" && strings.HasPrefix(line, prefix) {
						skip = true
						break
					}
				}
				if !skip {
					shortDesc += line + "\n"
				}
			}

		}
	}
	// return release note if filled, else description without rel note section
	if relNote != "" {
		return relNote
	} else {
		return "no relNote, desc: " + shortDesc
	}
}


const prTemplate = `
<!--  Thanks for sending a pull request!  Here are some tips for you:
1. Follow the instructions for writing a release note from k8s: https://git.k8s.io/community/contributors/guide/release-notes.md
-->
**What this PR does / why we need it**:
**Which issue(s) this PR fixes** 
Fixes #
**Special notes for your reviewer**:
**Release note**:
<!--  Write your release note:
1. Enter your extended release note in the below block. If the PR requires additional action from users switching to the new release, include the string "action required".
2. If no release note is required, just write "NONE".
-->
Signed-off-by:
`