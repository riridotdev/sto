package main

import (
	"fmt"

	"github.com/riridotdev/sto"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move [item] [new-path]",
	Short: "Move a file or directory managed under the Sto root to a new location",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		entryName := args[0]
		newPath := args[1]

		s := sto.ReadStoreOrFail(root)

		if err := s.MoveEntry(entryName, newPath); err != nil {
			sto.Fail("Error moving entry: %s", err)
		}

		if err := s.Write(); err != nil {
			sto.Fail("Error writing store: %s", err)
		}

		fmt.Printf("Successfully moved entry %s to %s", entryName, newPath)
	},
}
