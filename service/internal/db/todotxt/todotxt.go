package todotxt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	db "github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
	. "github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

const (
	defaultListID    = "inbox"
	defaultListName  = "Inbox"
	listsFilename    = "todotxt_lists.txt"
	searchesFilename = "searches.txt"
)

// TodoTxt implements db.DB using todo.txt format files.
// See http://todotxt.org and https://github.com/todotxt/todo.txt
type TodoTxt struct {
	mu             sync.RWMutex
	todoPath       string
	donePath       string
	listsPath      string
	dir            string
	tasks          []*Task
	lists          []db.DBList
	defaultListID  string
	watcher        *fsnotify.Watcher
	reloadDebounce chan struct{}
}

func (d *TodoTxt) connectInitPaths() error {
	d.dir = RuntimeConfig.Database.Database
	if d.dir == "" {
		d.dir = "."
	}
	d.todoPath = filepath.Join(d.dir, "todo.txt")
	d.donePath = filepath.Join(d.dir, "done.txt")
	d.listsPath = filepath.Join(d.dir, listsFilename)
	d.defaultListID = defaultListID
	if err := os.MkdirAll(d.dir, 0755); err != nil {
		return fmt.Errorf("todotxt: create dir: %w", err)
	}
	return nil
}

func (d *TodoTxt) connectStartWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Warnf("todotxt: fsnotify new watcher: %v", err)
		return
	}
	if err := watcher.Add(d.dir); err != nil {
		log.Warnf("todotxt: fsnotify add %q: %v", d.dir, err)
		_ = watcher.Close()
		return
	}
	d.watcher = watcher
	d.reloadDebounce = make(chan struct{}, 1)
	go d.runReloadLoop()
	go d.runFileWatcher()
}

func (d *TodoTxt) Connect() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := d.connectInitPaths(); err != nil {
		return err
	}
	if err := d.loadLists(); err != nil {
		log.Warnf("todotxt: load lists: %v", err)
	}
	if err := d.loadTasks(); err != nil {
		return err
	}
	d.connectStartWatcher()
	log.WithFields(log.Fields{
		"dir":   d.dir,
		"tasks": len(d.tasks),
	}).Info("todotxt: connected")
	return nil
}

func fileTriggersReload(ev fsnotify.Event, todoBase, doneBase, listsBase string) bool {
	if ev.Op&(fsnotify.Write|fsnotify.Create) == 0 {
		return false
	}
	base := filepath.Base(ev.Name)
	return base == todoBase || base == doneBase || base == listsBase
}

// runFileWatcher watches the todotxt directory and reloads tasks/lists when todo.txt, done.txt, or lists file change.
func (d *TodoTxt) runFileWatcher() {
	todoBase := filepath.Base(d.todoPath)
	doneBase := filepath.Base(d.donePath)
	listsBase := filepath.Base(d.listsPath)
	for ev := range d.watcher.Events {
		if !fileTriggersReload(ev, todoBase, doneBase, listsBase) {
			continue
		}
		select {
		case d.reloadDebounce <- struct{}{}:
		default:
		}
	}
	close(d.reloadDebounce)
}

func (d *TodoTxt) doReload() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := d.loadTasks(); err != nil {
		log.Warnf("todotxt: reload tasks after file change: %v", err)
	}
	if err := d.loadLists(); err != nil {
		log.Warnf("todotxt: reload lists after file change: %v", err)
	}
	log.Debug("todotxt: reloaded after file change")
}

// runReloadLoop consumes reload requests and debounces actual reloads.
func (d *TodoTxt) runReloadLoop() {
	const debounceDelay = 300 * time.Millisecond
	var timer *time.Timer
	for range d.reloadDebounce {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceDelay, d.doReload)
	}
	if timer != nil {
		timer.Stop()
	}
}

func (d *TodoTxt) Print() {
	d.mu.RLock()
	defer d.mu.RUnlock()
	log.Infof("todotxt: %d tasks, %d lists", len(d.tasks), len(d.lists))
}

