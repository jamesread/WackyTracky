// Package mcpserver exposes WackyTracky's task operations as Model Context
// Protocol (MCP) tools so that LLM clients can read and manage tasks. It serves
// over stdio, which is how desktop LLM clients (e.g. Claude Desktop, Cursor)
// launch a local MCP server.
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"time"

	"connectrpc.com/connect"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
	"github.com/wacky-tracky/wacky-tracky-server/internal/buildinfo"
)

// ClientService is the subset of the Connect client API that the MCP tools use.
// The real client API service satisfies it, so tools reuse the same business
// logic (task tree building, hide-at-times filtering, backend selection, etc.).
type ClientService interface {
	GetLists(context.Context, *connect.Request[pb.GetListsRequest]) (*connect.Response[pb.GetListsResponse], error)
	ListTasks(context.Context, *connect.Request[pb.ListTasksRequest]) (*connect.Response[pb.ListTasksResponse], error)
	SearchTasks(context.Context, *connect.Request[pb.SearchTasksRequest]) (*connect.Response[pb.SearchTasksResponse], error)
	CreateTask(context.Context, *connect.Request[pb.CreateTaskRequest]) (*connect.Response[pb.CreateTaskResponse], error)
	UpdateTask(context.Context, *connect.Request[pb.UpdateTaskRequest]) (*connect.Response[pb.UpdateTaskResponse], error)
	DoneTask(context.Context, *connect.Request[pb.DoneTaskRequest]) (*connect.Response[pb.DoneTaskResponse], error)
	MoveTask(context.Context, *connect.Request[pb.MoveTaskRequest]) (*connect.Response[pb.MoveTaskResponse], error)
	CreateList(context.Context, *connect.Request[pb.CreateListRequest]) (*connect.Response[pb.CreateListResponse], error)
	SetTaskMetadata(context.Context, *connect.Request[pb.SetTaskMetadataRequest]) (*connect.Response[pb.SetTaskMetadataResponse], error)
	DeleteTask(context.Context, *connect.Request[pb.DeleteTaskRequest]) (*connect.Response[pb.DeleteTaskResponse], error)
	DeleteList(context.Context, *connect.Request[pb.DeleteListRequest]) (*connect.Response[pb.DeleteListResponse], error)
	ListCompletedTasks(context.Context, *connect.Request[pb.ListCompletedTasksRequest]) (*connect.Response[pb.ListCompletedTasksResponse], error)
}

// taskMetadataFields are the per-task metadata fields the web UI reads and writes.
var taskMetadataFields = []string{"notes", "wait", "due", "priority"}

var errMoveTarget = errors.New("set exactly one of list_id or parent_task_id")

var errInvalidSnoozeUntil = errors.New("until must be an ISO date (YYYY-MM-DD) or RFC 3339 date-time (e.g. 2026-11-25T09:00:00Z)")

// Server wires the WackyTracky client API to an MCP server.
type Server struct {
	svc ClientService
}

// New returns an MCP Server backed by the given client API service.
func New(svc ClientService) *Server {
	return &Server{svc: svc}
}

