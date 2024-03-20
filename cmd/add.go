package main

import (
	"github.com/riridotdev/sto"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [source-path] [destination-path]",
	Short: "Add an item to the Sto store",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		destination := args[1]

		s := sto.ReadStoreOrFail(root)

		if err := s.Add(source, destination); err != nil {
			sto.Fail("Error adding item to store: %s", err)
		}

		if err := s.Write(); err != nil {
			sto.Fail("Error writing to store file: %s", err)
		}
	},
}