// Dir returns the todotxt data directory (for git status, etc.).
func (d *TodoTxt) Dir() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.dir
}

func shouldSkipTodoLine(line string) bool {
	line = strings.TrimSpace(line)
	return line == "" || strings.HasPrefix(line, "#")
}

func ensureTaskIDs(t *Task) {
	if t.Metadata["id"] == "" {
		t.Metadata["id"] = uuid.New().String()
	}
	if t.Metadata["listid"] == "" {
		t.Metadata["listid"] = defaultListID
	}
}

func parseTodoLineToTask(line string) (*Task, bool) {
	if shouldSkipTodoLine(line) {
		return nil, true
	}
	t := ParseLine(strings.TrimSpace(line))
	if t == nil {
		return nil, true
	}
	return t, false
}

func readTasksFromPath(path string) ([]*Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []*Task
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		t, skip := parseTodoLineToTask(sc.Text())
		if skip {
			continue
		}
		ensureTaskIDs(t)
		out = append(out, t)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("todotxt: read %s: %w", path, err)
	}
	return out, nil
}

func (d *TodoTxt) loadTasks() error {
	d.tasks = nil
	for _, path := range []string{d.todoPath, d.donePath} {
		tasks, err := readTasksFromPath(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("todotxt: open %s: %w", path, err)
		}
		d.tasks = append(d.tasks, tasks...)
	}
	return nil
}

func parseListLineParts(line string) (id, title string) {
	parts := strings.SplitN(line, "\t", 2)
	id = strings.TrimSpace(parts[0])
	title = defaultListName
	if len(parts) > 1 {
		title = strings.TrimSpace(parts[1])
	}
	return id, title
}

func parseListLine(line string) (id, title string, skip bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", true
	}
	id, title = parseListLineParts(line)
	if id == "" || id == defaultListID {
		return "", "", true
	}
	return id, title, false
}

func (d *TodoTxt) loadLists() error {
	d.lists = []db.DBList{
		{ID: defaultListID, Title: defaultListName, CountTasks: 0},
	}
	f, err := os.Open(d.listsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		id, title, skip := parseListLine(sc.Text())
		if skip {
			continue
		}
		d.lists = append(d.lists, db.DBList{ID: id, Title: title, CountTasks: 0})
	}
	return sc.Err()
}

func (d *TodoTxt) saveTasks() error {
	var incomplete, complete []string
	for _, t := range d.tasks {
		line := FormatLine(t)
		if t.Completed {
			complete = append(complete, line)
		} else {
			incomplete = append(incomplete, line)
		}
	}
	if err := os.WriteFile(d.todoPath, []byte(strings.Join(incomplete, "\n")+"\n"), 0644); err != nil {
		return err
	}
	return os.WriteFile(d.donePath, []byte(strings.Join(complete, "\n")+"\n"), 0644)
}

