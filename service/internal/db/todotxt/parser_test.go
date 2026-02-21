package todotxt

import (
	"strings"
	"testing"
	"time"
)

func TestParseLine_EmptyOrWhitespace(t *testing.T) {
	if ParseLine("") != nil {
		t.Error("ParseLine(\"\") should return nil")
	}
	if ParseLine("   ") != nil {
		t.Error("ParseLine(whitespace) should return nil")
	}
}

func TestParseLine_IncompleteSimple(t *testing.T) {
	task := ParseLine("buy milk")
	if task == nil {
		t.Fatal("expected task")
	}
	if task.Completed {
		t.Error("expected incomplete")
	}
	if task.Description != "buy milk" {
		t.Errorf("description: got %q", task.Description)
	}
}

func TestParseLine_WithPriority(t *testing.T) {
	task := ParseLine("(A) call mom")
	if task == nil {
		t.Fatal("expected task")
	}
	// Parser stores priority as 4-char token "(A) "
	if len(task.Priority) != 4 || task.Priority[0] != '(' || task.Priority[1] != 'A' {
		t.Errorf("priority: got %q", task.Priority)
	}
	if task.Description != "call mom" {
		t.Errorf("description: got %q", task.Description)
	}
}

func TestParseLine_Completed(t *testing.T) {
	task := ParseLine("x 2024-01-15 buy milk")
	if task == nil {
		t.Fatal("expected task")
	}
	if !task.Completed {
		t.Error("expected completed")
	}
	if task.CompletionDate == nil || task.CompletionDate.Format("2006-01-02") != "2024-01-15" {
		t.Errorf("completion date: got %v", task.CompletionDate)
	}
	if task.Description != "buy milk" {
		t.Errorf("description: got %q", task.Description)
	}
}

func TestParseLine_ProjectsContextsTags(t *testing.T) {
	task := ParseLine("task +project @work #urgent")
	if task == nil {
		t.Fatal("expected task")
	}
	if len(task.Projects) != 1 || task.Projects[0] != "project" {
		t.Errorf("projects: got %v", task.Projects)
	}
	if len(task.Contexts) != 1 || task.Contexts[0] != "work" {
		t.Errorf("contexts: got %v", task.Contexts)
	}
	if len(task.Tags) != 1 || task.Tags[0] != "urgent" {
		t.Errorf("tags: got %v", task.Tags)
	}
	if !strings.Contains(task.Description, "task") {
		t.Errorf("description: got %q", task.Description)
	}
}

func TestParseLine_Metadata(t *testing.T) {
	task := ParseLine("fix bug id:abc-123 listid:inbox due:2024-02-01")
	if task == nil {
		t.Fatal("expected task")
	}
	if task.Metadata["id"] != "abc-123" {
		t.Errorf("metadata id: got %q", task.Metadata["id"])
	}
	if task.Metadata["listid"] != "inbox" {
		t.Errorf("metadata listid: got %q", task.Metadata["listid"])
	}
	if task.Metadata["due"] != "2024-02-01" {
		t.Errorf("metadata due: got %q", task.Metadata["due"])
	}
}

func TestFormatLine_Roundtrip(t *testing.T) {
	lines := []string{
		"(B) write tests",
		"x 2024-01-10 2024-01-01 done task",
		"task +proj @ctx #tag due:2024-12-31",
	}
	for _, line := range lines {
		task := ParseLine(line)
		if task == nil {
			t.Errorf("ParseLine(%q) returned nil", line)
			continue
		}
		formatted := FormatLine(task)
		task2 := ParseLine(formatted)
		if task2 == nil {
			t.Errorf("roundtrip ParseLine(FormatLine(...)) returned nil for %q", line)
			continue
		}
		if task.Completed != task2.Completed || task.Description != task2.Description {
			t.Errorf("roundtrip mismatch: in %q -> %q -> parsed %+v", line, formatted, task2)
		}
	}
}

func TestFormatLine_PreservesMetadata(t *testing.T) {
	task := &Task{
		Description: "task",
		Metadata:    map[string]string{"id": "x-y-z", "listid": "inbox"},
	}
	out := FormatLine(task)
	if !strings.Contains(out, "id:x-y-z") || !strings.Contains(out, "listid:inbox") {
		t.Errorf("FormatLine should include metadata: got %q", out)
	}
}

func TestParseContent(t *testing.T) {
	desc, projects, contexts, tags, meta := ParseContent("content +p @c #t key:val")
	if desc != "content" {
		t.Errorf("description: got %q", desc)
	}
	if len(projects) != 1 || projects[0] != "p" {
		t.Errorf("projects: %v", projects)
	}
	if len(contexts) != 1 || contexts[0] != "c" {
		t.Errorf("contexts: %v", contexts)
	}
	if len(tags) != 1 || tags[0] != "t" {
		t.Errorf("tags: %v", tags)
	}
	if meta["key"] != "val" {
		t.Errorf("metadata: %v", meta)
	}
}

func TestParseLine_CreationDate(t *testing.T) {
	task := ParseLine("2024-01-05 create report")
	if task == nil {
		t.Fatal("expected task")
	}
	if task.CreationDate == nil || task.CreationDate.Format("2006-01-02") != "2024-01-05" {
		t.Errorf("creation date: got %v", task.CreationDate)
	}
	if task.Description != "create report" {
		t.Errorf("description: got %q", task.Description)
	}
}

func TestFormatLine_CompletedWithDates(t *testing.T) {
	done := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	task := &Task{
		Completed:      true,
		CompletionDate: &done,
		CreationDate:   &created,
		Description:    "done",
		Metadata:       make(map[string]string),
	}
	out := FormatLine(task)
	if !strings.HasPrefix(out, "x 2024-01-10 2024-01-01") {
		t.Errorf("expected completed line with dates: got %q", out)
	}
}
