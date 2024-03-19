package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move [item] [new-path]",
	Short: "Move a file or directory managed under the Sto root to a new location",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		entryName := args[0]
		newPath := args[1]

		s := readStoreOrFail(root)

		if err := s.moveEntry(entryName, newPath); err != nil {
			fail("Error moving entry: %s", err)
		}

		if err := s.write(); err != nil {
			fail("Error writing store: %s", err)
		}

		fmt.Printf("Successfully moved entry %s to %s", entryName, newPath)
	},
}
