package main

import (
	"github.com/riridotdev/sto"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename [entry] [new-name]",
	Short: "Change the name under which a Sto entry is managed",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		entryName := args[0]
		newName := args[1]

		s := sto.ReadStoreOrFail(root)

		if err := s.RenameEntry(entryName, newName); err != nil {
			sto.Fail("Error renaming entry: %s", err)
		}

		if err := s.Write(); err != nil {
			sto.Fail("Error writing to store: %s", err)
		}
	},
}