// MCP builds an MCP server with all WackyTracky tools registered.
func (s *Server) MCP() *server.MCPServer {
	srv := server.NewMCPServer("wacky-tracky", buildinfo.Version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	s.registerTools(srv)
	return srv
}

// HTTPHandler returns a Streamable HTTP handler for /mcp on the main server.
func (s *Server) HTTPHandler() http.Handler {
	return server.NewStreamableHTTPServer(s.MCP(), server.WithEndpointPath("/mcp"))
}

// Serve runs the MCP server over stdio until the client disconnects.
func (s *Server) Serve() error {
	return server.ServeStdio(s.MCP())
}

func (s *Server) registerTools(srv *server.MCPServer) {
	srv.AddTool(mcp.NewTool("list_lists",
		mcp.WithDescription("List all task lists with their IDs, titles, and item counts."),
	), s.handleListLists)

	srv.AddTool(mcp.NewTool("list_tasks",
		mcp.WithDescription("List the open tasks in a list (including subtasks), given a list ID from list_lists. Each task includes notes, waitUntil (snoozed until), dueDate and priority when set. Completed tasks are not included; use list_completed_tasks."),
		mcp.WithString("list_id", mcp.Required(), mcp.Description("The list ID to read tasks from.")),
	), s.handleListTasks)

	srv.AddTool(mcp.NewTool("search_tasks",
		mcp.WithDescription("Search tasks across all lists. Supports plain terms, #tag, @context, and -term to exclude."),
		mcp.WithString("query", mcp.Required(), mcp.Description("The search query.")),
	), s.handleSearchTasks)

	srv.AddTool(mcp.NewTool("create_task",
		mcp.WithDescription("Create a task. Content is a todo.txt-style line (text, +projects, @contexts, #tags). Notes, snooze, due date and priority can be set in the same call."),
		mcp.WithString("content", mcp.Required(), mcp.Description("The task content.")),
		mcp.WithString("list_id", mcp.Description("Optional list ID to add the task to.")),
		mcp.WithString("parent_task_id", mcp.Description("Optional task ID to create this as a subtask of.")),
		mcp.WithString("notes", mcp.Description("Optional free-text notes.")),
		mcp.WithString("wait_until", mcp.Description("Optional snooze: hide the task until this ISO date (YYYY-MM-DD, midnight UTC) or RFC 3339 date-time.")),
		mcp.WithString("due_date", mcp.Description("Optional due date, YYYY-MM-DD.")),
		mcp.WithString("priority", mcp.Description("Optional priority, a single letter A-Z (A highest).")),
	), s.handleCreateTask)

	srv.AddTool(mcp.NewTool("update_task",
		mcp.WithDescription("Replace a task's content with a new todo.txt-style line."),
		mcp.WithString("id", mcp.Required(), mcp.Description("The task ID.")),
		mcp.WithString("content", mcp.Required(), mcp.Description("The new task content.")),
	), s.handleUpdateTask)

	srv.AddTool(mcp.NewTool("complete_task",
		mcp.WithDescription("Mark a task as done."),
		mcp.WithString("id", mcp.Required(), mcp.Description("The task ID to complete.")),
	), s.handleCompleteTask)

	srv.AddTool(mcp.NewTool("move_task",
		mcp.WithDescription("Move a task, with its subtasks, to another list or under another task. Set exactly one of list_id or parent_task_id. The task keeps its ID, notes and other metadata. A subtask moved to a list becomes a top-level task there; a task moved under another task joins that task's list."),
		mcp.WithString("id", mcp.Required(), mcp.Description("The task ID to move.")),
		mcp.WithString("list_id", mcp.Description("The ID of the list to move the task to, from list_lists.")),
		mcp.WithString("parent_task_id", mcp.Description("The ID of the task to make this a subtask of. Cannot be the task itself or one of its subtasks.")),
	), s.handleMoveTask)

	srv.AddTool(mcp.NewTool("delete_task",
		mcp.WithDescription("Permanently delete a task (open or completed) with its subtasks and notes. It is not kept in the completed history and cannot be undone. Use complete_task for work that was actually done."),
		mcp.WithString("id", mcp.Required(), mcp.Description("The task ID to delete.")),
		mcp.WithDestructiveHintAnnotation(true),
	), s.handleDeleteTask)

	srv.AddTool(mcp.NewTool("delete_list",
		mcp.WithDescription("Delete a list. Its tasks are not deleted: they move to the Inbox. The Inbox itself cannot be deleted."),
		mcp.WithString("id", mcp.Required(), mcp.Description("The list ID to delete, from list_lists.")),
		mcp.WithDestructiveHintAnnotation(true),
	), s.handleDeleteList)

	srv.AddTool(mcp.NewTool("list_completed_tasks",
		mcp.WithDescription("List completed tasks, most recently completed first, each with completedAt (the day it was marked done)."),
		mcp.WithString("since", mcp.Description("Optional; only tasks completed on or after this ISO date (YYYY-MM-DD) or RFC 3339 date-time. Completion is recorded per day.")),
		mcp.WithString("list_id", mcp.Description("Optional; only tasks in this list.")),
	), s.handleListCompletedTasks)

	srv.AddTool(mcp.NewTool("snooze_task",
		mcp.WithDescription("Snooze a task: hide it from the task list until a given time, after which it reappears. Omit until (or pass an empty string) to un-snooze. Returns the stored time as waitUntil."),
		mcp.WithString("id", mcp.Required(), mcp.Description("The task ID to snooze.")),
		mcp.WithString("until", mcp.Description("When the task should reappear: an ISO date (YYYY-MM-DD, treated as midnight UTC) or an RFC 3339 date-time (e.g. 2026-11-25T09:00:00Z). Empty or omitted clears the snooze.")),
	), s.handleSnoozeTask)

	srv.AddTool(mcp.NewTool("update_task_metadata",
		mcp.WithDescription("Set one metadata field on a task without touching its content: notes (free text), wait (snooze until, ISO date-time; prefer snooze_task), due (ISO date YYYY-MM-DD), or priority (a single letter A-Z, A highest). An empty value clears the field."),
		mcp.WithString("id", mcp.Required(), mcp.Description("The task ID.")),
		mcp.WithString("field", mcp.Required(), mcp.Enum(taskMetadataFields...), mcp.Description("The metadata field to set.")),
		mcp.WithString("value", mcp.Description("The new value. Empty or omitted clears the field.")),
	), s.handleUpdateTaskMetadata)

	srv.AddTool(mcp.NewTool("create_list",
		mcp.WithDescription("Create a new task list."),
		mcp.WithString("title", mcp.Required(), mcp.Description("The list title.")),
	), s.handleCreateList)
}

// toolJSON renders a protobuf response as the API's JSON (camelCase fields),
// which LLM clients parse reliably.
func toolJSON(m proto.Message) (*mcp.CallToolResult, error) {
	b, err := protojson.Marshal(m)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to encode result", err), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func (s *Server) handleListLists(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := s.svc.GetLists(ctx, connect.NewRequest(&pb.GetListsRequest{}))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("list_lists failed", err), nil
	}
	return toolJSON(resp.Msg)
}

