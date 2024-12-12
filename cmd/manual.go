package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// nolint: gochecknoglobals
var manualCmd = &cobra.Command{
	Use:   "manual",
	Short: "Shows the full documentation for this tool",
	Long:  ``,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(`Omat Manual

-----------
Error Codes
-----------

The following are error codes that may be returned from multiple sub-commands.
See the help for each sub-command for details on what other errors it may
return.

10 - Couldn't find the SSM parameter specifying the name of the role to assume
     for admin access.
11 - Some other error occurred when fetching the admin role name from SSM.
12 - The SSM parameter specifying the name of the role to assume for admin
     access was found, but its value was empty.
13 - AWS API error.  This is a generic error code for any AWS API not
     specifically handled by any command.`)
	},
}

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(manualCmd)
}
