package main

import "github.com/spf13/cobra"

var addCmd = &cobra.Command{
	Use:   "add [source-path] [destination-path]",
	Short: "Add an item to the Sto store",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		destination := args[1]

		s := readStoreOrFail(root)

		if err := s.add(source, destination); err != nil {
			fail("Error adding item to store: %s", err)
		}

		if err := s.write(); err != nil {
			fail("Error writing to store file: %s", err)
		}
	},
}
