package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/hashicorp/go-multierror"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
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

	return cmd
}

// NewMigrateUpCommand - Creates new migrate up command
func NewMigrateUpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "migrate the database",
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
		Short: "migration status for the current database",
		RunE:  migrateStatus(),
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

		switch flags[databaseEngine] {
		case "postgres":
			flags[databaseEngine] = "pgx"
		}

		db, err := goose.OpenDBWithDriver(flags[databaseEngine], flags[databaseURI])
		if err != nil {
			color.Warn.Println("migration failed: Database Connection Error")
			return err
		}

		p, err := strconv.ParseInt(flags["target"], 10, 64)
		if err != nil {
			color.Warn.Println("migration failed: bad target input Error")
			return err
		}

		if p == 0 {
			if err := goose.Up(db, "internal/storage/postgres/migrations"); err != nil {
				color.Warn.Println("migration failed: up error " + err.Error())
				return nil
			}

			color.Success.Println("migration successfully up: ✓ ✅ ")
			return nil
		}

		if err := goose.UpTo(db, "internal/storage/postgres/migrations", p); err != nil {
			color.Warn.Println("migration failed: Goose Up Error")
			return nil
		}

		color.Success.Println("migration successfully up: ✓ ✅ ")
		return nil
	}
}

// migrateDown - permify migrate down command
func migrateDown() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		flags, err := getFlags(cmd, []string{target, databaseEngine, databaseURI})
		if err != nil {
			color.Warn.Println("migration failed: flags error")
			return nil
		}

		switch flags[databaseEngine] {
		case "postgres":
			flags[databaseEngine] = "pgx"
		}

		db, err := goose.OpenDBWithDriver(flags[databaseEngine], flags[databaseURI])
		if err != nil {
			color.Warn.Println("migration failed: database connection error")
			return nil
		}

		p, err := strconv.ParseInt(flags["target"], 10, 64)
		if err != nil {
			return nil
		}

		if p == 0 {
			var count int
			err = filepath.Walk("internal/storage/postgres/migrations", func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
					count++
				}
				return nil
			})
			if err != nil {
				color.Warn.Println("migration failed: down error " + err.Error())
				return nil
			}

			for i := 0; i < count; i++ {
				if err := goose.Down(db, "internal/storage/postgres/migrations"); err != nil {
					color.Warn.Println("migration failed: down error " + err.Error())
					return nil
				}
			}

			color.Success.Println("migration successfully down: ✓ ✅ ")
			return nil
		}

		if err := goose.DownTo(db, "internal/storage/postgres/migrations", p); err != nil {
			color.Warn.Println("migration failed: down error " + err.Error())

			return nil
		}

		color.Success.Println("migration successfully down: ✓ ✅ ")
		return nil
	}
}

// migrateStatus - permify migrate status command
func migrateStatus() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		flags, err := getFlags(cmd, []string{databaseEngine, databaseURI})
		if err != nil {
			color.Warn.Println("migration failed: flags error")
			return err
		}

		switch flags[databaseEngine] {
		case "postgres":
			flags[databaseEngine] = "pgx"
		}

		db, err := goose.OpenDBWithDriver(flags[databaseEngine], flags[databaseURI])
		if err != nil {
			color.Warn.Println("migration failed: database connection error")
			return nil
		}

		if err := goose.Status(db, "internal/storage/postgres/migrations"); err != nil {
			color.Warn.Println("migration failed: check status error " + err.Error())
			return nil
		}

		return nil
	}
}

func getFlags(cmd *cobra.Command, flags []string) (map[string]string, error) {
	resp := make(map[string]string, len(flags))

	var multiErr *multierror.Error

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
