package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/riridotdev/sto"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [target-path]",
	Short: "Move a file/directory into the Sto root, and create a symlink in its place",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetDir := args[0]

		s := sto.ReadStoreOrFail(root)

		info, err := os.Stat(targetDir)
		if err != nil {
			sto.Fail("Error reading file %q: %s\n", targetDir, err)
		}
		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			sto.Fail("Error building absolute path: %s\n", err)
		}

		itemName := info.Name()

		if entry, ok := s.Store[itemName]; ok {
			fmt.Printf("Entry for %s already exists\n", itemName)
			fmt.Printf("\t%s -> %s", entry.Source, entry.Destination)
			os.Exit(1)
		}

		storePath := fmt.Sprintf("%s/%s", s.Root, itemName)

		input := bufio.NewReader(os.Stdin)

		fmt.Printf("Sto preparing to move files:\n")
		fmt.Printf("\t%q -> %q\n", absPath, storePath)
		fmt.Printf("Commit changes? [y/n]\n")

		line, _, err := input.ReadLine()
		if err != nil {
			sto.Fail("Error reading input: %s\n", err)
		}
		if !(line[0] == 'y' || line[0] == 'Y') {
			os.Exit(0)
		}

		rollback := func() {
			os.Rename(storePath, targetDir)
		}

		if err := os.Rename(targetDir, storePath); err != nil {
			sto.Fail("Error moving target %q: %s\n", targetDir, err)
		}

		if err := s.Add(itemName, targetDir); err != nil {
			rollback()
			sto.Fail("Error adding link to store: %s", err)
		}

		if err := s.Write(); err != nil {
			rollback()
			sto.Fail("Error writing store: %s", err)
		}
	},
}
