package main

import (
	"errors"
	"fmt"

	"github.com/riridotdev/sto"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List active and inactive symlinks for the selected profile",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		s := sto.ReadStoreOrFail(root)

		for _, entry := range s.Entries() {
			linked, err := s.CheckEntry(entry.Name)
			if err != nil {
				var existingSymlink sto.ErrExistingSymlinkMismatch

				switch {
				case errors.As(err, &existingSymlink):
					fmt.Printf("%s:\t[Unlinked]\t%s/%s -> %s\n", entry.Name, s.Root, entry.Source, entry.Destination)
					fmt.Printf("\tConflict: %s -> %s\n", entry.Destination, existingSymlink)

				case errors.Is(err, sto.ErrExistingFileAtDestination):
					fmt.Printf("%s:\t[Unlinked]\t%s/%s -> %s\n", entry.Name, s.Root, entry.Source, entry.Destination)
					fmt.Printf("\tFile Already Exists: %s\n", entry.Destination)

				case errors.Is(err, sto.ErrEntrySourceInvalid):
					fmt.Printf("%s:\t[Broken]\t%s/%s -> %s\n", entry.Name, s.Root, entry.Source, entry.Destination)

				default:
					fmt.Printf("Error checking entry %q: %s\n", entry.Name, err)
				}

				continue
			}

			if linked {
				fmt.Printf("%s:\t[Linked]\t%s/%s -> %s\n", entry.Name, s.Root, entry.Source, entry.Destination)
			} else {
				fmt.Printf("%s:\t[Unlinked]\t%s/%s -> %s\n", entry.Name, s.Root, entry.Source, entry.Destination)
			}
		}
	},
}
