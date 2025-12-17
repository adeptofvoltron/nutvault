package main

import (
	"fmt"
	"os"

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
	Use:   "collect",
	Short: "Collect operation",
	Long:  "Collect operation placeholder - implementation coming soon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("collect command - placeholder")
	},
}

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
	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(fillCmd)
	rootCmd.AddCommand(swapCmd)
	rootCmd.AddCommand(clearCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

