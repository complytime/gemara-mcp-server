package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/complytime/gemara-mcp-server/mcp"
	"github.com/spf13/cobra"

	"github.com/complytime/gemara-mcp-server/version"
)

var (
	transport string
	host      string
	port      int
	debug     bool
)

var rootCmd = &cobra.Command{
	Use:   "gemara-mcp-server",
	Short: "Gemara CUE MCP Server",
	Long:  "A Model Context Protocol server for Gemara (GRC Engineering Model for Automated Risk Assessment)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set up structured logging with slog (after flags are parsed)
		logLevel := slog.LevelInfo
		if debug {
			logLevel = slog.LevelDebug
		}

		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))
		slog.SetDefault(logger)

		// Log startup information for debugging
		slog.Info("Starting Gemara MCP Server",
			"version", version.GetVersion(),
			"working_dir", getWorkingDir(),
			"executable", getExecutablePath(),
			"debug", debug,
		)

		cfg := mcp.ServerConfig{
			Version:   version.GetVersion(),
			Transport: transport,
			Host:      host,
			Port:      port,
			Logger:    logger,
		}

		server, err := mcp.NewServer(&cfg)
		if err != nil {
			slog.Error("Failed to create MCP server", "error", err)
			return fmt.Errorf("failed to create server: %w", err)
		}

		slog.Info("MCP server created successfully", "transport", transport)
		if err := server.Start(); err != nil {
			slog.Error("Server stopped with error", "error", err)
			return fmt.Errorf("server error: %w", err)
		}

		return nil
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

	rootCmd.Flags().StringVar(&transport, "transport", "stdio", "transport mode (stdio/streamable)")
	rootCmd.Flags().StringVar(&host, "host", "0.0.0.0", "host for StreamableHTTP transport")
	rootCmd.Flags().IntVar(&port, "port", 8080, "port for StreamableHTTP transport")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Using debug log level")

	// Set up default logger (will be reconfigured in RunE after flags are parsed)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

// getWorkingDir returns the current working directory, or "unknown" if unavailable
func getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return wd
}

// getExecutablePath returns the path to the executable, or "unknown" if unavailable
func getExecutablePath() string {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return exe
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Command execution failed", "error", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
