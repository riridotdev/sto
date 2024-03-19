package main

import (
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename [entry] [new-name]",
	Short: "Change the name under which a Sto entry is managed",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		entryName := args[0]
		newName := args[1]

		s := readStoreOrFail(root)

		if err := s.renameEntry(entryName, newName); err != nil {
			fail("Error renaming entry: %s", err)
		}

		if err := s.write(); err != nil {
			fail("Error writing to store: %s", err)
		}
	},
}
