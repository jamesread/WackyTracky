package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wacky-tracky/wacky-tracky-server/internal/buildinfo"
	"github.com/wacky-tracky/wacky-tracky-server/internal/clientapi"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db"
	"github.com/wacky-tracky/wacky-tracky-server/internal/gitssh"
	"github.com/wacky-tracky/wacky-tracky-server/internal/httpserver"
	"github.com/wacky-tracky/wacky-tracky-server/internal/mcpserver"
	"github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

var root = &cobra.Command{
	Use: "wt",
	Run: mainRoot,
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run an MCP (Model Context Protocol) server over stdio",
	Long:  "Serves WackyTracky task operations as MCP tools over stdio, for use by LLM clients such as Claude Desktop or Cursor. Uses the same backend and config as the HTTP server.",
	Run:   mainMCP,
}

func mainRoot(cmd *cobra.Command, args []string) {
	httpserver.StartServer()
}

func mainMCP(cmd *cobra.Command, args []string) {
	// Stdio MCP calls the client API in-process (no HTTP auth wrap). Use
	// WACKYTRACKY_DEV_DISABLE_AUTH=true for local HTTP MCP without a Bearer key.
	svc := clientapi.GetNewClientAPI(db.GetNewDatabaseConnection(), nil)
	if err := mcpserver.New(svc).Serve(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func disableLogTimestamps() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors:    false,
		DisableTimestamp: true,
	})
}

func main() {
	disableLogTimestamps()

	log.WithFields(log.Fields{
		"version": buildinfo.Version,
		"commit":  buildinfo.Commit,
	}).Info("wacky-tracky")

	if err := runtimeconfig.Load(); err != nil {
		log.Fatalf("Config error: %v", err)
	}
	gitssh.SetupAndLog()

	root.AddCommand(mcpCmd)

	root.Execute()
}
