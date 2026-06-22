package gitssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	keyBaseName   = "wackytracky_git_sync"
	keyComment    = "wacky-tracky-git-sync"
	containerHome = "/home/wt"
	gitUserName   = "wt"
	gitUserEmail  = "wt@container"
)

type SetupResult struct {
	PrivateKeyPath string
	PublicKeyPath  string
	Fingerprint    string
	ExistingKey    bool
}

func ResolveConfigDir() string {
	if f := viper.ConfigFileUsed(); f != "" {
		return filepath.Dir(f)
	}
	if info, err := os.Stat("/config"); err == nil && info.IsDir() {
		return "/config"
	}
	return "."
}

func Setup(configDir string) (SetupResult, error) {
	privateKey := filepath.Join(configDir, keyBaseName)
	publicKey := privateKey + ".pub"
	existing, err := ensureKeyPair(privateKey)
	if err != nil {
		return SetupResult{}, err
	}
	if err := linkKeysToSSHDir(privateKey, publicKey); err != nil {
		return SetupResult{}, err
	}
	fingerprint, err := keyFingerprint(publicKey)
	if err != nil {
		return SetupResult{}, err
	}
	return SetupResult{
		PrivateKeyPath: privateKey,
		PublicKeyPath:  publicKey,
		Fingerprint:    fingerprint,
		ExistingKey:    existing,
	}, nil
}

func SetupAndLog() {
	if err := ensureGitIdentity(); err != nil {
		log.Warnf("Git identity setup failed: %v", err)
	}
	dir := ResolveConfigDir()
	res, err := Setup(dir)
	if err != nil {
		log.Warnf("Git SSH key setup failed: %v", err)
		return
	}
	if res.ExistingKey {
		log.Infof("Git sync SSH key already exists; not overwriting %s", res.PrivateKeyPath)
	} else {
		log.Infof("Generated new git sync SSH key at %s", res.PrivateKeyPath)
	}
	log.Infof("Git sync SSH key fingerprint (add public key to remote authorized_keys or deploy keys): %s", res.Fingerprint)
	log.Infof("Git sync SSH public key file: %s", res.PublicKeyPath)
}

func PrivateKeyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", keyBaseName)
}

func ConfigureGitCommand(cmd *exec.Cmd) {
	keyPath := PrivateKeyPath()
	if _, err := os.Stat(keyPath); err != nil {
		return
	}
	cmd.Env = append(os.Environ(),
		"GIT_SSH_COMMAND=ssh -i "+shellEscape(keyPath)+" -o IdentitiesOnly=yes",
	)
}

func ensureKeyPair(privateKey string) (bool, error) {
	if _, err := os.Stat(privateKey); err == nil {
		return true, nil
	}
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", privateKey, "-N", "", "-C", keyComment)
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("ssh-keygen: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return false, nil
}

func linkKeysToSSHDir(privateKey, publicKey string) error {
	sshDir, err := sshHomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", sshDir, err)
	}
	if err := ensureSymlink(privateKey, filepath.Join(sshDir, keyBaseName)); err != nil {
		return err
	}
	return ensureSymlink(publicKey, filepath.Join(sshDir, keyBaseName+".pub"))
}

func sshHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".ssh"), nil
}

func ensureSymlink(target, linkPath string) error {
	if sameSymlink(linkPath, target) {
		return nil
	}
	if err := refuseOverwriteExistingFile(linkPath); err != nil {
		return err
	}
	return replaceSymlink(target, linkPath)
}

func replaceSymlink(target, linkPath string) error {
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", linkPath, err)
	}
	if err := os.Symlink(target, linkPath); err != nil {
		return fmt.Errorf("symlink %s -> %s: %w", linkPath, target, err)
	}
	return nil
}

func refuseOverwriteExistingFile(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	return fmt.Errorf("refusing to overwrite existing file %s", path)
}

func sameSymlink(linkPath, target string) bool {
	current, err := os.Readlink(linkPath)
	if err != nil {
		return false
	}
	want, err := filepath.EvalSymlinks(target)
	if err != nil {
		want = target
	}
	got, err := filepath.EvalSymlinks(linkPath)
	if err != nil {
		got = current
	}
	return got == want || current == target
}

func keyFingerprint(publicKey string) (string, error) {
	out, err := exec.Command("ssh-keygen", "-lf", publicKey).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ssh-keygen -lf: %s: %w", strings.TrimSpace(string(out)), err)
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return "", fmt.Errorf("ssh-keygen -lf returned no output")
	}
	return line, nil
}

func shellEscape(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func ensureGitIdentity() error {
	if !isContainerHome() {
		return nil
	}
	if err := setGlobalGitConfig("user.name", gitUserName); err != nil {
		return err
	}
	return setGlobalGitConfig("user.email", gitUserEmail)
}

func isContainerHome() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return home == containerHome
}

func setGlobalGitConfig(key, value string) error {
	cmd := exec.Command("git", "config", "--global", key, value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git config --global %s: %s: %w", key, strings.TrimSpace(string(out)), err)
	}
	return nil
}
