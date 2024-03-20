package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/riridotdev/sto"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise a new Sto profile in the current directory",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		currentDir, err := os.Getwd()
		if err != nil {
			sto.Fail("Error getting current directory: %s", err)
		}

		stoFilePath := fmt.Sprintf("%s/.sto", currentDir)

		_, err = os.Stat(stoFilePath)
		if err == nil {
			sto.Fail("Sto profile already exists in this directory: %q", stoFilePath)
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			sto.Fail("Error checking for existing Sto profile: %s", err)
		}

		file, err := os.Create(stoFilePath)
		if err != nil {
			sto.Fail("Error creating Sto file at %q: %s", stoFilePath, err)
		}
		file.Close()

		if _, err := os.Stat(statePath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				sto.Fail("Error checking for state directory at %q: %s", statePath, err)
			}
			if err := os.MkdirAll(statePath, 0700); err != nil {
				sto.Fail("Error creating state directory at %q: %s", statePath, err)
			}
		}

		file, err = os.OpenFile(fmt.Sprintf("%s/current-profile", statePath), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			sto.Fail("Error creating state file at %q: %s", statePath, err)
		}
		file.WriteString(currentDir)
		file.Close()

		fmt.Printf("Initialised new Sto profile at %q", stoFilePath)
	},
}
