package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink [name]",
	Short: "Remove the symlink to a currently linked Sto item",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := readStoreOrFail(root)

		for _, arg := range args {
			if err := s.unapplyEntry(arg); err != nil {
				fmt.Printf("Error unlinking entry %q: %s\n", arg, err)
				continue
			}
			fmt.Printf("Successfully unlinked %q\n", arg)
		}
	},
}
