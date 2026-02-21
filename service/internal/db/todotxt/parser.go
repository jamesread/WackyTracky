package todotxt

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

// Task represents a parsed todo.txt task (in-memory).
type Task struct {
	Completed      bool
	CompletionDate *time.Time
	CreationDate   *time.Time
	Priority       string // (A)..(Z) or empty
	Description    string // body text
	Projects       []string
	Contexts       []string
	Tags           []string // #hashtag parsed as tag (e.g. #bar -> "bar")
	Metadata       map[string]string
}

// parseCompletedPrefix parses "x " and optional completion + creation dates; returns updated task and remaining rest.
func parseCompletedPrefix(t *Task, rest string) string {
	if !strings.HasPrefix(rest, "x ") {
		return rest
	}
	t.Completed = true
	rest = strings.TrimSpace(rest[2:])
	rest = parseOptionalDateInto(&t.CompletionDate, rest)
	rest = parseOptionalDateInto(&t.CreationDate, rest)
	return rest
}

func parseOptionalDateInto(dst **time.Time, s string) string {
	if len(s) < 10 || !isDate(s[:10]) {
		return s
	}
	if d, err := time.Parse("2006-01-02", s[:10]); err == nil {
		*dst = &d
	}
	return strings.TrimSpace(s[10:])
}

func isValidPriorityChar(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func hasValidPriorityPrefix(rest string) bool {
	return len(rest) >= 4 && rest[0] == '(' && rest[2] == ')' && rest[3] == ' ' && isValidPriorityChar(rest[1])
}

func parsePriorityPrefix(rest string) (priority string, restOut string) {
	if !hasValidPriorityPrefix(rest) {
		return "", rest
	}
	return rest[:4], strings.TrimSpace(rest[4:])
}

func parseCreationDateAtStart(rest string) (creation *time.Time, restOut string) {
	if len(rest) < 10 || !isDate(rest[:10]) {
		return nil, rest
	}
	if d, err := time.Parse("2006-01-02", rest[:10]); err == nil {
		return &d, strings.TrimSpace(rest[10:])
	}
	return nil, rest
}

// ParseLine parses a single todo.txt line per http://todotxt.org / GitHub todo.txt format.
// - Completed: line starts with "x " (lowercase x + space).
// - Completion date: YYYY-MM-DD immediately after "x ".
// - Creation date: YYYY-MM-DD after priority (or at start if no priority).
// - Priority: (A)..(Z) at start (incomplete) or preserved in key:value when complete.
// - Projects: +word, Contexts: @word.
// - key:value for extra metadata (id, listid, due:, etc.).
func parseLineIncompletePrefix(t *Task, rest string) string {
	if t.Completed {
		return rest
	}
	pri, restAfter := parsePriorityPrefix(rest)
	if pri != "" {
		t.Priority = pri
		rest = restAfter
	}
	return rest
}

func ParseLine(line string) *Task {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	t := &Task{Metadata: make(map[string]string)}
	rest := parseCompletedPrefix(t, line)
	if rest == "" {
		return t
	}
	rest = parseLineIncompletePrefix(t, rest)
	if creation, restAfter := parseCreationDateAtStart(rest); creation != nil {
		t.CreationDate = creation
		rest = restAfter
	}
	t.Description, t.Projects, t.Contexts, t.Tags, t.Metadata = parseBody(rest)
	return t
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

var dateDigitIndices = []int{0, 1, 2, 3, 5, 6, 8, 9}

func allDigitsAt(s string, indices []int) bool {
	for _, i := range indices {
		if !isDigit(s[i]) {
			return false
		}
	}
	return true
}

func isDate(s string) bool {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return false
	}
	return allDigitsAt(s, dateDigitIndices)
}

var (
	projectRe = regexp.MustCompile(`\+(\S+)`)
	contextRe = regexp.MustCompile(`@(\S+)`)
	hashtagRe = regexp.MustCompile(`#(\S+)`) // #bar -> tag "bar"
	kvRe      = regexp.MustCompile(`(\S+):(\S+)`)
)

func extractUnique(seen map[string]bool, matches [][]string, subIdx int) []string {
	var out []string
	for _, m := range matches {
		v := m[subIdx]
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func parseBody(body string) (description string, projects, contexts, tags []string, metadata map[string]string) {
	metadata = make(map[string]string)
	for _, m := range kvRe.FindAllStringSubmatch(body, -1) {
		metadata[m[1]] = m[2]
	}
	body = kvRe.ReplaceAllString(body, "")
	projects = extractUnique(make(map[string]bool), projectRe.FindAllStringSubmatch(body, -1), 1)
	contexts = extractUnique(make(map[string]bool), contextRe.FindAllStringSubmatch(body, -1), 1)
	tags = extractUnique(make(map[string]bool), hashtagRe.FindAllStringSubmatch(body, -1), 1)
	body = projectRe.ReplaceAllString(body, "")
	body = contextRe.ReplaceAllString(body, "")
	body = hashtagRe.ReplaceAllString(body, "")
	body = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(body), " ")
	description = strings.TrimSpace(body)
	return
}

// ParseContent parses a content string (no leading "x " or priority/dates) into description, projects, contexts, tags, and metadata.
// Used when updating a task's body without changing id/listid.
func ParseContent(content string) (description string, projects, contexts, tags []string, metadata map[string]string) {
	return parseBody(strings.TrimSpace(content))
}

func appendDatePart(parts []string, d *time.Time) []string {
	if d == nil {
		return parts
	}
	return append(parts, d.Format("2006-01-02"))
}

func formatLineHeaderParts(t *Task) []string {
	if t.Completed {
		parts := []string{"x"}
		parts = appendDatePart(parts, t.CompletionDate)
		parts = appendDatePart(parts, t.CreationDate)
		return parts
	}
	var parts []string
	if t.Priority != "" {
		parts = append(parts, t.Priority)
	}
	return appendDatePart(parts, t.CreationDate)
}

func appendSortedWithPrefix(desc string, prefix string, items []string) string {
	for _, p := range sortedCopy(items) {
		desc += " " + prefix + p
	}
	return desc
}

func formatLineDescriptionWithMeta(t *Task) string {
	desc := appendSortedWithPrefix(t.Description, "+", t.Projects)
	desc = appendSortedWithPrefix(desc, "@", t.Contexts)
	desc = appendSortedWithPrefix(desc, "#", t.Tags)
	metaKeys := make([]string, 0, len(t.Metadata))
	for k := range t.Metadata {
		metaKeys = append(metaKeys, k)
	}
	sort.Strings(metaKeys)
	for _, k := range metaKeys {
		desc += " " + k + ":" + t.Metadata[k]
	}
	return strings.TrimSpace(desc)
}

func sortedCopy(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}

// FormatLine serializes a Task back to a todo.txt line (with id: and listid: if present).
func FormatLine(t *Task) string {
	parts := formatLineHeaderParts(t)
	desc := formatLineDescriptionWithMeta(t)
	if desc != "" {
		parts = append(parts, desc)
	}
	return strings.Join(parts, " ")
}
