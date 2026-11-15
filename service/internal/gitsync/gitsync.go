package gitsync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/wacky-tracky/wacky-tracky-server/internal/gitssh"
)

type Result struct {
	Message string
	Steps   []string
	Success bool
}

type gitRunner struct {
	dir string
}

func (r *gitRunner) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", r.dir}, args...)...)
	gitssh.ConfigureGitCommand(cmd)
	b, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(b))
	if err != nil {
		if out != "" {
			return out, fmt.Errorf("%s", out)
		}
		return out, err
	}
	return out, nil
}

func Sync(ctx context.Context, dir, serverName string) Result {
	res := Result{Steps: []string{}}
	runner, msg := prepareRunner(ctx, dir)
	if msg != "" {
		return fail(res, msg)
	}
	if msg := validateBeforeFetch(ctx, runner); msg != "" {
		return fail(res, msg)
	}
	res.Steps = append(res.Steps, "Validated repository state")
	if msg := fetchAndCheckRemote(ctx, runner); msg != "" {
		return fail(res, msg)
	}
	res.Steps = append(res.Steps, "Fetched remote updates")
	return finalizeSync(ctx, runner, res, serverName)
}

func prepareRunner(ctx context.Context, dir string) (*gitRunner, string) {
	if strings.TrimSpace(dir) == "" {
		return nil, "No todotxt directory configured"
	}
	gitDir, err := resolveGitDir(ctx, dir)
	if err != nil {
		return nil, "Not a git repository"
	}
	if msg := checkInProgress(gitDir); msg != "" {
		return nil, msg
	}
	return &gitRunner{dir: dir}, ""
}

func validateBeforeFetch(ctx context.Context, runner *gitRunner) string {
	if msg := checkConflicts(ctx, runner); msg != "" {
		return msg
	}
	if msg := checkAttachedHead(ctx, runner); msg != "" {
		return msg
	}
	return checkUpstream(ctx, runner)
}

func fetchAndCheckRemote(ctx context.Context, runner *gitRunner) string {
	if _, err := runner.run(ctx, "fetch"); err != nil {
		return "git fetch failed: " + err.Error()
	}
	return checkBehindRemote(ctx, runner)
}

func finalizeSync(ctx context.Context, runner *gitRunner, res Result, serverName string) Result {
	if _, err := runner.run(ctx, "add", "-A"); err != nil {
		return fail(res, "git add failed: "+err.Error())
	}
	res.Steps = append(res.Steps, "Staged changes")
	hasStaged, err := hasStagedChanges(ctx, runner)
	if err != nil {
		return fail(res, "Could not inspect staged changes: "+err.Error())
	}
	if hasStaged {
		if _, err := runner.run(ctx, "commit", "-m", commitMessage(serverName)); err != nil {
			return fail(res, "git commit failed: "+err.Error())
		}
		res.Steps = append(res.Steps, "Committed changes")
	} else {
		res.Steps = append(res.Steps, "No changes to commit")
	}
	return pushIfNeeded(ctx, runner, res)
}

func pushIfNeeded(ctx context.Context, runner *gitRunner, res Result) Result {
	ahead, err := commitsAhead(ctx, runner)
	if err != nil {
		return fail(res, "Could not compare with remote: "+err.Error())
	}
	if ahead == 0 {
		res.Success = true
		res.Message = "Already up to date"
		return res
	}
	if _, err := runner.run(ctx, "push"); err != nil {
		return fail(res, "git push failed: "+err.Error())
	}
	res.Steps = append(res.Steps, "Pushed to remote")
	res.Success = true
	res.Message = "Changes pushed to remote"
	return res
}

func fail(res Result, message string) Result {
	res.Success = false
	res.Message = message
	return res
}

func commitMessage(serverName string) string {
	name := strings.TrimSpace(serverName)
	name = strings.ReplaceAll(name, "\n", " ")
	name = strings.ReplaceAll(name, "\r", " ")
	if name == "" {
		return "wt sync"
	}
	return "wt sync: " + name
}

var errNotGitRepository = errors.New("not a git repository")

func resolveGitDir(ctx context.Context, dir string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--git-dir").CombinedOutput()
	if err != nil {
		return "", errNotGitRepository
	}
	gitDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(dir, gitDir)
	}
	return gitDir, nil
}

func checkInProgress(gitDir string) string {
	markers := []struct {
		path string
		msg  string
	}{
		{filepath.Join(gitDir, "MERGE_HEAD"), "Merge in progress; resolve or abort before syncing"},
		{filepath.Join(gitDir, "CHERRY_PICK_HEAD"), "Cherry-pick in progress; resolve or abort before syncing"},
		{filepath.Join(gitDir, "REBASE_HEAD"), "Rebase in progress; resolve or abort before syncing"},
		{filepath.Join(gitDir, "rebase-merge"), "Rebase in progress; resolve or abort before syncing"},
		{filepath.Join(gitDir, "rebase-apply"), "Rebase in progress; resolve or abort before syncing"},
	}
	for _, marker := range markers {
		if _, err := os.Stat(marker.path); err == nil {
			return marker.msg
		}
	}
	return ""
}

func checkConflicts(ctx context.Context, runner *gitRunner) string {
	out, err := runner.run(ctx, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return "Could not inspect merge conflicts: " + err.Error()
	}
	if strings.TrimSpace(out) != "" {
		return "Unresolved merge conflicts; resolve before syncing"
	}
	return checkPorcelainConflicts(ctx, runner)
}

func checkPorcelainConflicts(ctx context.Context, runner *gitRunner) string {
	status, err := runner.run(ctx, "status", "--porcelain")
	if err != nil {
		return "Could not inspect repository status: " + err.Error()
	}
	if slices.ContainsFunc(strings.Split(status, "\n"), porcelainLineConflicted) {
		return "Unresolved merge conflicts; resolve before syncing"
	}
	return ""
}

func porcelainLineConflicted(line string) bool {
	if len(line) < 2 {
		return false
	}
	code := line[:2]
	return strings.ContainsAny(code, "U") || code == "AA" || code == "DD"
}

func checkAttachedHead(ctx context.Context, runner *gitRunner) string {
	if _, err := runner.run(ctx, "symbolic-ref", "-q", "HEAD"); err != nil {
		return "Detached HEAD; checkout a branch before syncing"
	}
	return ""
}

func checkUpstream(ctx context.Context, runner *gitRunner) string {
	if _, err := runner.run(ctx, "rev-parse", "--abbrev-ref", "@{upstream}"); err != nil {
		return "No upstream branch configured; set upstream before syncing"
	}
	return ""
}

func checkBehindRemote(ctx context.Context, runner *gitRunner) string {
	behind, err := revListCount(ctx, runner, "HEAD..@{upstream}")
	if err != nil {
		return "Could not compare with remote: " + err.Error()
	}
	if behind > 0 {
		return fmt.Sprintf("Local branch is %d commit(s) behind remote; pull before syncing", behind)
	}
	return ""
}

func hasStagedChanges(ctx context.Context, runner *gitRunner) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", runner.dir, "diff", "--cached", "--quiet")
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) {
		return true, nil
	}
	return false, err
}

func commitsAhead(ctx context.Context, runner *gitRunner) (int, error) {
	return revListCount(ctx, runner, "@{upstream}..HEAD")
}

func revListCount(ctx context.Context, runner *gitRunner, rangeSpec string) (int, error) {
	out, err := runner.run(ctx, "rev-list", "--count", rangeSpec)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, fmt.Errorf("invalid rev-list count: %q", out)
	}
	return count, nil
}
