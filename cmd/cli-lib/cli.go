// Package clilib provides shared CLI utilities for Kranix CLI tools.
// This library includes common flag parsing, output formatting, and
// configuration management for any Kranix CLI tool.
package clilib

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
)

// GlobalFlags contains common flags used across Kranix CLI tools
type GlobalFlags struct {
	ServerURL     string
	APIKey        string
	Namespace     string
	Output        OutputFormat
	Verbose       bool
	Timeout       int
	ConfigFile    string
	Context       string
	SkipTLSVerify bool
}

// AddGlobalFlags adds common flags to a cobra command
func AddGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().StringVarP(&flags.ServerURL, "server", "s", "http://localhost:8080", "Kranix API server URL")
	cmd.Flags().StringVarP(&flags.APIKey, "api-key", "k", "", "API key for authentication")
	cmd.Flags().StringVarP(&flags.Namespace, "namespace", "n", "default", "Namespace scope")
	cmd.Flags().StringVarP((*string)(&flags.Output), "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().IntVar(&flags.Timeout, "timeout", 30, "Request timeout in seconds")
	cmd.Flags().StringVar(&flags.ConfigFile, "config", "", "Path to config file")
	cmd.Flags().StringVar(&flags.Context, "context", "", "Context to use")
	cmd.Flags().BoolVar(&flags.SkipTLSVerify, "insecure-skip-tls-verify", false, "Skip TLS verification")
}

// BindFlagsToEnv binds flags to environment variables
func BindFlagsToEnv(cmd *cobra.Command, bindings map[string]string) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if envVar, ok := bindings[f.Name]; ok {
			if err := os.Setenv(envVar, f.Value.String()); err == nil {
				f.Changed = true
			}
		}
	})
}

// ValidateFlags validates the global flags
func ValidateFlags(flags *GlobalFlags) error {
	if flags.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}

	if flags.Output != FormatTable && flags.Output != FormatJSON && flags.Output != FormatYAML {
		return fmt.Errorf("invalid output format: %s (must be table, json, or yaml)", flags.Output)
	}

	if flags.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// PrintOutput prints data in the specified format
func PrintOutput(format OutputFormat, data interface{}) error {
	switch format {
	case FormatJSON:
		return printJSON(data)
	case FormatYAML:
		return printYAML(data)
	case FormatTable:
		return printTable(data)
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}

// printJSON prints data as JSON
func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printYAML prints data as YAML
func printYAML(data interface{}) error {
	return yaml.NewEncoder(os.Stdout).Encode(data)
}

// TablePrinter is an interface for types that can print themselves as tables
type TablePrinter interface {
	PrintTable() error
}

// RegisterTablePrinter registers a table printer for a specific type
var tablePrinters = make(map[interface{}]func(interface{}) error)

// RegisterTablePrinterFunc registers a function to print a specific type as a table
func RegisterTablePrinterFunc(dataType interface{}, printer func(interface{}) error) {
	tablePrinters[dataType] = printer
}

// printTable prints data as a table using registered printers
func printTable(data interface{}) error {
	// Check if data implements TablePrinter interface
	if printer, ok := data.(TablePrinter); ok {
		return printer.PrintTable()
	}

	// Check if there's a registered printer for this type
	dataType := fmt.Sprintf("%T", data)
	if printer, ok := tablePrinters[dataType]; ok {
		return printer(data)
	}

	// Default to JSON
	return printJSON(data)
}

// PrintError prints an error message to stderr
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// PrintSuccess prints a success message to stdout
func PrintSuccess(msg string) {
	fmt.Fprintf(os.Stdout, "✓ %s\n", msg)
}

// PrintWarning prints a warning message to stderr
func PrintWarning(msg string) {
	fmt.Fprintf(os.Stderr, "Warning: %s\n", msg)
}

// PrintInfo prints an info message to stdout
func PrintInfo(msg string) {
	fmt.Fprintf(os.Stdout, "Info: %s\n", msg)
}

// PrintDebug prints a debug message to stderr if verbose mode is enabled
func PrintDebug(verbose bool, msg string) {
	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: %s\n", msg)
	}
}

// ConfirmAction prompts the user for confirmation
func ConfirmAction(message string) bool {
	fmt.Printf("%s [y/N]: ", message)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

// TableBuilder helps build formatted tables
type TableBuilder struct {
	headers []string
	rows    [][]string
	widths  []int
}

// NewTableBuilder creates a new table builder
func NewTableBuilder(headers []string) *TableBuilder {
	tb := &TableBuilder{
		headers: headers,
		widths:  make([]int, len(headers)),
	}
	for i, h := range headers {
		tb.widths[i] = len(h)
	}
	return tb
}

// AddRow adds a row to the table
func (tb *TableBuilder) AddRow(row []string) {
	if len(row) != len(tb.headers) {
		return
	}
	for i, cell := range row {
		if len(cell) > tb.widths[i] {
			tb.widths[i] = len(cell)
		}
	}
	tb.rows = append(tb.rows, row)
}

// Print prints the table
func (tb *TableBuilder) Print() error {
	// Print headers
	for i, h := range tb.headers {
		fmt.Printf("%-*s", tb.widths[i]+2, h)
	}
	fmt.Println()

	// Print separator
	for _, w := range tb.widths {
		fmt.Printf("%s", strings.Repeat("-", w+2))
	}
	fmt.Println()

	// Print rows
	for _, row := range tb.rows {
		for i, cell := range row {
			fmt.Printf("%-*s", tb.widths[i]+2, cell)
		}
		fmt.Println()
	}

	return nil
}

// Spinner provides a simple loading indicator
type Spinner struct {
	chars  []string
	index  int
	active bool
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
	return &Spinner{
		chars: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Start starts the spinner
func (s *Spinner) Start(message string) {
	s.active = true
	s.index = 0
	go s.animate(message)
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.active = false
}

// animate animates the spinner
func (s *Spinner) animate(message string) {
	for s.active {
		fmt.Printf("\r%s %s", s.chars[s.index], message)
		s.index = (s.index + 1) % len(s.chars)
	}
	fmt.Print("\r")
}

// ProgressTracker tracks progress for long-running operations
type ProgressTracker struct {
	total   int
	current int
	message string
	verbose bool
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(total int, message string, verbose bool) *ProgressTracker {
	return &ProgressTracker{
		total:   total,
		current: 0,
		message: message,
		verbose: verbose,
	}
}

// Increment increments the progress counter
func (pt *ProgressTracker) Increment() {
	pt.current++
	if pt.verbose {
		percentage := float64(pt.current) / float64(pt.total) * 100
		fmt.Printf("\r%s: %d/%d (%.1f%%)", pt.message, pt.current, pt.total, percentage)
	}
	if pt.current >= pt.total {
		fmt.Println()
	}
}

// Complete marks the progress as complete
func (pt *ProgressTracker) Complete() {
	pt.current = pt.total
	if pt.verbose {
		fmt.Printf("\r%s: Complete!\n", pt.message)
	}
}
