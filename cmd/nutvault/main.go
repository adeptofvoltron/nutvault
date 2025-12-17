package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adeptofvoltron/nutvault/internal/env"
	"github.com/adeptofvoltron/nutvault/internal/vault"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nutvault",
	Short: "nutvault - A CLI tool for managing env vault operations",
	Long: `nutvault is a CLI tool that provides commands for managing vault operations.
It supports operations like collect, fill, swap, and remove.`,
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
	fillEnvFile    string
	fillKeyFile    string
	swapEnvFile    string
	swapKeyFile    string
	removeKeyFile  string
)

var fillCmd = &cobra.Command{
	Use:   "fill [projectName]",
	Short: "Fill empty variables in .env file with values from vault",
	Long: `Fill reads variables from a vault project and fills only empty variables in a .env file.
Variables that already have values are not modified.

Examples:
  # Fill empty variables in default .env file
  nutvault fill myproject

  # Fill empty variables in custom .env file
  nutvault fill myproject --env-file .env.production

  # Fill with custom key file
  nutvault fill myproject --key-file ~/.nutvault/mykey.hex`,
	Args: cobra.ExactArgs(1),
	RunE: runFill,
}

var swapCmd = &cobra.Command{
	Use:   "swap [projectName]",
	Short: "Swap all variables in .env file with values from vault",
	Long: `Swap reads variables from a vault project and replaces all variable values in a .env file.
All existing variable values will be overwritten with values from the vault.

Examples:
  # Swap all variables in default .env file
  nutvault swap myproject

  # Swap all variables in custom .env file
  nutvault swap myproject --env-file .env.production

  # Swap with custom key file
  nutvault swap myproject --key-file ~/.nutvault/mykey.hex`,
	Args: cobra.ExactArgs(1),
	RunE: runSwap,
}

var removeCmd = &cobra.Command{
	Use:   "remove [projectName]",
	Short: "Remove (delete) a vault project",
	Long: `Remove deletes an entire vault project and all its contents.
This operation cannot be undone. All variables stored in the project will be permanently deleted.

Examples:
  # Remove project with default key
  nutvault remove myproject

  # Remove project with custom key file
  nutvault remove myproject --key-file ~/.nutvault/mykey.hex`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all vault projects",
	Long: `List displays all vault projects stored in ~/.nutvault/projects/.
For each project, it shows the project name, hash, path, and number of variables.

Examples:
  # List all projects
  nutvault list`,
	Args: cobra.NoArgs,
	RunE: runList,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  "Print the version number of nutvault",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nutvault version 1.0.0")
	},
}

