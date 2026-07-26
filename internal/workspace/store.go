package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

const (
	authorName  = "cpenv"
	authorEmail = "cpenv@local"
)

type store struct {
	mu         sync.Mutex
	path       string
	baseBranch string
	repo       *git.Repository
}

func (s *store) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	head, err := s.repo.Head()
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return fmt.Errorf("save: unborn branch: %w", err)
	} else if err != nil {
		return fmt.Errorf("save: resolve HEAD: %w", err)
	}

	if !head.Name().IsBranch() {
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

	return nil
}

func (s *store) exists(name string) (bool, error) {
	branch := plumbing.NewBranchReferenceName(name)
	if err := branch.Validate(); err != nil {
		return false, fmt.Errorf("exists: invalid branch name %q: %w", name, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.repo.Reference(branch, true); err == nil {
		return true, nil
	} else if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return false, nil
	} else {
		return false, fmt.Errorf("exists: resolve branch %q: %w", name, err)
	}
}

func (s *store) open(name string) error {
	branch := plumbing.NewBranchReferenceName(name)
	if err := branch.Validate(); err != nil {
		return fmt.Errorf("load: invalid branch name %q: %w", name, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	w, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("load: get worktree: %w", err)
	}

	if err := w.Checkout(&git.CheckoutOptions{Branch: branch}); err != nil {
		return fmt.Errorf("load: checkout %q: %w", name, err)
	}

	if err := w.Clean(&git.CleanOptions{Dir: true}); err != nil {
		return fmt.Errorf("load: clean: %w", err)
	}

	return nil
}

func (s *store) create(name string) error {
	branch := plumbing.NewBranchReferenceName(name)
	if err := branch.Validate(); err != nil {
		return fmt.Errorf("create: invalid branch name %q: %w", name, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	w, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("create: get worktree: %w", err)
	}

	ref, err := s.repo.Reference(plumbing.NewBranchReferenceName(s.baseBranch), true)
	if err != nil {
		return fmt.Errorf("load: resolve branch %q: %w", s.baseBranch, err)
	}

	if err := w.Checkout(&git.CheckoutOptions{
		Branch: branch,
		Create: true,
		Hash:   ref.Hash(),
	}); err != nil {
		return fmt.Errorf("load: checkout %q: %w", name, err)
	}

	if err := w.Clean(&git.CleanOptions{Dir: true}); err != nil {
		return fmt.Errorf("load: clean: %w", err)
	}

	return nil
}

func ensureRepo(path string) (*git.Repository, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("ensure repo %q: path must be absolute", path)
	}

	if repo, err := git.PlainOpen(path); err == nil {
		return repo, nil
	} else if !errors.Is(err, git.ErrRepositoryNotExists) {
		return nil, fmt.Errorf("ensure repo %q: open: %w", path, err)
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("ensure repo %q: mkdir: %w", path, err)
	}

	repo, err := git.PlainInit(path, false)
	if err != nil {
		return nil, fmt.Errorf("ensure repo %q: init: %w", path, err)
	}

	return repo, nil
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

	treeObj := storer.NewEncodedObject()
	treeObj.SetType(plumbing.TreeObject)
	if err := (&object.Tree{}).Encode(treeObj); err != nil {
		return fmt.Errorf("ensure branch %q: encode tree: %w", name, err)
	}

	treeHash, err := storer.SetEncodedObject(treeObj)
	if err != nil {
		return fmt.Errorf("ensure branch %q: save tree object: %w", name, err)
	}

	signature := object.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  time.Now(),
	}
	commit := &object.Commit{
		Author:    signature,
		Committer: signature,
		Message:   "Ensure branch",
		TreeHash:  treeHash,
	}

	commitObj := storer.NewEncodedObject()
	commitObj.SetType(plumbing.CommitObject)
	if err := commit.Encode(commitObj); err != nil {
		return fmt.Errorf("ensure branch %q: encode commit: %w", name, err)
	}

	commitHash, err := storer.SetEncodedObject(commitObj)
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

func newStore(path, baseBranch string) (*store, error) {
	repo, err := ensureRepo(path)
	if err != nil {
		return nil, fmt.Errorf("new store: %w", err)
	}

	if err := ensureBranch(repo, baseBranch); err != nil {
		return nil, fmt.Errorf("new store: base branch: %w", err)
	}

	return &store{
		path:       path,
		baseBranch: baseBranch,
		repo:       repo,
	}, nil
}
