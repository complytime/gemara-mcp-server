package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/complytime/gemara-mcp-server/mcp"
	"github.com/spf13/cobra"

	"github.com/complytime/gemara-mcp-server/version"
)

var rootCmd = &cobra.Command{
	Use:   "gemara-mcp-server",
	Short: "Gemara CUE MCP Server",
	Long:  "A Model Context Protocol server for Gemara (GRC Engineering Model for Automated Risk Assessment)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := mcp.ServerConfig{
			Version: version.GetVersion(),
		}
		server := mcp.NewServer(&cfg)
		return server.Start()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Gemara CUE MCP Server %s\n", version.GetVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	
	// Set up structured logging with slog
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Command execution failed", "error", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
