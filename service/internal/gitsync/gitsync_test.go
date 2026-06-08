package gitsync

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func isCI() bool {
	return os.Getenv("CI") != ""
}

func TestCommitMessage(t *testing.T) {
	assert.Equal(t, "wt sync: homelab", commitMessage("homelab"))
	assert.Equal(t, "wt sync: homelab", commitMessage("  homelab  "))
	assert.Equal(t, "wt sync", commitMessage(""))
	assert.Equal(t, "wt sync: prod east", commitMessage("prod\neast"))
}

func TestSync_NotGitRepo(t *testing.T) {
	dir := t.TempDir()
	res := Sync(context.Background(), dir, "test")
	assert.False(t, res.Success)
	assert.Contains(t, res.Message, "Not a git repository")
}

func TestSync_EmptyDir(t *testing.T) {
	res := Sync(context.Background(), "", "test")
	assert.False(t, res.Success)
	assert.Contains(t, res.Message, "No todotxt directory")
}

func TestSync_AlreadyUpToDate(t *testing.T) {
	dir := initSyncedRepo(t)
	res := Sync(context.Background(), dir, "test")
	assert.True(t, res.Success)
	assert.Equal(t, "Already up to date", res.Message)
}

func TestSync_CommitAndPush(t *testing.T) {
	if isCI() {
		t.Skip("Skipping push test in CI environment")
	}

	dir, remote := initRepoWithRemote(t)
	runGit(t, dir, "push", "-u", "origin", "main")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "todo.txt"), []byte("new task\n"), 0o644))

	res := Sync(context.Background(), dir, "homelab")
	assert.True(t, res.Success, res.Message)
	assert.Equal(t, "Changes pushed to remote", res.Message)

	out, err := exec.Command("git", "-C", remote, "log", "-1", "--format=%s").CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "wt sync: homelab", strings.TrimSpace(string(out)))
}

func TestSync_MergeConflict(t *testing.T) {
	dir := initRepoWithConflict(t)
	res := Sync(context.Background(), dir, "test")
	assert.False(t, res.Success)
	assert.Contains(t, res.Message, "Merge in progress")
}

func TestSync_BehindRemote(t *testing.T) {
	if isCI() {
		t.Skip("Skipping behind remote test in CI environment")
	}

	dir, remote := initRepoWithRemote(t)
	runGit(t, dir, "push", "-u", "origin", "main")

	remoteClone := filepath.Join(t.TempDir(), "clone")
	parent := t.TempDir()
	runGit(t, parent, "clone", remote, remoteClone)
	runGit(t, remoteClone, "config", "user.email", "remote@example.com")
	runGit(t, remoteClone, "config", "user.name", "Remote User")
	require.NoError(t, os.WriteFile(filepath.Join(remoteClone, "todo.txt"), []byte("remote change\n"), 0o644))
	runGit(t, remoteClone, "add", "todo.txt")
	runGit(t, remoteClone, "commit", "-m", "remote")
	runGit(t, remoteClone, "push")

	res := Sync(context.Background(), dir, "test")
	assert.False(t, res.Success)
	assert.Contains(t, res.Message, "behind remote")
}

func TestSync_NoUpstream(t *testing.T) {
	dir := initLocalRepo(t)
	res := Sync(context.Background(), dir, "test")
	assert.False(t, res.Success)
	assert.Contains(t, res.Message, "No upstream branch")
}

func TestSync_DetachedHead(t *testing.T) {
	dir, _ := initRepoWithRemote(t)
	runGit(t, dir, "checkout", "--detach", "HEAD")
	res := Sync(context.Background(), dir, "test")
	assert.False(t, res.Success)
	assert.Contains(t, res.Message, "Detached HEAD")
}

func initLocalRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "todo.txt"), []byte("task\n"), 0o644))
	runGit(t, dir, "add", "todo.txt")
	runGit(t, dir, "commit", "-m", "init")
	return dir
}

func initSyncedRepo(t *testing.T) string {
	t.Helper()
	dir, remote := initRepoWithRemote(t)
	runGit(t, dir, "push", "-u", "origin", "main")
	runGit(t, remote, "config", "receive.denyCurrentBranch", "updateInstead")
	return dir
}

func initRepoWithRemote(t *testing.T) (workdir, remoteDir string) {
	t.Helper()
	remoteDir = t.TempDir()
	runGit(t, remoteDir, "init", "--bare")

	workdir = initLocalRepo(t)
	runGit(t, workdir, "branch", "-M", "main")
	runGit(t, workdir, "remote", "add", "origin", remoteDir)
	return workdir, remoteDir
}

func initRepoWithConflict(t *testing.T) string {
	t.Helper()
	dir := initLocalRepo(t)
	runGit(t, dir, "branch", "-M", "main")
	runGit(t, dir, "checkout", "-b", "other")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "todo.txt"), []byte("other branch\n"), 0o644))
	runGit(t, dir, "add", "todo.txt")
	runGit(t, dir, "commit", "-m", "other")
	runGit(t, dir, "checkout", "main")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "todo.txt"), []byte("main branch\n"), 0o644))
	runGit(t, dir, "add", "todo.txt")
	runGit(t, dir, "commit", "-m", "main")
	err := exec.Command("git", "-C", dir, "merge", "other").Run()
	require.Error(t, err)
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}
