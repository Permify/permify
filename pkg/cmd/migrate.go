package cmd

import (
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/Permify/permify/internal/storage"
)

const (
	target         = "target"
	databaseEngine = "database-engine"
	databaseURI    = "database-uri"
)

// NewMigrateCommand - Creates new migrate command
func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate the database",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(NewMigrateUpCommand())
	cmd.AddCommand(NewMigrateDownCommand())
	cmd.AddCommand(NewMigrateStatusCommand())
	cmd.AddCommand(NewMigrateResetCommand())

	return cmd
}

// NewMigrateUpCommand - Creates new migrate up command
func NewMigrateUpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "migrate the DB to the most recent version available",
		RunE:  migrateUp(),
		Args:  cobra.NoArgs,
	}

	// add flags to the migration up command
	cmd.PersistentFlags().String(target, "0", "version to migrate")
	cmd.PersistentFlags().String(databaseEngine, "", "database engine")
	cmd.PersistentFlags().String(databaseURI, "", "database URI")

	return cmd
}

// NewMigrateDownCommand - Creates new migrate down command
func NewMigrateDownCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "roll back the migration for the current database",
		RunE:  migrateDown(),
		Args:  cobra.NoArgs,
	}

	// add flags to the migration down command
	cmd.PersistentFlags().String(target, "0", "version to rollback")
	cmd.PersistentFlags().String(databaseEngine, "", "database engine")
	cmd.PersistentFlags().String(databaseURI, "", "database URI")

	return cmd
}

// NewMigrateStatusCommand - Creates new migrate status command
func NewMigrateStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "dump the migration status for the current DB",
		RunE:  migrateStatus(),
		Args:  cobra.NoArgs,
	}

	// add flags to the migration status command
	cmd.PersistentFlags().String(databaseEngine, "", "database engine")
	cmd.PersistentFlags().String(databaseURI, "", "database URI")

	return cmd
}

// NewMigrateResetCommand - Creates new migrate reset command
func NewMigrateResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "roll back all migrations",
		RunE:  migrateReset(),
		Args:  cobra.NoArgs,
	}

	// add flags to the migration status command
	cmd.PersistentFlags().String(databaseEngine, "", "database engine")
	cmd.PersistentFlags().String(databaseURI, "", "database URI")

	return cmd
}

// migrateUp - permify migrate up command
func migrateUp() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		flags, err := getFlags(cmd, []string{target, databaseEngine, databaseURI})
		if err != nil {
			return err
		}

		p, err := strconv.ParseInt(flags[target], 10, 64)
		if err != nil {
			return err
		}

		if p == 0 {
			return storage.MigrateUp(flags[databaseEngine], flags[databaseURI])
		}

		return storage.MigrateUpTo(flags[databaseEngine], flags[databaseURI], p)
	}
}

// migrateDown - permify migrate down command
func migrateDown() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		flags, err := getFlags(cmd, []string{target, databaseEngine, databaseURI})
		if err != nil {
			return err
		}

		p, err := strconv.ParseInt(flags[target], 10, 64)
		if err != nil {
			return err
		}

		if p == 0 {
			return storage.MigrateDown(flags[databaseEngine], flags[databaseURI])
		}

		return storage.MigrateDownTo(flags[databaseEngine], flags[databaseURI], p)
	}
}

// migrateReset - permify migrate reset command
func migrateReset() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		flags, err := getFlags(cmd, []string{databaseEngine, databaseURI})
		if err != nil {
			return err
		}

		return storage.MigrateReset(flags[databaseEngine], flags[databaseURI])
	}
}

// migrateStatus - permify migrate status command
func migrateStatus() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		flags, err := getFlags(cmd, []string{databaseEngine, databaseURI})
		if err != nil {
			return err
		}

		return storage.MigrateStatus(flags[databaseEngine], flags[databaseURI])
	}
}

func getFlags(cmd *cobra.Command, flags []string) (map[string]string, error) {
	resp := make(map[string]string, len(flags))

	// Initialize multiErr
	multiErr := &multierror.Error{}

	for i := range flags {
		value, err := cmd.Flags().GetString(flags[i])
		if err != nil {
			multiErr.Errors = append(multiErr.Errors, err)
		}
		resp[flags[i]] = value
	}

	if multiErr.ErrorOrNil() != nil {
		return nil, multiErr
	}

	return resp, nil
}
