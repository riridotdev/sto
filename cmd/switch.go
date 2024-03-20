package main

import "github.com/spf13/cobra"

var switchCmd = &cobra.Command{
	Use:   "switch [profile]",
	Short: "Change to a different Sto profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		panic("Unimplemented")
	},
}
