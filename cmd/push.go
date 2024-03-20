package main

import (
	"errors"
	"fmt"

	"github.com/riridotdev/sto"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [entries]",
	Short: "Create a symlink for the named Sto entry",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := sto.ReadStoreOrFail(root)

		fmt.Printf("Creating symlinks...\n")
		for _, arg := range args {

			if err := s.ApplyEntry(arg); err != nil {
				if errors.Is(err, sto.ErrLinkAlreadyExists) {
					fmt.Printf("\t%s: Already exists\n", arg)
					continue
				}

				fmt.Printf("Error creating symlink for %q: %s\n", arg, err)
				continue
			}

			// TODO: Create a result type to be returned form applyEntry to be used for reporting
			/* fmt.Printf("\t%s: %q -> %q\n", arg) */
		}
		fmt.Println("Finished creating symlinks")
	},
}