func (d *TodoTxt) saveLists() error {
	var lines []string
	for _, l := range d.lists {
		if l.ID == defaultListID {
			continue
		}
		lines = append(lines, l.ID+"\t"+l.Title)
	}
	return os.WriteFile(d.listsPath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

func taskParentTypeAndID(t *Task) (parentType, parentID string) {
	parentID = t.Metadata["parent"]
	if parentID != "" {
		return "task", parentID
	}
	listID := t.Metadata["listid"]
	if listID == "" {
		listID = defaultListID
	}
	return "list", listID
}

func normalizePriority(p string) string {
	p = strings.ToUpper(strings.TrimSpace(p))
	if len(p) > 1 {
		return p[:1]
	}
	return p
}

func taskToDB(t *Task, subcount int32, waitUntil, dueDate string) db.DBTask {
	id := t.Metadata["id"]
	if id == "" {
		id = uuid.New().String()
	}
	listID := t.Metadata["listid"]
	if listID == "" {
		listID = defaultListID
	}
	parentType, parentID := taskParentTypeAndID(t)
	content := t.Description
	for _, p := range t.Projects {
		content += " +" + p
	}
	content = strings.TrimSpace(content)
	tags := make([]string, len(t.Tags))
	copy(tags, t.Tags)
	contexts := make([]string, len(t.Contexts))
	copy(contexts, t.Contexts)
	priority := normalizePriority(t.Metadata["priority"])
	return db.DBTask{
		ID:            id,
		Content:       content,
		ParentId:      parentID,
		ParentType:    parentType,
		CountSubitems: subcount,
		Tags:          tags,
		Contexts:      contexts,
		WaitUntil:     strings.TrimSpace(waitUntil),
		Priority:      priority,
		DueDate:       strings.TrimSpace(dueDate),
	}
}

func countSubtasks(tasks []*Task, taskID string) int32 {
	var n int32
	for _, t := range tasks {
		if t.Completed {
			continue
		}
		if t.Metadata["parent"] == taskID {
			n++
		}
	}
	return n
}

func addContextsToTagList(tags *[]db.DBTag, seen map[string]bool, contexts []string) {
	for _, c := range contexts {
		if seen[c] {
			continue
		}
		seen[c] = true
		*tags = append(*tags, db.DBTag{ID: c, Title: "@" + c})
	}
}

func addTagsToTagList(tags *[]db.DBTag, seen map[string]bool, tagList []string) {
	for _, tag := range tagList {
		key := "#" + tag
		if seen[key] {
			continue
		}
		seen[key] = true
		*tags = append(*tags, db.DBTag{ID: tag, Title: key})
	}
}

func (d *TodoTxt) GetTags() ([]db.DBTag, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	seen := make(map[string]bool)
	var tags []db.DBTag
	for _, t := range d.tasks {
		addContextsToTagList(&tags, seen, t.Contexts)
		addTagsToTagList(&tags, seen, t.Tags)
	}
	return tags, nil
}

func (d *TodoTxt) GetTask(id string) (*db.DBTask, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, t := range d.tasks {
		if t.Metadata["id"] == id {
			ret := taskToDB(t, countSubtasks(d.tasks, id), d.getTaskMetadataField(id, "wait"), d.getTaskMetadataField(id, "due"))
			return &ret, nil
		}
	}
	return nil, nil
}

func isRootTaskInList(t *Task, listId string) bool {
	if t.Completed || t.Metadata["parent"] != "" {
		return false
	}
	lid := t.Metadata["listid"]
	if lid == "" {
		lid = defaultListID
	}
	return lid == listId
}

func (d *TodoTxt) GetTasks(listId string) ([]db.DBTask, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var out []db.DBTask
	for _, t := range d.tasks {
		if !isRootTaskInList(t, listId) {
			continue
		}
		out = append(out, taskToDB(t, countSubtasks(d.tasks, t.Metadata["id"]), d.getTaskMetadataField(t.Metadata["id"], "wait"), d.getTaskMetadataField(t.Metadata["id"], "due")))
	}
	return out, nil
}

func (d *TodoTxt) GetSubtasks(itemId string) ([]db.DBTask, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var out []db.DBTask
	for _, t := range d.tasks {
		if t.Completed {
			continue
		}
		if t.Metadata["parent"] != itemId {
			continue
		}
		out = append(out, taskToDB(t, countSubtasks(d.tasks, t.Metadata["id"]), d.getTaskMetadataField(t.Metadata["id"], "wait"), d.getTaskMetadataField(t.Metadata["id"], "due")))
	}
	return out, nil
}

func countRootTasksByList(tasks []*Task) map[string]int32 {
	counts := make(map[string]int32)
	for _, t := range tasks {
		if t.Completed || t.Metadata["parent"] != "" {
			continue
		}
		lid := t.Metadata["listid"]
		if lid == "" {
			lid = defaultListID
		}
		counts[lid]++
	}
	return counts
}

func (d *TodoTxt) GetLists() ([]db.DBList, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	counts := countRootTasksByList(d.tasks)
	out := make([]db.DBList, len(d.lists))
	for i := range d.lists {
		out[i] = d.lists[i]
		out[i].CountTasks = counts[d.lists[i].ID]
	}
	return out, nil
}

func (d *TodoTxt) ensureParentHasProject(parentTaskId string) {
	for i := range d.tasks {
		if d.tasks[i].Metadata["id"] != parentTaskId {
			continue
		}
		if d.tasks[i].Metadata["project"] != "true" {
			d.tasks[i].Metadata["project"] = "true"
		}
		break
	}
}

func mergeContentMetaInto(meta map[string]string, contentMeta map[string]string) {
	for k, v := range contentMeta {
		if k == "id" || k == "listid" || k == "parent" {
			continue
		}
		meta[k] = v
	}
}

func (d *TodoTxt) CreateTask(content string, listId string, parentTaskId string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	id := uuid.New().String()
	if listId == "" {
		listId = defaultListID
	}
	meta := map[string]string{"id": id, "listid": listId}
	if parentTaskId != "" {
		meta["parent"] = parentTaskId
		d.ensureParentHasProject(parentTaskId)
	}
	desc, projects, contexts, tags, contentMeta := ParseContent(content)
	mergeContentMetaInto(meta, contentMeta)
	t := &Task{
		Description: desc,
		Projects:    projects,
		Contexts:    contexts,
		Tags:        tags,
		Metadata:    meta,
	}
	d.tasks = append(d.tasks, t)
	if err := d.saveTasks(); err != nil {
		return "", err
	}
	return id, nil
}

func (d *TodoTxt) CreateList(title string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	id := uuid.New().String()
	d.lists = append(d.lists, db.DBList{ID: id, Title: title, CountTasks: 0})
	return d.saveLists()
}

func (d *TodoTxt) UpdateList(id string, title string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.lists {
		if d.lists[i].ID == id {
			d.lists[i].Title = title
			return d.saveLists()
		}
	}
	return nil
}

func (d *TodoTxt) removeListByID(id string) {
	for i, l := range d.lists {
		if l.ID == id {
			d.lists = append(d.lists[:i], d.lists[i+1:]...)
			return
		}
	}
}

func (d *TodoTxt) reassignTasksToDefaultList(listId string) {
	for _, t := range d.tasks {
		if t.Metadata["listid"] == listId {
			t.Metadata["listid"] = defaultListID
		}
	}
}

func (d *TodoTxt) DeleteList(id string) error {
	if id == defaultListID {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.removeListByID(id)
	d.reassignTasksToDefaultList(id)
	if err := d.saveLists(); err != nil {
		return err
	}
	return d.saveTasks()
}

func searchableMatchesInclude(searchable string, include []string) bool {
	if len(include) == 0 {
		return true
	}
	for _, term := range include {
		if strings.Contains(searchable, term) {
			return true
		}
	}
	return false
}

func searchableMatchesExclude(searchable string, exclude []string) bool {
	for _, term := range exclude {
		if strings.Contains(searchable, term) {
			return true
		}
	}
	return false
}

func taskMatchesQuery(searchable string, include, exclude []string) bool {
	if !searchableMatchesInclude(searchable, include) {
		return false
	}
	return !searchableMatchesExclude(searchable, exclude)
}

// SearchTasks implements db.Searchable: freetext search over task content in memory.
// Only incomplete tasks (todo.txt) are searched; completed tasks (done.txt) are excluded.
// Query supports negation: "-word" excludes tasks containing "word". Other terms are required (any match).
// Example: "work -home" matches tasks containing "work" and not containing "home".
func (d *TodoTxt) SearchTasks(query string) ([]db.DBTask, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	include, exclude := parseSearchQuery(query)
	var out []db.DBTask
	for _, t := range d.tasks {
		if t.Completed {
			continue
		}
		searchable := strings.ToLower(buildSearchableTextWithAncestors(d.tasks, t))
		if !taskMatchesQuery(searchable, include, exclude) {
			continue
		}
		out = append(out, taskToDB(t, countSubtasks(d.tasks, t.Metadata["id"]), d.getTaskMetadataField(t.Metadata["id"], "wait"), d.getTaskMetadataField(t.Metadata["id"], "due")))
	}
	return out, nil
}

func parseOneSearchTerm(w string) (term string, exclude bool) {
	w = strings.TrimSpace(w)
	if w == "" {
		return "", false
	}
	lower := strings.ToLower(w)
	if strings.HasPrefix(lower, "-") && len(lower) > 1 {
		return lower[1:], true
	}
	return lower, false
}

// parseSearchQuery splits query into include terms (plain words) and exclude terms (prefix "-").
func parseSearchQuery(query string) (include, exclude []string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	for _, w := range strings.Fields(query) {
		term, isExclude := parseOneSearchTerm(w)
		if term == "" {
			continue
		}
		if isExclude {
			exclude = append(exclude, term)
		} else {
			include = append(include, term)
		}
	}
	return include, exclude
}

func buildSearchableText(t *Task) string {
	var parts []string
	parts = append(parts, t.Description)
	for _, p := range t.Projects {
		parts = append(parts, "+"+p)
	}
	for _, c := range t.Contexts {
		parts = append(parts, "@"+c)
	}
	for _, tag := range t.Tags {
		parts = append(parts, "#"+tag)
	}
	return strings.Join(parts, " ")
}

func addUniquePrefixed(parts *[]string, seen map[string]struct{}, prefix string, items []string) {
	for _, p := range items {
		key := prefix + p
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		*parts = append(*parts, key)
	}
}

func collectUniqueIntoParts(parts *[]string, seen map[string]struct{}, projects, contexts, tags []string) {
	addUniquePrefixed(parts, seen, "+", projects)
	addUniquePrefixed(parts, seen, "@", contexts)
	addUniquePrefixed(parts, seen, "#", tags)
}

func findTaskByID(tasks []*Task, id string) *Task {
	for _, c := range tasks {
		if c.Metadata["id"] == id {
			return c
		}
	}
	return nil
}

// buildSearchableTextWithAncestors returns searchable text for t including projects, contexts, and tags from all ancestor tasks (so subtasks match when a parent has the tag/context).
func buildSearchableTextWithAncestors(tasks []*Task, t *Task) string {
	seen := make(map[string]struct{})
	var parts []string
	parts = append(parts, t.Description)
	collectUniqueIntoParts(&parts, seen, t.Projects, t.Contexts, t.Tags)
	parentID := t.Metadata["parent"]
	for parentID != "" {
		parent := findTaskByID(tasks, parentID)
		if parent == nil {
			break
		}
		collectUniqueIntoParts(&parts, seen, parent.Projects, parent.Contexts, parent.Tags)
		parentID = parent.Metadata["parent"]
	}
	return strings.Join(parts, " ")
}

func applyContentToTask(t *Task, content string) {
	desc, projects, contexts, tags, meta := ParseContent(content)
	t.Description = desc
	t.Projects = projects
	t.Contexts = contexts
	t.Tags = tags
	for k, v := range meta {
		if k == "id" || k == "listid" {
			continue
		}
		t.Metadata[k] = v
	}
}

// UpdateTask implements db.Updatable: update task content in memory and persist.
func (d *TodoTxt) UpdateTask(id string, content string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, t := range d.tasks {
		if t.Metadata["id"] != id {
			continue
		}
		applyContentToTask(t, content)
		return d.saveTasks()
	}
	return nil
}

// DoneTask marks a task as completed (moves it to done.txt) and persists.
func (d *TodoTxt) DoneTask(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, t := range d.tasks {
		if t.Metadata["id"] == id {
			t.Completed = true
			now := time.Now().UTC()
			t.CompletionDate = &now
			return d.saveTasks()
		}
	}
	return nil
}

// GetSavedSearches implements db.SavedSearchesStore: read saved searches from searches.txt.
func (d *TodoTxt) GetSavedSearches() ([]db.SavedSearch, error) {
	d.mu.RLock()
	path := filepath.Join(d.dir, searchesFilename)
	d.mu.RUnlock()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var list []db.SavedSearch
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// SetSavedSearches implements db.SavedSearchesStore: write saved searches to searches.txt.
func (d *TodoTxt) SetSavedSearches(searches []db.SavedSearch) error {
	d.mu.RLock()
	path := filepath.Join(d.dir, searchesFilename)
	d.mu.RUnlock()
	b, err := json.MarshalIndent(searches, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

const metadataDirName = "metadata"

func isLetterRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigitRune(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAlnumRune(r rune) bool {
	return isLetterRune(r) || isDigitRune(r)
}

func isAlnumOrDashUnderscore(r rune) bool {
	return isAlnumRune(r) || r == '-' || r == '_'
}

func sanitizeTaskIDForPath(taskId string) string {
	return strings.Map(func(r rune) rune {
		if isAlnumOrDashUnderscore(r) {
			return r
		}
		return '_'
	}, taskId)
}

// metadataPath returns the path for a task's metadata field file: metadata/<taskid>/<field>.txt
func (d *TodoTxt) metadataPath(taskId, field string) string {
	return filepath.Join(d.dir, metadataDirName, sanitizeTaskIDForPath(taskId), field+".txt")
}

// getTaskMetadataField reads a single metadata field file; returns empty string if missing or on error.
func (d *TodoTxt) getTaskMetadataField(taskId, field string) string {
	p := d.metadataPath(taskId, field)
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func (d *TodoTxt) readPriorityFromTasks(taskId string) string {
	for _, t := range d.tasks {
		if t.Metadata["id"] != taskId {
			continue
		}
		p := strings.TrimSpace(t.Metadata["priority"])
		if p == "" {
			return ""
		}
		if len(p) > 1 {
			p = p[:1]
		}
		return strings.ToUpper(p)
	}
	return ""
}

func (d *TodoTxt) readMetadataFieldFiles(taskId string, fields []string) (map[string]string, error) {
	out := make(map[string]string)
	for _, field := range fields {
		p := d.metadataPath(taskId, field)
		b, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		out[field] = string(b)
	}
	return out, nil
}

// GetTaskMetadata implements db.TaskMetadataStore: read metadata from task line (priority) and metadata/<taskid>/<field>.txt (notes, wait)
func (d *TodoTxt) GetTaskMetadata(taskId string) (map[string]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make(map[string]string)
	if p := d.readPriorityFromTasks(taskId); p != "" {
		out["priority"] = p
	}
	fileMeta, err := d.readMetadataFieldFiles(taskId, []string{"notes", "wait", "due"})
	if err != nil {
		return nil, err
	}
	for k, v := range fileMeta {
		out[k] = v
	}
	return out, nil
}

func (d *TodoTxt) setPriorityInTaskLine(taskId, value string) error {
	value = strings.TrimSpace(value)
	if len(value) > 1 {
		value = value[:1]
	}
	value = strings.ToUpper(value)
	for _, t := range d.tasks {
		if t.Metadata["id"] != taskId {
			continue
		}
		if value == "" {
			delete(t.Metadata, "priority")
		} else {
			t.Metadata["priority"] = value
		}
		return d.saveTasks()
	}
	return nil
}

// SetTaskMetadata implements db.TaskMetadataStore: for "priority" write into task line as priority:X; else metadata/<taskid>/<field>.txt
func (d *TodoTxt) SetTaskMetadata(taskId, field, value string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if field == "priority" {
		return d.setPriorityInTaskLine(taskId, value)
	}
	p := d.metadataPath(taskId, field)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(value), 0644)
}
