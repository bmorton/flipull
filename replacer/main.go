package replacer

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/go-github/v48/github"
	"strings"
	"time"
)

var ErrNoContentChange = errors.New("no content changed")

type Replacer struct {
	Owner           string
	Repo            string
	BaseBranch      string
	TargetBranch    string
	Path            string
	Find            string
	Replace         string
	Limit           int
	Title           string
	Description     string
	DryRun          bool
	SkipPullRequest bool
	gh              *github.Client
}

func New(gh *github.Client, repository string) (*Replacer, error) {
	repoParts := strings.Split(repository, "/")
	if len(repoParts) != 2 {
		return nil, fmt.Errorf("invalid repository: %s", repository)
	}
	branchParts := strings.Split(repoParts[1], "@")
	var baseBranch string
	if len(branchParts) == 1 {
		baseBranch = "main"
	} else if len(branchParts) == 2 {
		baseBranch = branchParts[1]
	} else {
		return nil, fmt.Errorf("invalid repository: %s", repository)
	}

	return &Replacer{
		gh:           gh,
		Owner:        repoParts[0],
		Repo:         branchParts[0],
		BaseBranch:   baseBranch,
		TargetBranch: fmt.Sprintf("flipull/replacer/%d", time.Now().Unix()),
	}, nil
}

func (r Replacer) Run(ctx context.Context) error {
	fmt.Printf("-- Modifying %s...\n", r.Path)
	newContent, err := r.generateContent(ctx)
	if err != nil {
		return err
	}

	if r.DryRun {
		fmt.Println(newContent)
	} else {
		fmt.Printf("-- Committing changes...\n")
		_, err = r.commitToBranch(ctx, newContent, r.Description, branchRef(r.BaseBranch, false))
		if err != nil {
			return err
		}

		if r.SkipPullRequest {
			fmt.Printf("Success! (skipped pull request)\n")
		} else {
			fmt.Printf("-- Creating pull request...\n")
			create, err := r.createPullRequest(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("Success!\n%s\n", create.GetHTMLURL())
		}
	}

	return nil
}

func (r Replacer) getRef(ctx context.Context, ref string) (*github.Reference, error) {
	gitRef, _, err := r.gh.Git.GetRef(ctx, r.Owner, r.Repo, ref)
	return gitRef, err
}

func (r Replacer) getContent(ctx context.Context, sha string, path string) (string, error) {
	file, _, _, err := r.gh.Repositories.GetContents(ctx, r.Owner, r.Repo, path,
		&github.RepositoryContentGetOptions{
			Ref: sha,
		},
	)
	if err != nil {
		return "", nil
	}

	return file.GetContent()
}

func (r Replacer) createPullRequest(ctx context.Context) (*github.PullRequest, error) {
	baseBranch := branchRef(r.BaseBranch, true)
	targetBranch := branchRef(r.TargetBranch, true)
	newPull := &github.NewPullRequest{
		Title: &r.Title,
		Body:  &r.Description,
		Base:  &baseBranch,
		Head:  &targetBranch,
	}
	create, _, err := r.gh.PullRequests.Create(ctx, r.Owner, r.Repo, newPull)
	if err != nil {
		return nil, err
	}
	return create, nil
}

func (r Replacer) commitToBranch(ctx context.Context, content string, message string, baseRef string) (*github.Reference, error) {
	ref, err := r.getRef(ctx, baseRef)

	blob := &github.Blob{
		Content: &content,
	}
	createBlob, _, err := r.gh.Git.CreateBlob(ctx, r.Owner, r.Repo, blob)
	if err != nil {
		return nil, err
	}

	fileMode := "100644"
	newTree := []*github.TreeEntry{
		{
			Path: &r.Path,
			Mode: &fileMode,
			SHA:  createBlob.SHA,
		},
	}
	tree, _, err := r.gh.Git.CreateTree(ctx, r.Owner, r.Repo, ref.GetObject().GetSHA(), newTree)
	if err != nil {
		return nil, err
	}

	newParentCommit := &github.Commit{
		SHA: ref.GetObject().SHA,
	}

	newCommit := &github.Commit{
		Parents: []*github.Commit{newParentCommit},
		Message: &message,
		Tree:    tree,
	}
	commit, _, err := r.gh.Git.CreateCommit(ctx, r.Owner, r.Repo, newCommit)
	if err != nil {
		return nil, err
	}

	targetBranch := branchRef(r.TargetBranch, true)
	newBranch := &github.Reference{
		Ref: &targetBranch,
		Object: &github.GitObject{
			SHA: commit.SHA,
		},
	}
	newRef, _, err := r.gh.Git.CreateRef(ctx, r.Owner, r.Repo, newBranch)

	return newRef, err
}

func (r Replacer) generateContent(ctx context.Context) (string, error) {
	content, err := r.getContent(ctx, branchRef(r.BaseBranch, true), r.Path)
	if err != nil {
		return "", err
	}

	newContent := strings.Replace(content, r.Find, r.Replace, r.Limit)

	if newContent == content {
		return "", ErrNoContentChange
	}

	return newContent, nil
}

func branchRef(name string, prefix bool) string {
	var format string
	if prefix {
		format = "refs/heads/%s"
	} else {
		format = "heads/%s"
	}
	return fmt.Sprintf(format, name)
}