func (s *Server) handleListTasks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	listID, err := req.RequireString("list_id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	resp, err := s.svc.ListTasks(ctx, connect.NewRequest(&pb.ListTasksRequest{ParentId: listID, ParentType: "list"}))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("list_tasks failed", err), nil
	}
	return toolJSON(resp.Msg)
}

func (s *Server) handleSearchTasks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	resp, err := s.svc.SearchTasks(ctx, connect.NewRequest(&pb.SearchTasksRequest{Query: query}))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("search_tasks failed", err), nil
	}
	return toolJSON(resp.Msg)
}

func (s *Server) handleCreateTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	meta, err := createTaskMetadata(req)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	resp, err := s.svc.CreateTask(ctx, connect.NewRequest(&pb.CreateTaskRequest{
		Content:      content,
		ParentListId: req.GetString("list_id", ""),
		ParentTaskId: req.GetString("parent_task_id", ""),
	}))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("create_task failed", err), nil
	}
	if err := s.applyTaskMetadata(ctx, resp.Msg.Task, meta); err != nil {
		return mcp.NewToolResultErrorFromErr("create_task created task "+resp.Msg.Task.GetId()+" but could not set its metadata", err), nil
	}
	return toolJSON(resp.Msg)
}

// taskMetadataValue is one metadata field to set on a task, in the order given.
type taskMetadataValue struct {
	field string
	value string
}

// createTaskMetadata validates the optional create_task metadata up front, so
// bad input is rejected before a task is created.
func createTaskMetadata(req mcp.CallToolRequest) ([]taskMetadataValue, error) {
	wait, err := normalizeSnoozeUntil(req.GetString("wait_until", ""))
	if err != nil {
		return nil, err
	}
	candidates := []taskMetadataValue{
		{"notes", req.GetString("notes", "")},
		{"wait", wait},
		{"due", req.GetString("due_date", "")},
		{"priority", req.GetString("priority", "")},
	}
	return slices.DeleteFunc(candidates, func(m taskMetadataValue) bool { return m.value == "" }), nil
}

func (s *Server) applyTaskMetadata(ctx context.Context, task *pb.Task, meta []taskMetadataValue) error {
	for _, m := range meta {
		if _, err := s.svc.SetTaskMetadata(ctx, connect.NewRequest(&pb.SetTaskMetadataRequest{TaskId: task.GetId(), Field: m.field, Value: m.value})); err != nil {
			return err
		}
		setTaskField(task, m)
	}
	return nil
}

func setTaskField(task *pb.Task, m taskMetadataValue) {
	switch m.field {
	case "notes":
		task.Notes = m.value
	case "wait":
		task.WaitUntil = m.value
	case "due":
		task.DueDate = m.value
	case "priority":
		task.Priority = m.value
	}
}

func (s *Server) handleUpdateTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	resp, err := s.svc.UpdateTask(ctx, connect.NewRequest(&pb.UpdateTaskRequest{Id: id, Content: content}))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("update_task failed", err), nil
	}
	return toolJSON(resp.Msg)
}

