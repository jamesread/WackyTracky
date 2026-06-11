// Package mcpserver exposes WackyTracky's task operations as Model Context
// Protocol (MCP) tools so that LLM clients can read and manage tasks. It serves
// over stdio, which is how desktop LLM clients (e.g. Claude Desktop, Cursor)
// launch a local MCP server.
package mcpserver

import (
	"context"

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
	CreateList(context.Context, *connect.Request[pb.CreateListRequest]) (*connect.Response[pb.CreateListResponse], error)
}

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
	srv := server.NewMCPServer("wacky-tracky", buildinfo.Version, server.WithToolCapabilities(false))
	s.registerTools(srv)
	return srv
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
		mcp.WithDescription("List the tasks in a list (including subtasks), given a list ID from list_lists."),
		mcp.WithString("list_id", mcp.Required(), mcp.Description("The list ID to read tasks from.")),
	), s.handleListTasks)

	srv.AddTool(mcp.NewTool("search_tasks",
		mcp.WithDescription("Search tasks across all lists. Supports plain terms, #tag, @context, and -term to exclude."),
		mcp.WithString("query", mcp.Required(), mcp.Description("The search query.")),
	), s.handleSearchTasks)

	srv.AddTool(mcp.NewTool("create_task",
		mcp.WithDescription("Create a task. Content is a todo.txt-style line (text, +projects, @contexts, #tags)."),
		mcp.WithString("content", mcp.Required(), mcp.Description("The task content.")),
		mcp.WithString("list_id", mcp.Description("Optional list ID to add the task to.")),
		mcp.WithString("parent_task_id", mcp.Description("Optional task ID to create this as a subtask of.")),
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
	resp, err := s.svc.CreateTask(ctx, connect.NewRequest(&pb.CreateTaskRequest{
		Content:      content,
		ParentListId: req.GetString("list_id", ""),
		ParentTaskId: req.GetString("parent_task_id", ""),
	}))
	if err != nil {
		return mcp.NewToolResultErrorFromErr("create_task failed", err), nil
	}
	return toolJSON(resp.Msg)
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