func init() {
	collectCmd.Flags().StringVarP(&collectEnvFile, "env-file", "e", ".env", "Path to .env file (default: .env in current directory)")
	collectCmd.Flags().StringVarP(&collectKeyFile, "key-file", "k", "", "Path to key file in hex format (default: use default user key)")

	fillCmd.Flags().StringVarP(&fillEnvFile, "env-file", "e", ".env", "Path to .env file (default: .env in current directory)")
	fillCmd.Flags().StringVarP(&fillKeyFile, "key-file", "k", "", "Path to key file in hex format (default: use default user key)")

	swapCmd.Flags().StringVarP(&swapEnvFile, "env-file", "e", ".env", "Path to .env file (default: .env in current directory)")
	swapCmd.Flags().StringVarP(&swapKeyFile, "key-file", "k", "", "Path to key file in hex format (default: use default user key)")

	removeCmd.Flags().StringVarP(&removeKeyFile, "key-file", "k", "", "Path to key file in hex format (default: use default user key)")

	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(fillCmd)
	rootCmd.AddCommand(swapCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
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

// runFill executes the fill command.
// It fills only empty variables in .env file with values from vault.
func runFill(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Determine .env file path (default to .env in current directory)
	envFilePath := fillEnvFile
	if !filepath.IsAbs(envFilePath) {
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

	// Load key (default or from file)
	var key []byte
	if fillKeyFile != "" {
		key, err = loadKeyFromFile(fillKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load key from file: %w", err)
		}
	} else {
		key = nil
	}

	// Open vault project
	project, err := vault.NewProject(projectName, key)
	if err != nil {
		return fmt.Errorf("failed to open vault project: %w", err)
	}

	// Get all variables from vault
	vaultVariables := project.GetAllVariables()
	if len(vaultVariables) == 0 {
		return fmt.Errorf("no variables found in vault project")
	}

	// Fill only empty variables
	filledCount := 0
	for key, value := range vaultVariables {
		// Only fill if variable exists in .env and is empty
		if envFile.IsVariableEmpty(key) {
			if err := envFile.SetVariable(key, value); err != nil {
				return fmt.Errorf("failed to set variable %s: %w", key, err)
			}
			filledCount++
		}
	}

	// Save .env file
	if err := envFile.SaveEnvFile(); err != nil {
		return fmt.Errorf("failed to save .env file: %w", err)
	}

	fmt.Printf("Successfully filled %d empty variable(s) in %s\n", filledCount, envFilePath)
	return nil
}

// runSwap executes the swap command.
// It replaces all variable values in .env file with values from vault.
func runSwap(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Determine .env file path (default to .env in current directory)
	envFilePath := swapEnvFile
	if !filepath.IsAbs(envFilePath) {
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

	// Load key (default or from file)
	var key []byte
	if swapKeyFile != "" {
		key, err = loadKeyFromFile(swapKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load key from file: %w", err)
		}
	} else {
		key = nil
	}

	// Open vault project
	project, err := vault.NewProject(projectName, key)
	if err != nil {
		return fmt.Errorf("failed to open vault project: %w", err)
	}

	// Get all variables from vault
	vaultVariables := project.GetAllVariables()
	if len(vaultVariables) == 0 {
		return fmt.Errorf("no variables found in vault project")
	}

	// Get all variables from .env file
	envVariables := envFile.GetAllVariables()

	// Swap all variables that exist in both .env and vault
	swappedCount := 0
	for key, value := range vaultVariables {
		// Only swap if variable exists in .env file
		if envFile.HasVariable(key) {
			if err := envFile.SetVariable(key, value); err != nil {
				return fmt.Errorf("failed to set variable %s: %w", key, err)
			}
			swappedCount++
		}
	}

	// Save .env file
	if err := envFile.SaveEnvFile(); err != nil {
		return fmt.Errorf("failed to save .env file: %w", err)
	}

	fmt.Printf("Successfully swapped %d variable(s) in %s\n", swappedCount, envFilePath)
	if len(envVariables) > swappedCount {
		fmt.Printf("Note: %d variable(s) in .env file were not found in vault and were not modified\n", len(envVariables)-swappedCount)
	}

	return nil
}

// runRemove executes the remove command.
// It removes the entire vault project and all its contents.
func runRemove(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Load key (default or from file)
	var key []byte
	var err error
	if removeKeyFile != "" {
		key, err = loadKeyFromFile(removeKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load key from file: %w", err)
		}
	} else {
		key = nil
	}

	// Open vault project
	project, err := vault.NewProject(projectName, key)
	if err != nil {
		return fmt.Errorf("failed to open vault project: %w", err)
	}

	// Get project directory before removing
	projectDir := project.GetProjectDir()

	// Remove the project
	if err := project.ClearProject(); err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}

	fmt.Printf("Successfully removed project: %s\n", projectDir)
	return nil
}

// runList executes the list command.
// It lists all vault projects with their information.
func runList(cmd *cobra.Command, args []string) error {
	projects, err := vault.ListProjects()
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No vault projects found.")
		return nil
	}

	fmt.Printf("Found %d vault project(s):\n\n", len(projects))
	for i, project := range projects {
		fmt.Printf("%d. %s\n", i+1, project.Name)
		fmt.Printf("   Hash: %s\n", project.Hash)
		fmt.Printf("   Path: %s\n", project.Path)
		fmt.Printf("   Variables: %d\n", project.VarCount)
		if i < len(projects)-1 {
			fmt.Println()
		}
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
