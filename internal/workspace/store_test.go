package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
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

func initRepo(t *testing.T, path string, isBare bool) *git.Repository {
	t.Helper()

	repo, err := git.PlainInit(path, isBare)
	require.NoError(t, err)
	return repo
}

func setRemote(t *testing.T, repo *git.Repository, name, url string) {
	t.Helper()

	_, err := repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	require.NoError(t, err)
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

	now := time.Now()
	hash, err := worktree(t, repo).Commit("commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  now,
		},
	})
	require.NoError(t, err)
	return hash
}

func createBranch(t *testing.T, repo *git.Repository, name string, hash plumbing.Hash) {
	t.Helper()

	branch := plumbing.NewBranchReferenceName(name)
	require.NoError(t, repo.Storer.SetReference(
		plumbing.NewHashReference(branch, hash),
	))
}

func branchCommit(t *testing.T, repo *git.Repository, name string) *object.Commit {
	t.Helper()

	branch := plumbing.NewBranchReferenceName(name)
	ref, err := repo.Reference(branch, true)
	require.NoError(t, err)

	commit, err := repo.CommitObject(ref.Hash())
	require.NoError(t, err)
	return commit
}

func TestStore(t *testing.T) {
	t.Parallel()

	t.Run("core behavior", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		archiveDir := filepath.Join(dir, "archive.git")
		workspaceDir := filepath.Join(dir, "workspace")
		baseBranch := "base"

		store, err := newStore(storeOptions{
			archiveDir:   archiveDir,
			workspaceDir: workspaceDir,
			baseBranch:   baseBranch,
		})
		require.NoError(t, err)

		require.NoError(t, store.load("foo"))
		writeFiles(t, workspaceDir, map[string]string{
			"foo": "foo",
		})
		require.NoError(t, store.save())

		require.NoError(t, store.load("bar"))
		writeFiles(t, workspaceDir, map[string]string{
			"bar": "bar",
		})
		require.NoError(t, store.save())
		writeFiles(t, workspaceDir, map[string]string{
			"baz": "baz",
		})
		require.NoError(t, store.save())

		require.NoError(t, store.load("foo"))
		assert.Equal(t, map[string]string{
			"foo": "foo",
		}, readFiles(t, workspaceDir))

		writeFiles(t, workspaceDir, map[string]string{
			"qux": "qux",
		})
		require.NoError(t, store.save())

		require.NoError(t, store.load("bar"))
		assert.Equal(t, map[string]string{
			"bar": "bar",
			"baz": "baz",
		}, readFiles(t, workspaceDir))

		require.NoError(t, store.load("foo"))
		assert.Equal(t, map[string]string{
			"foo": "foo",
			"qux": "qux",
		}, readFiles(t, workspaceDir))
	})

	t.Run("base branch behavior", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		archiveDir := filepath.Join(dir, "archive.git")
		workspaceDir := filepath.Join(dir, "workspace")
		baseBranch := "base"

		store, err := newStore(storeOptions{
			archiveDir:   archiveDir,
			workspaceDir: workspaceDir,
			baseBranch:   baseBranch,
		})
		require.NoError(t, err)

		require.NoError(t, store.load("foo"))
		assert.Equal(t, map[string]string{}, readFiles(t, workspaceDir))

		require.NoError(t, store.load(baseBranch))
		writeFiles(t, workspaceDir, map[string]string{
			"base": "base",
		})
		require.NoError(t, store.save())

		require.NoError(t, store.load("bar"))
		assert.Equal(t, map[string]string{
			"base": "base",
		}, readFiles(t, workspaceDir))

		require.NoError(t, store.load("foo"))
		assert.Equal(t, map[string]string{}, readFiles(t, workspaceDir))
	})

	t.Run("ensure repo", func(t *testing.T) {
		t.Parallel()

		t.Run("match", func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			repoDir := filepath.Join(dir, "repo")

			_ = initRepo(t, repoDir, false)

			_, err := ensureRepo(repoDir, false)
			assert.NoError(t, err)
		})

		t.Run("mismatch", func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			repoDir := filepath.Join(dir, "repo")

			_ = initRepo(t, repoDir, true)

			_, err := ensureRepo(repoDir, false)
			require.Error(t, err)
			assert.ErrorContains(t, err, "bare: expected false, got true")
		})
	})

	t.Run("ensure remote", func(t *testing.T) {
		t.Parallel()

		t.Run("match", func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			remoteDir := filepath.Join(dir, "remote.git")
			localDir := filepath.Join(dir, "local")

			_ = initRepo(t, remoteDir, true)
			repo := initRepo(t, localDir, false)
			setRemote(t, repo, remoteName, remoteDir)

			err := ensureRemote(repo, remoteName, remoteDir)
			assert.NoError(t, err)
		})

		t.Run("mismatch", func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			remoteDir1 := filepath.Join(dir, "remote1.git")
			remoteDir2 := filepath.Join(dir, "remote2.git")
			localDir := filepath.Join(dir, "local")

			_ = initRepo(t, remoteDir1, true)
			_ = initRepo(t, remoteDir2, true)
			repo := initRepo(t, localDir, false)
			setRemote(t, repo, remoteName, remoteDir1)

			err := ensureRemote(repo, remoteName, remoteDir2)
			require.Error(t, err)
			assert.ErrorContains(t, err, fmt.Sprintf("expected %q, got %q", remoteDir2, remoteDir1))
		})
	})

	t.Run("ensure branch", func(t *testing.T) {
		t.Parallel()

		t.Run("is idempotent", func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			remoteDir := filepath.Join(dir, "remote.git")
			localDir := filepath.Join(dir, "local")
			baseBranch := "base"

			remoteRepo := initRepo(t, remoteDir, true)
			localRepo := initRepo(t, localDir, false)
			setRemote(t, localRepo, remoteName, remoteDir)
			writeFiles(t, localDir, map[string]string{
				"foo": "foo",
			})
			addAll(t, localRepo)
			hash := commit(t, localRepo)
			createBranch(t, localRepo, baseBranch, hash)

			require.NoError(t, ensureBranch(localRepo, baseBranch))

			localCommit := branchCommit(t, localRepo, baseBranch)
			assert.Equal(t, hash, localCommit.Hash)

			remoteCommit := branchCommit(t, remoteRepo, baseBranch)
			assert.Equal(t, hash, remoteCommit.Hash)
		})

		t.Run("is isolated", func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			remoteDir := filepath.Join(dir, "remote.git")
			localDir := filepath.Join(dir, "local")
			baseBranch := "base"

			remoteRepo := initRepo(t, remoteDir, true)
			localRepo := initRepo(t, localDir, false)
			setRemote(t, localRepo, remoteName, remoteDir)
			writeFiles(t, localDir, map[string]string{
				"foo": "foo",
			})
			addAll(t, localRepo)
			_ = commit(t, localRepo)
			writeFiles(t, localDir, map[string]string{
				"bar": "bar",
			})
			addAll(t, localRepo)
			writeFiles(t, localDir, map[string]string{
				"baz": "baz",
			})

			headBefore, err := localRepo.Head()
			require.NoError(t, err)
			statusBefore, err := worktree(t, localRepo).Status()
			require.NoError(t, err)
			filesBefore := readFiles(t, localDir)

			require.NoError(t, ensureBranch(localRepo, baseBranch))

			headAfter, err := localRepo.Head()
			require.NoError(t, err)
			statusAfter, err := worktree(t, localRepo).Status()
			require.NoError(t, err)
			filesAfter := readFiles(t, localDir)

			assert.Equal(t, headBefore, headAfter)
			assert.Equal(t, statusBefore, statusAfter)
			assert.Equal(t, filesBefore, filesAfter)

			localCommit := branchCommit(t, localRepo, baseBranch)
			tree, err := localCommit.Tree()
			require.NoError(t, err)
			assert.Empty(t, tree.Entries)

			remoteCommit := branchCommit(t, remoteRepo, baseBranch)
			assert.Equal(t, localCommit.Hash, remoteCommit.Hash)
		})
	})
}
