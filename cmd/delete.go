package main

import (
	"fmt"

	"github.com/riridotdev/sto"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [entries]",
	Short: "Delete an entry from the Sto store",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := sto.ReadStoreOrFail(root)

		for _, arg := range args {
			if err := s.RemoveEntry(arg); err != nil {
				fmt.Printf("Error deleting entry %q: %s\n", arg, err)
				continue
			}
			fmt.Printf("Successfully deleted entry %q\n", arg)
		}

		if err := s.Write(); err != nil {
			sto.Fail("Error writing to store file: %s", err)
		}
	},
}