func (s *Server) handleCompleteTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	if _, err := s.svc.DoneTask(ctx, connect.NewRequest(&pb.DoneTaskRequest{Id: id})); err != nil {
		return mcp.NewToolResultErrorFromErr("complete_task failed", err), nil
	}
	return mcp.NewToolResultText("Task " + id + " marked done."), nil
}

func (s *Server) handleMoveTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moveReq, err := moveTaskRequest(req)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	resp, err := s.svc.MoveTask(ctx, connect.NewRequest(moveReq))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("move_task failed", err), nil
	}
	return toolJSON(resp.Msg)
}

func moveTaskRequest(req mcp.CallToolRequest) (*pb.MoveTaskRequest, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return nil, err
	}
	listID, parentID := req.GetString("list_id", ""), req.GetString("parent_task_id", "")
	if (listID == "") == (parentID == "") {
		return nil, errMoveTarget
	}
	return &pb.MoveTaskRequest{TaskId: id, TargetListId: listID, TargetParentTaskId: parentID}, nil
}

func (s *Server) handleDeleteTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	if _, err := s.svc.DeleteTask(ctx, connect.NewRequest(&pb.DeleteTaskRequest{Id: id})); err != nil {
		return mcp.NewToolResultErrorFromErr("delete_task failed", err), nil
	}
	return mcp.NewToolResultText("Task " + id + " and its subtasks deleted."), nil
}

func (s *Server) handleDeleteList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	if _, err := s.svc.DeleteList(ctx, connect.NewRequest(&pb.DeleteListRequest{Id: id})); err != nil {
		return mcp.NewToolResultErrorFromErr("delete_list failed", err), nil
	}
	return mcp.NewToolResultText("List " + id + " deleted; its tasks moved to the Inbox."), nil
}

func (s *Server) handleListCompletedTasks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := s.svc.ListCompletedTasks(ctx, connect.NewRequest(&pb.ListCompletedTasksRequest{
		Since:  req.GetString("since", ""),
		ListId: req.GetString("list_id", ""),
	}))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("list_completed_tasks failed", err), nil
	}
	return toolJSON(resp.Msg)
}

func (s *Server) handleCreateList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, err := req.RequireString("title")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	if _, err := s.svc.CreateList(ctx, connect.NewRequest(&pb.CreateListRequest{Title: title})); err != nil {
		return mcp.NewToolResultErrorFromErr("create_list failed", err), nil
	}
	return mcp.NewToolResultText("List " + title + " created."), nil
}

// normalizeSnoozeUntil stores times as RFC 3339 UTC, matching what the web UI writes.
func normalizeSnoozeUntil(until string) (string, error) {
	if until == "" {
		return "", nil
	}
	if t, err := time.Parse(time.RFC3339, until); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	if t, err := time.Parse(time.DateOnly, until); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	return "", errInvalidSnoozeUntil
}

func snoozeResult(id, until string) *mcp.CallToolResult {
	if until == "" {
		return mcp.NewToolResultText("Task " + id + " is no longer snoozed.")
	}
	b, err := json.Marshal(map[string]string{"id": id, "waitUntil": until})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to encode result", err)
	}
	return mcp.NewToolResultText(string(b))
}

func (s *Server) handleSnoozeTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	until, err := normalizeSnoozeUntil(req.GetString("until", ""))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	if _, err := s.svc.SetTaskMetadata(ctx, connect.NewRequest(&pb.SetTaskMetadataRequest{TaskId: id, Field: "wait", Value: until})); err != nil {
		return mcp.NewToolResultErrorFromErr("snooze_task failed", err), nil
	}
	return snoozeResult(id, until), nil
}

func requireTaskMetadataField(req mcp.CallToolRequest) (string, error) {
	field, err := req.RequireString("field")
	if err != nil {
		return "", err
	}
	if !slices.Contains(taskMetadataFields, field) {
		return "", errors.New("field must be one of: notes, wait, due, priority")
	}
	return field, nil
}

func (s *Server) handleUpdateTaskMetadata(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	field, err := requireTaskMetadataField(req)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("invalid arguments", err), nil
	}
	value := req.GetString("value", "")
	if _, err := s.svc.SetTaskMetadata(ctx, connect.NewRequest(&pb.SetTaskMetadataRequest{TaskId: id, Field: field, Value: value})); err != nil {
		return mcp.NewToolResultErrorFromErr("update_task_metadata failed", err), nil
	}
	return mcp.NewToolResultText("Task " + id + " " + field + " updated."), nil
}
