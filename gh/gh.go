package gh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

type Client struct {
  ctx context.Context
  GithubRef string
  Org string
  Repo string

  ghClient *github.Client
}

func CreateClient(ctx context.Context, org string, repo string, apiKey string) (*Client) {
  token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})
  tokenClient := oauth2.NewClient(ctx, token)
  ghClient := github.NewClient(tokenClient)

  return &Client {
    Org: org,
    Repo: repo,
    GithubRef: os.Getenv("GITHUB_REF"),

    ghClient: ghClient,
    ctx: ctx,
  }
}

// Gets the pull-request(s) associated with the current action.
func (client Client) GetCurrentPullRequests() ([]*github.PullRequest, error) {
  eventName := os.Getenv("GITHUB_EVENT_NAME")
  if eventName == "pull_request" {
    // Ref is a PR. Fetch PR by PR number, supplied by ref.
    prNum, err := strconv.Atoi(strings.Split(client.GithubRef, "/")[2])
    if err != nil { return nil, errors.New("Failed to split and/or parse GitHub pull-request ref string.") }
    pr, _, err := client.ghClient.PullRequests.Get(client.ctx, client.Org, client.Repo, prNum)
    if err != nil { return nil, errors.New("Failed to get pull-request from GitHub API.") }
    
    return []*github.PullRequest{pr}, nil
  } else if eventName == "push" {
    // Ref is a branch. Get PR by searching for branch, supplied by ref.
    opts := github.PullRequestListOptions {
      Head: fmt.Sprintf("%s: %s", client.Repo, client.GithubRef),
    }
    prs, _, err := client.ghClient.PullRequests.List(client.ctx, client.Org, client.Repo, &opts)
    if err != nil { return nil, errors.New("Failed to list pull-requests from GitHub API.") }

    return prs, nil
  } else {
    return nil, errors.New("Unsupported action event type.")
  }
}

// func startPipeline() {
//   fmt.Printf("Attempting to start pipeline for ref: %s\n", headRef)
//   num, err := cciClient.CreatePipeline(headRef)
//   if err != nil {
//     if safeMatchString("Ref not found", err) {
//       fmt.Printf("CCI: Ref %s not found. Ignoring.\n", headRef)
//       return
//     }
//     panic(err)
//   }
// 
//   fmt.Printf("CCI: Started pipline #%d", num)
// }

