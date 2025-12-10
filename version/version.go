package version

// Version information
// These can be set via ldflags during build:
// -X github.com/complytime/gemara-mcp-server/version.Version=...
// -X github.com/complytime/gemara-mcp-server/version.Build=...
var (
	Version = "0.1.0"
	Build   = "dev"
)

// GetVersion returns the version string
func GetVersion() string {
	return Version + "-" + Build
}
