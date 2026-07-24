package workspace

import (
	"errors"
	"os"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

const authorName = "cpenv"
const authorEmail = "cpenv@local"
const remoteName = "origin"

type store struct {
	repo *git.Repository
}

func (s *store) save() error {
	head, err := s.repo.Head()
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return errors.New("missing branch")
	} else if err != nil {
		return err
	}

	ref := head.Name()
	if !ref.IsBranch() {
		return errors.New("detached HEAD")
	}

	w, err := s.repo.Worktree()
	if err != nil {
		return err
	}

	if err := w.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return err
	}

	now := time.Now().UTC()
	if _, err := w.Commit(now.Format(time.RFC3339), &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  now,
		},
	}); err != nil && !errors.Is(err, git.ErrEmptyCommit) {
		return err
	}

	if err := s.repo.Push(&git.PushOptions{
		RemoteName: remoteName,
		RefSpecs: []config.RefSpec{
			config.RefSpec(ref.String() + ":" + ref.String()),
		},
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	return nil
}

func (s *store) load(problemID string) error {
	w, err := s.repo.Worktree()
	if err != nil {
		return err
	}

	branch := plumbing.NewBranchReferenceName(problemID)
	if _, err := s.repo.Reference(branch, true); err == nil {
		if err := w.Checkout(&git.CheckoutOptions{
			Branch: branch,
			Force:  true,
		}); err != nil {
			return err
		}
		return nil
	} else if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return err
	}

	if err := s.repo.Storer.SetReference(
		plumbing.NewSymbolicReference(plumbing.HEAD, branch),
	); err != nil {
		return err
	}

	if err := w.RemoveGlob("*"); err != nil {
		return err
	}

	if err := w.Clean(&git.CleanOptions{Dir: true}); err != nil {
		return err
	}

	return nil
}

func ensureRepo(path string, isBare bool) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}

		repo, err = git.PlainInit(path, isBare)
		if err != nil {
			return nil, err
		}

		return repo, nil
	} else if err != nil {
		return nil, err
	}

	cfg, err := repo.Config()
	if err != nil {
		return nil, err
	}

	if cfg.Core.IsBare != isBare {
		return nil, errors.New("repository bare mismatch")
	}

	return repo, nil
}

func ensureRemote(repo *git.Repository, name, url string) error {
	remote, err := repo.Remote(name)
	if errors.Is(err, git.ErrRemoteNotFound) {
		if _, err = repo.CreateRemote(&config.RemoteConfig{
			Name: name,
			URLs: []string{url},
		}); err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	cfg := remote.Config()
	if !(len(cfg.URLs) > 0 && cfg.URLs[0] == url) {
		return errors.New("remote URL mismatch")
	}

	return nil
}

func newStore(archiveDir, workspaceDir string) (*store, error) {
	_, err := ensureRepo(archiveDir, true)
	if err != nil {
		return nil, err
	}

	repo, err := ensureRepo(workspaceDir, false)
	if err != nil {
		return nil, err
	}

	if err := ensureRemote(repo, remoteName, archiveDir); err != nil {
		return nil, err
	}

	return &store{repo: repo}, nil
}
