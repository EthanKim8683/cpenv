package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()

	for path, content := range files {
		path = filepath.Join(dir, path)
		path = filepath.FromSlash(path)

		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}
}

func readFiles(t *testing.T, dir string) map[string]string {
	t.Helper()

	files := make(map[string]string)
	require.NoError(t, filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		require.NoError(t, err)

		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		require.NoError(t, err)

		content, err := os.ReadFile(path)
		require.NoError(t, err)

		files[filepath.ToSlash(relPath)] = string(content)
		return nil
	}))
	return files
}

func initRepo(t *testing.T, path string) *git.Repository {
	t.Helper()

	repo, err := git.PlainInit(path, false)
	require.NoError(t, err)
	return repo
}

func worktree(t *testing.T, repo *git.Repository) *git.Worktree {
	t.Helper()

	w, err := repo.Worktree()
	require.NoError(t, err)
	return w
}

func addAll(t *testing.T, repo *git.Repository) {
	t.Helper()

	require.NoError(t, worktree(t, repo).AddWithOptions(&git.AddOptions{All: true}))
}

func commit(t *testing.T, repo *git.Repository) plumbing.Hash {
	t.Helper()

	hash, err := worktree(t, repo).Commit("commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
	return hash
}

func createBranch(t *testing.T, repo *git.Repository, name string, hash plumbing.Hash) {
	t.Helper()

	require.NoError(t, repo.Storer.SetReference(
		plumbing.NewHashReference(plumbing.NewBranchReferenceName(name), hash),
	))
}

func branchCommit(t *testing.T, repo *git.Repository, name string) *object.Commit {
	t.Helper()

	ref, err := repo.Reference(plumbing.NewBranchReferenceName(name), true)
	require.NoError(t, err)

	commit, err := repo.CommitObject(ref.Hash())
	require.NoError(t, err)
	return commit
}

func TestStore(t *testing.T) {
	t.Parallel()

	t.Run("core behavior", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "repo")
		baseBranch := "base"

		store, err := newStore(storeOptions{
			path:       path,
			baseBranch: baseBranch,
		})
		require.NoError(t, err)

		require.NoError(t, store.load("foo"))
		writeFiles(t, path, map[string]string{
			"foo": "foo",
		})
		require.NoError(t, store.save())

		require.NoError(t, store.load("bar"))
		writeFiles(t, path, map[string]string{
			"bar": "bar",
		})
		require.NoError(t, store.save())
		writeFiles(t, path, map[string]string{
			"baz": "baz",
		})
		require.NoError(t, store.save())

		require.NoError(t, store.load("foo"))
		assert.Equal(t, map[string]string{
			"foo": "foo",
		}, readFiles(t, path))

		writeFiles(t, path, map[string]string{
			"qux": "qux",
		})
		require.NoError(t, store.save())

		require.NoError(t, store.load("bar"))
		assert.Equal(t, map[string]string{
			"bar": "bar",
			"baz": "baz",
		}, readFiles(t, path))

		require.NoError(t, store.load("foo"))
		assert.Equal(t, map[string]string{
			"foo": "foo",
			"qux": "qux",
		}, readFiles(t, path))
	})

	t.Run("base branch behavior", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "repo")
		baseBranch := "base"

		store, err := newStore(storeOptions{
			path:       path,
			baseBranch: baseBranch,
		})
		require.NoError(t, err)

		require.NoError(t, store.load("foo"))
		assert.Equal(t, map[string]string{}, readFiles(t, path))

		require.NoError(t, store.load(baseBranch))
		writeFiles(t, path, map[string]string{
			"base": "base",
		})
		require.NoError(t, store.save())

		require.NoError(t, store.load("bar"))
		assert.Equal(t, map[string]string{
			"base": "base",
		}, readFiles(t, path))

		require.NoError(t, store.load("foo"))
		assert.Equal(t, map[string]string{}, readFiles(t, path))
	})

	t.Run("ensure repo", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "repo")

		repo := initRepo(t, path)
		writeFiles(t, path, map[string]string{
			"foo": "foo",
		})
		addAll(t, repo)
		hash := commit(t, repo)

		ensuredRepo, err := ensureRepo(path)
		require.NoError(t, err)

		head, err := ensuredRepo.Head()
		require.NoError(t, err)
		assert.Equal(t, hash, head.Hash())
	})

	t.Run("ensure branch", func(t *testing.T) {
		t.Parallel()

		t.Run("is idempotent", func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "repo")
			baseBranch := "base"

			repo := initRepo(t, path)
			writeFiles(t, path, map[string]string{
				"foo": "foo",
			})
			addAll(t, repo)
			hash := commit(t, repo)
			createBranch(t, repo, baseBranch, hash)

			require.NoError(t, ensureBranch(repo, baseBranch))

			commit := branchCommit(t, repo, baseBranch)
			assert.Equal(t, hash, commit.Hash)
		})

		t.Run("is isolated", func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "repo")
			baseBranch := "base"

			repo := initRepo(t, path)
			writeFiles(t, path, map[string]string{
				"foo": "foo",
			})
			addAll(t, repo)
			_ = commit(t, repo)
			writeFiles(t, path, map[string]string{
				"bar": "bar",
			})
			addAll(t, repo)
			writeFiles(t, path, map[string]string{
				"baz": "baz",
			})

			headBefore, err := repo.Head()
			require.NoError(t, err)
			statusBefore, err := worktree(t, repo).Status()
			require.NoError(t, err)
			filesBefore := readFiles(t, path)

			require.NoError(t, ensureBranch(repo, baseBranch))

			headAfter, err := repo.Head()
			require.NoError(t, err)
			statusAfter, err := worktree(t, repo).Status()
			require.NoError(t, err)
			filesAfter := readFiles(t, path)

			assert.Equal(t, headBefore, headAfter)
			assert.Equal(t, statusBefore, statusAfter)
			assert.Equal(t, filesBefore, filesAfter)

			commit := branchCommit(t, repo, baseBranch)
			tree, err := commit.Tree()
			require.NoError(t, err)
			assert.Empty(t, tree.Entries)
		})
	})
}
