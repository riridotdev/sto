package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [entries]",
	Short: "Delete an entry from the Sto store",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := readStoreOrFail(root)

		for _, arg := range args {
			if err := s.removeEntry(arg); err != nil {
				fmt.Printf("Error deleting entry %q: %s\n", arg, err)
				continue
			}
			fmt.Printf("Successfully deleted entry %q\n", arg)
		}

		if err := s.write(); err != nil {
			fail("Error writing to store file: %s", err)
		}
	},
}
