package main

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io"
)

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AuthorCommits map[Author][]object.Commit

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

func GroupCommitsByAuthor(r *git.Repository) (*AuthorCommits, error) {
	authorCommits := make(AuthorCommits)

	cIter, err := r.Log(&git.LogOptions{All: true})
	if err != nil {
		return nil, err
	}
	defer cIter.Close()

	cIter.ForEach(func(c *object.Commit) error {

		author := Author{Name: c.Author.Name, Email: c.Author.Email}
		commits, found := authorCommits[author]
		if !found {
			commits = make([]object.Commit, 0, 10)
		}
		commits = append(commits, *c)
		authorCommits[author] = commits

		return nil
	})

	return &authorCommits, nil
}

func Pull(r *git.Repository, auth *http.BasicAuth) error {
	wt, err := r.Worktree()
	logIfError(err)
	err = wt.Pull(&git.PullOptions{Auth: auth})
	logIfError(err)
	return err
}

func GetPatch(hash []byte, err error, r *git.Repository) *object.Patch {
	var hashArr [20]byte
	copy(hashArr[:], hash)
	c, err := object.GetCommit(r.Storer, hashArr)
	logIfError(err)
	patch, err := GetCommitPatch(c)
	return patch
}
