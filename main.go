// Package main is the entry point for the DuckOps CLI.
//
//	@title			DuckOps API
//	@version		1.0
//	@description	DuckOps is a terminal-based AI coding assistant. This API is served over a Unix socket (or Windows named pipe) and provides programmatic access to workspaces, sessions, agents, LSP, MCP, and more.
//
// @contact.name	SecDuckOps
// @license.name	MIT
// @license.url	https://github.com/SecDuckOps/duckops/blob/main/
// @BasePath		/v1
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/SecDuckOps/duckops/internal/cmd"

	_ "net/http/pprof"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if os.Getenv("DUCKOPS_PROFILE") != "" {
		go func() {
			slog.Info("Serving pprof at localhost:6060")
			if httpErr := http.ListenAndServe("localhost:6060", nil); httpErr != nil {
				slog.Error("Failed to pprof listen", "error", httpErr)
			}
		}()
	}

	cmd.Execute()
}
