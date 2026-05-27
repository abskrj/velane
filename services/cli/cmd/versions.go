package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Manage snippet versions",
}

var versionsListCmd = &cobra.Command{
	Use:   "list <snippet-id>",
	Short: "List all versions for a snippet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snippetID := args[0]
		c, err := newClient()
		if err != nil {
			return err
		}

		versions, err := c.ListVersions(context.Background(), snippetID)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NUM\tID\tSTATUS\tCREATED")
		for _, v := range versions {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", v.VersionNumber, v.ID, v.Status, v.CreatedAt)
		}
		return w.Flush()
	},
}

var versionsPublishCmd = &cobra.Command{
	Use:   "publish <snippet-id> <version-num> <env>",
	Short: "Publish a version to an environment",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		snippetID := args[0]
		versionNum, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		env := args[2]

		c, err := newClient()
		if err != nil {
			return err
		}

		v, err := c.PublishVersion(context.Background(), snippetID, versionNum, env)
		if err != nil {
			return err
		}

		fmt.Printf("Published version %d to %s (status: %s)\n", v.VersionNumber, env, v.Status)
		return nil
	},
}

func init() {
	versionsCmd.AddCommand(versionsListCmd)
	versionsCmd.AddCommand(versionsPublishCmd)
}
