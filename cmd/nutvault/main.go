package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bernard/nutvault/internal/env"
	"github.com/bernard/nutvault/internal/vault"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nutvault",
	Short: "nutvault - A CLI tool for managing env vault operations",
	Long: `nutvault is a CLI tool that provides commands for managing vault operations.
It supports operations like collect, fill, swap, and clear.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, show help
		cmd.Help()
	},
}

var collectCmd = &cobra.Command{
	Use:   "collect [projectName]",
	Short: "Collect variables from .env file and save to vault",
	Long: `Collect reads all variables from a .env file and saves them to a vault project.
The project will be stored at ~/.nutvault/projects/<projectName>.<hash>.

Examples:
  # Collect from default .env file with default key
  nutvault collect myproject

  # Collect from custom .env file
  nutvault collect myproject --env-file .env.production

  # Collect with custom key file
  nutvault collect myproject --key-file ~/.nutvault/mykey.hex`,
	Args: cobra.ExactArgs(1),
	RunE: runCollect,
}

var (
	collectEnvFile string
	collectKeyFile string
)

var fillCmd = &cobra.Command{
	Use:   "fill",
	Short: "Fill operation",
	Long:  "Fill operation placeholder - implementation coming soon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("fill command - placeholder")
	},
}

var swapCmd = &cobra.Command{
	Use:   "swap",
	Short: "Swap operation",
	Long:  "Swap operation placeholder - implementation coming soon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("swap command - placeholder")
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear operation",
	Long:  "Clear operation placeholder - implementation coming soon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("clear command - placeholder")
	},
}

func init() {
	collectCmd.Flags().StringVarP(&collectEnvFile, "env-file", "e", ".env", "Path to .env file (default: .env in current directory)")
	collectCmd.Flags().StringVarP(&collectKeyFile, "key-file", "k", "", "Path to key file in hex format (default: use default user key)")

	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(fillCmd)
	rootCmd.AddCommand(swapCmd)
	rootCmd.AddCommand(clearCmd)
}

// loadKeyFromFile loads a key from a file in hex format.
// The file should contain exactly 64 hex characters (32 bytes).
func loadKeyFromFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Remove whitespace and newlines
	hexString := strings.TrimSpace(string(data))
	hexString = strings.ReplaceAll(hexString, "\n", "")
	hexString = strings.ReplaceAll(hexString, " ", "")

	// Validate length (64 hex chars = 32 bytes)
	if len(hexString) != 64 {
		return nil, fmt.Errorf("invalid key length: expected 64 hex characters (32 bytes), got %d", len(hexString))
	}

	// Decode hex
	key, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex key: %w", err)
	}

	return key, nil
}

// runCollect executes the collect command.
func runCollect(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Determine .env file path (default to .env in current directory)
	envFilePath := collectEnvFile
	if !filepath.IsAbs(envFilePath) {
		// If relative path, make it absolute from current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		envFilePath = filepath.Join(cwd, envFilePath)
	}

	// Read .env file
	envFile, err := env.ReadEnvFile(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	// Get all variables from .env file
	variables := envFile.GetAllVariables()
	if len(variables) == 0 {
		return fmt.Errorf("no variables found in .env file")
	}

	// Load key (default or from file)
	var key []byte
	if collectKeyFile != "" {
		key, err = loadKeyFromFile(collectKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load key from file: %w", err)
		}
	} else {
		// Use default key (nil will trigger default key generation in vault.NewProject)
		key = nil
	}

	// Create or open vault project
	project, err := vault.NewProject(projectName, key)
	if err != nil {
		return fmt.Errorf("failed to create vault project: %w", err)
	}

	// Save all variables to vault
	savedCount := 0
	for varKey, varValue := range variables {
		if err := project.SaveVariable(varKey, varValue); err != nil {
			return fmt.Errorf("failed to save variable %s: %w", varKey, err)
		}
		savedCount++
	}

	// Success message
	projectDir := project.GetProjectDir()
	fmt.Printf("Successfully collected %d variable(s) from %s\n", savedCount, envFilePath)
	fmt.Printf("Project saved to: %s\n", projectDir)

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

