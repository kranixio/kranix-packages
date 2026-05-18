# Kranix CLI Helper Library

Shared CLI utilities for Kranix CLI tools. This library provides common flag parsing, output formatting, and configuration management for any Kranix CLI tool.

## Features

- **Common Flag Parsing**: Shared flags for server URL, API key, namespace, output format, etc.
- **Output Formatting**: Support for table, JSON, and YAML output formats
- **Configuration Management**: Context-based configuration with file persistence
- **Error Handling**: Consistent error messages and formatting
- **Interactive Prompts**: Confirmation dialogs and progress tracking
- **Table Building**: Helper utilities for building formatted tables
- **Spinner**: Loading indicator for long-running operations

## Installation

```bash
go get github.com/kranix-io/kranix-packages/cmd/cli-lib
```

## Usage

### Basic Setup

```go
package main

import (
    "github.com/kranix-io/kranix-packages/cmd/cli-lib"
    "github.com/spf13/cobra"
)

var globalFlags cli.GlobalFlags

func main() {
    rootCmd := &cobra.Command{
        Use:   "my-kranix-tool",
        Short: "My Kranix CLI tool",
        Run: func(cmd *cobra.Command, args []string) {
            if err := cli.ValidateFlags(&globalFlags); err != nil {
                cli.PrintError(err)
                os.Exit(1)
            }
            // Your command logic here
        },
    }

    cli.AddGlobalFlags(rootCmd, &globalFlags)
    rootCmd.Execute()
}
```

### Output Formatting

```go
// Print data as JSON
data := map[string]interface{}{
    "name": "my-workload",
    "status": "running",
}
cli.PrintOutput(cli.FormatJSON, data)

// Print data as YAML
cli.PrintOutput(cli.FormatYAML, data)

// Print data as table
cli.PrintOutput(cli.FormatTable, data)
```

### Custom Table Printers

```go
type WorkloadList []*Workload

func (wl WorkloadList) PrintTable() error {
    tb := cli.NewTableBuilder([]string{"NAME", "STATUS", "IMAGE"})
    for _, w := range wl {
        tb.AddRow([]string{w.Name, w.Status, w.Image})
    }
    return tb.Print()
}

// Register the printer
cli.RegisterTablePrinterFunc(WorkloadList(nil), func(data interface{}) error {
    return data.(WorkloadList).PrintTable()
})
```

### Configuration Management

```go
// Load configuration
config, err := cli.LoadConfig("")
if err != nil {
    cli.PrintError(err)
    os.Exit(1)
}

// Get current context
ctx, err := config.GetCurrentContext()
if err != nil {
    cli.PrintError(err)
    os.Exit(1)
}

// Add a new context
config.AddContext(cli.Context{
    Name:      "production",
    ServerURL: "https://api.kranix.io",
    APIKey:    "prod-api-key",
    Namespace: "production",
})

// Save configuration
err = config.SaveConfig("")
if err != nil {
    cli.PrintError(err)
}
```

### Interactive Prompts

```go
// Confirm an action
if cli.ConfirmAction("Are you sure you want to delete this workload?") {
    // Proceed with deletion
}

// Show a spinner
spinner := cli.NewSpinner()
spinner.Start("Deploying workload...")
defer spinner.Stop()

// Track progress
tracker := cli.NewProgressTracker(100, "Processing items", true)
for i := 0; i < 100; i++ {
    // Process item
    tracker.Increment()
}
tracker.Complete()
```

### Error and Success Messages

```go
cli.PrintError(fmt.Errorf("operation failed"))
cli.PrintSuccess("Operation completed successfully")
cli.PrintWarning("This is a warning")
cli.PrintInfo("This is informational")
cli.PrintDebug(verbose, "Debug information")
```

## Configuration File Format

The CLI library supports YAML configuration files with the following structure:

```yaml
currentContext: default
contexts:
  default:
    name: default
    serverURL: http://localhost:8080
    apiKey: ""
    namespace: default
    timeout: 30
    insecureSkipTLSVerify: false
  production:
    name: production
    serverURL: https://api.kranix.io
    apiKey: prod-key
    namespace: production
    timeout: 60
    insecureSkipTLSVerify: false
defaults:
  namespace: default
  output: table
  timeout: 30
```

Configuration files are searched in the following locations (in order):
1. `~/.kranix/config.yaml`
2. `~/.config/kranix/config.yaml`
3. `.kranix.yaml` (in current directory)

## Environment Variables

The CLI library supports binding flags to environment variables:

```go
bindings := map[string]string{
    "server":   "KRANIX_SERVER",
    "api-key":  "KRANIX_API_KEY",
    "namespace": "KRANIX_NAMESPACE",
}
cli.BindFlagsToEnv(cmd, bindings)
```

## License

Apache 2.0
