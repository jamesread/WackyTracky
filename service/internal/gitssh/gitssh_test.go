package gitssh

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sshKeygenAvailable() bool {
	_, err := exec.LookPath("ssh-keygen")
	return err == nil
}

func TestSetup_GeneratesKeyAndSymlinks(t *testing.T) {
	if !sshKeygenAvailable() {
		t.Skip("ssh-keygen not available")
	}

	configDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	res, err := Setup(configDir)
	require.NoError(t, err)

	assert.FileExists(t, res.PrivateKeyPath)
	assert.FileExists(t, res.PublicKeyPath)
	assert.NotEmpty(t, res.Fingerprint)
	assert.Contains(t, res.Fingerprint, "SHA256:")

	linkPrivate := filepath.Join(homeDir, ".ssh", keyBaseName)
	linkPublic := linkPrivate + ".pub"
	targetPrivate, err := os.Readlink(linkPrivate)
	require.NoError(t, err)
	assert.Equal(t, res.PrivateKeyPath, targetPrivate)
	targetPublic, err := os.Readlink(linkPublic)
	require.NoError(t, err)
	assert.Equal(t, res.PublicKeyPath, targetPublic)
}

func TestSetup_Idempotent(t *testing.T) {
	if !sshKeygenAvailable() {
		t.Skip("ssh-keygen not available")
	}

	configDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	first, err := Setup(configDir)
	require.NoError(t, err)
	assert.False(t, first.ExistingKey)

	privateBefore, err := os.ReadFile(first.PrivateKeyPath)
	require.NoError(t, err)

	second, err := Setup(configDir)
	require.NoError(t, err)
	assert.True(t, second.ExistingKey)

	privateAfter, err := os.ReadFile(second.PrivateKeyPath)
	require.NoError(t, err)
	assert.Equal(t, privateBefore, privateAfter)
	assert.Equal(t, first.Fingerprint, second.Fingerprint)
	assert.Equal(t, first.PrivateKeyPath, second.PrivateKeyPath)
}

func TestEnsureSymlink_RefusesRegularFile(t *testing.T) {
	homeDir := t.TempDir()
	linkPath := filepath.Join(homeDir, keyBaseName)
	require.NoError(t, os.WriteFile(linkPath, []byte("keep-me"), 0o600))

	err := ensureSymlink("/config/wackytracky_git_sync", linkPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refusing to overwrite")

	contents, err := os.ReadFile(linkPath)
	require.NoError(t, err)
	assert.Equal(t, []byte("keep-me"), contents)
}

func TestConfigureGitCommand_UsesSymlinkedKey(t *testing.T) {
	if !sshKeygenAvailable() {
		t.Skip("ssh-keygen not available")
	}

	configDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	_, err := Setup(configDir)
	require.NoError(t, err)

	cmd := exec.Command("git", "version")
	ConfigureGitCommand(cmd)
	require.NotNil(t, cmd.Env)

	found := false
	for _, entry := range cmd.Env {
		if entry == "GIT_SSH_COMMAND=ssh -i '"+PrivateKeyPath()+"' -o IdentitiesOnly=yes" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected GIT_SSH_COMMAND in cmd.Env: %v", cmd.Env)
}

func TestShellEscape(t *testing.T) {
	assert.Equal(t, "'plain'", shellEscape("plain"))
	assert.Equal(t, "'it'\\''s'", shellEscape("it's"))
}
