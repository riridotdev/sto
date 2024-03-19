package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var userHome = userHomeMust()

var configPath string
var defaultConfigPath = fmt.Sprintf("%s/.config/sto/config", userHome)

var root string
var defaultRoot = fmt.Sprintf("%s/.config/sto", userHome)

var statePath = fmt.Sprintf("%s/.local/share/sto/", userHome)

var version = "Version not set, build using 'make'"

func main() {
	rootCmd := &cobra.Command{
		Use:   "sto",
		Short: "Sto is a simple command line tool for managing symlinks",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			stateFilePath := fmt.Sprintf("%s/current-profile", statePath)
			stateBytes, err := os.ReadFile(stateFilePath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					if !(cmd.Name() == "init" || cmd.Name() == "switch") {
						fail("No active profile set, use 'sto init', or 'sto switch [profile]'")
					}
					return
				}
				fail("Error reading state file at %q: %s", stateFilePath, err)
			}
			root = string(stateBytes)
		},
	}

	rootCmd.AddCommand(
		initCmd,
		listCmd,
		pushCmd,
		pullCmd,
		addCmd,
		renameCmd,
		moveCmd,
		unlinkCmd,
		deleteCmd,
		switchCmd,
		versionCmd,
	)

	rootCmd.PersistentFlags().StringVarP(&root, "root", "", defaultRoot, "Sto root path")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "", defaultConfigPath, "Sto config file path")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
