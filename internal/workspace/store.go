package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

const (
	authorName  = "cpenv"
	authorEmail = "cpenv@local"
	remoteName  = "origin"
)

type storeOptions struct {
	archiveDir   string
	workspaceDir string
	baseBranch   string
}

type store struct {
	options storeOptions
	repo    *git.Repository
}

func (s *store) save() error {
	head, err := s.repo.Head()
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return fmt.Errorf("save: unborn branch: %w", err)
	} else if err != nil {
		return fmt.Errorf("save: resolve HEAD: %w", err)
	}

	ref := head.Name()
	if !ref.IsBranch() {
		return errors.New("save: detached HEAD")
	}

	w, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("save: get worktree: %w", err)
	}

	if err := w.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return fmt.Errorf("save: add: %w", err)
	}

	now := time.Now()
	if _, err := w.Commit(now.Format("3:04 PM (Jan 2, 2006)"), &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  now,
		},
	}); err != nil && !errors.Is(err, git.ErrEmptyCommit) {
		return fmt.Errorf("save: commit: %w", err)
	}

	if err := s.repo.Push(&git.PushOptions{
		RemoteName: remoteName,
		RefSpecs: []config.RefSpec{
			config.RefSpec(ref.String() + ":" + ref.String()),
		},
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("save: push: %w", err)
	}

	return nil
}

func (s *store) load(name string) error {
	branch := plumbing.NewBranchReferenceName(name)
	if err := branch.Validate(); err != nil {
		return fmt.Errorf("load: invalid branch name %q: %w", name, err)
	}

	w, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("load: get worktree: %w", err)
	}

	if _, err := s.repo.Reference(branch, true); err == nil {
		if err := w.Checkout(&git.CheckoutOptions{Branch: branch}); err != nil {
			return fmt.Errorf("load: checkout %q: %w", name, err)
		}
	} else if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return fmt.Errorf("load: resolve branch %q: %w", name, err)
	} else {
		baseBranch := plumbing.NewBranchReferenceName(s.options.baseBranch)
		baseRef, err := s.repo.Reference(baseBranch, true)
		if err != nil {
			return fmt.Errorf("load: resolve branch %q: %w", s.options.baseBranch, err)
		}

		if err := w.Checkout(&git.CheckoutOptions{
			Branch: branch,
			Create: true,
			Hash:   baseRef.Hash(),
		}); err != nil {
			return fmt.Errorf("load: checkout %q: %w", name, err)
		}
	}

	if err := w.Clean(&git.CleanOptions{Dir: true}); err != nil {
		return fmt.Errorf("load: clean: %w", err)
	}

	return nil
}

func ensureRepo(path string, isBare bool) (*git.Repository, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("ensure repo %q: path must be absolute", path)
	}

	repo, err := git.PlainOpen(path)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("ensure repo %q: mkdir: %w", path, err)
		}

		repo, err = git.PlainInit(path, isBare)
		if err != nil {
			return nil, fmt.Errorf("ensure repo %q: init: %w", path, err)
		}

		return repo, nil
	} else if err != nil {
		return nil, fmt.Errorf("ensure repo %q: open: %w", path, err)
	}

	cfg, err := repo.Config()
	if err != nil {
		return nil, fmt.Errorf("ensure repo %q: get config: %w", path, err)
	}

	if cfg.Core.IsBare != isBare {
		return nil, fmt.Errorf("ensure repo %q: bare: expected %t, got %t", path, isBare, cfg.Core.IsBare)
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
			return fmt.Errorf("ensure remote %q: create: %w", name, err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("ensure remote %q: get: %w", name, err)
	}

	cfg := remote.Config()
	if len(cfg.URLs) == 0 {
		return fmt.Errorf("ensure remote %q: missing url", name)
	}
	if cfg.URLs[0] != url {
		return fmt.Errorf("ensure remote %q: url: expected %q, got %q", name, url, cfg.URLs[0])
	}

	return nil
}

func ensureBranch(repo *git.Repository, name string) error {
	branch := plumbing.NewBranchReferenceName(name)
	if err := branch.Validate(); err != nil {
		return fmt.Errorf("ensure branch %q: invalid name: %w", name, err)
	}

	if _, err := repo.Reference(branch, true); err == nil {
		return nil
	} else if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return fmt.Errorf("ensure branch %q: resolve: %w", name, err)
	}

	storer := repo.Storer

	tree := &object.Tree{}
	treeObject := storer.NewEncodedObject()
	treeObject.SetType(plumbing.TreeObject)
	if err := tree.Encode(treeObject); err != nil {
		return fmt.Errorf("ensure branch %q: encode tree: %w", name, err)
	}
	treeHash, err := storer.SetEncodedObject(treeObject)
	if err != nil {
		return fmt.Errorf("ensure branch %q: save tree object: %w", name, err)
	}

	now := time.Now()
	signature := object.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  now,
	}
	commit := &object.Commit{
		Author:    signature,
		Committer: signature,
		Message:   "Ensure branch",
		TreeHash:  treeHash,
	}
	commitObject := storer.NewEncodedObject()
	commitObject.SetType(plumbing.CommitObject)
	if err := commit.Encode(commitObject); err != nil {
		return fmt.Errorf("ensure branch %q: encode commit: %w", name, err)
	}
	commitHash, err := storer.SetEncodedObject(commitObject)
	if err != nil {
		return fmt.Errorf("ensure branch %q: save commit object: %w", name, err)
	}

	if err := storer.SetReference(
		plumbing.NewHashReference(branch, commitHash),
	); err != nil {
		return fmt.Errorf("ensure branch %q: set branch to commit %s: %w", name, commitHash, err)
	}

	return nil
}

func newStore(options storeOptions) (*store, error) {
	_, err := ensureRepo(options.archiveDir, true)
	if err != nil {
		return nil, fmt.Errorf("new store: archive: %w", err)
	}

	repo, err := ensureRepo(options.workspaceDir, false)
	if err != nil {
		return nil, fmt.Errorf("new store: workspace: %w", err)
	}

	if err := ensureRemote(repo, remoteName, options.archiveDir); err != nil {
		return nil, fmt.Errorf("new store: %w", err)
	}

	if err := ensureBranch(repo, options.baseBranch); err != nil {
		return nil, fmt.Errorf("new store: base branch: %w", err)
	}

	return &store{
		repo:    repo,
		options: options,
	}, nil
}
