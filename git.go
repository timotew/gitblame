package main

import (
	"encoding/hex"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io"
	"time"
)

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Commit struct {
	Hash    string    `json:"hash"`
	Message string    `json:"message"`
	When    time.Time `json:"when"`
}

type AuthorCommits struct {
	Author  `json:"author"`
	Commits []Commit `json:"commits"`
}

func GroupCommitsByAuthor(r *git.Repository) ([]AuthorCommits, error) {
	authors := make(map[Author][]Commit)

	cIter, err := r.Log(&git.LogOptions{All: true})
	if err != nil {
		return nil, err
	}
	defer cIter.Close()

	cIter.ForEach(func(c *object.Commit) error {

		author := Author{Name: c.Author.Name, Email: c.Author.Email}
		commits, found := authors[author]
		if !found {
			commits = make([]Commit, 0, 10)
		}
		commits = append(commits,
			Commit{Message: c.Message,
				Hash: hex.EncodeToString(c.Hash[:]),
				When: c.Author.When})
		authors[author] = commits
		return nil
	})

	response := toSlice(authors)

	return response, nil
}
func GetCommitPatch(c *object.Commit) (*object.Patch, error) {

	tree, err := c.Tree()
	if err != nil {
		return nil, err
	}

	parents := c.Parents()
	defer parents.Close()

	parent, err := parents.Next()
	if err != nil && err != io.EOF {
		return nil, err
	}

	var prevTree *object.Tree
	if parent != nil {
		prevTree, err = parent.Tree()
		if err != nil {
			return nil, err
		}
	}

	changes, err := prevTree.Diff(tree)
	if err != nil {
		return nil, err
	}

	patch, err := changes.Patch()
	if err != nil {
		return nil, err
	}

	return patch, nil
}

func Pull(r *git.Repository, auth *http.BasicAuth) error {
	wt, err := r.Worktree()
	if err != nil {
		return err
	}

	err = wt.Pull(&git.PullOptions{Auth: auth})
	if err != nil {
		return err
	}

	return err
}

func GetPatch(hash []byte, err error, r *git.Repository) (string, error) {

	var hashArr [20]byte
	copy(hashArr[:], hash)
	c, err := object.GetCommit(r.Storer, hashArr)
	if err != nil {
		return "", err
	}

	patch, err := GetCommitPatch(c)
	if err != nil {
		return "", err
	}

	return patch.String(), nil
}

// --------

// TODO call from api
func getStats(c *object.Commit) (int, int, error) {
	var (
		additions int
		deletions int
	)
	fileStats, err := c.Stats()
	if err != nil {
		return 0, 0, err
	}

	for index := range fileStats {
		additions += fileStats[index].Addition
		deletions += fileStats[index].Deletion
	}
	return additions, deletions, nil
}

func toSlice(authors map[Author][]Commit) []AuthorCommits {
	authorCommits := make([]AuthorCommits, 0)
	for k, v := range authors {
		authorCommits = append(authorCommits, AuthorCommits{Author: k, Commits: v})
	}
	return authorCommits
}
