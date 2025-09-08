package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/Permify/permify/pkg/database/postgres"
)

const (
	repairDatabaseEngine = "database-engine"
	repairDatabaseURI    = "database-uri"
	repairBatchSize      = "batch-size"
	repairDryRun         = "dry-run"
	repairVerbose        = "verbose"
	repairRetries        = "retries"
)

// NewRepairCommand - Creates new repair command
func NewRepairCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repair",
		Short: "repair PostgreSQL datastore after migration",
		Long:  "Repair PostgreSQL datastore to fix XID wraparound issues and transaction ID synchronization problems that can occur after database migration.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(NewRepairDatastoreCommand())

	return cmd
}

// NewRepairDatastoreCommand - Creates new repair datastore command
func NewRepairDatastoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datastore",
		Short: "repair PostgreSQL XID counter to prevent wraparound issues",
		Long: `Repair PostgreSQL XID counter using a safe approach.

This command prevents XID wraparound issues by:
- Analyzing maximum referenced XIDs in transactions table
- Advancing PostgreSQL's XID counter to stay ahead of referenced XIDs
- Using safe batch processing to avoid performance impact

This approach does NOT modify existing data, only advances the XID counter.
Use --dry-run to see what would be changed without making actual modifications.`,
		RunE: repairDatastore(),
		Args: cobra.NoArgs,
	}

	// Add flags
	cmd.PersistentFlags().String(repairDatabaseEngine, "postgres", "database engine (only postgres supported)")
	cmd.PersistentFlags().String(repairDatabaseURI, "", "database URI (required)")
	cmd.PersistentFlags().Int(repairBatchSize, 1000, "batch size for XID advancement")
	cmd.PersistentFlags().Bool(repairDryRun, false, "perform a dry run without making changes")
	cmd.PersistentFlags().Bool(repairVerbose, true, "enable verbose logging")
	cmd.PersistentFlags().Int(repairRetries, 3, "maximum number of retries")

	// Mark required flags
	if err := cmd.MarkPersistentFlagRequired(repairDatabaseURI); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}

	return cmd
}

// repairDatastore - permify repair datastore command
func repairDatastore() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Get database URI and engine
		databaseURI, _ := cmd.Flags().GetString(repairDatabaseURI)
		databaseEngine, _ := cmd.Flags().GetString(repairDatabaseEngine)

		// Validate database engine
		if databaseEngine != "postgres" {
			return fmt.Errorf("only postgres database engine is supported for repair")
		}

		// Create PostgreSQL instance
		pg, err := postgres.New(databaseURI)
		if err != nil {
			return fmt.Errorf("failed to create PostgreSQL instance: %w", err)
		}
		defer pg.Close()

		// Parse flags directly from cobra
		batchSize, _ := cmd.Flags().GetInt(repairBatchSize)
		retries, _ := cmd.Flags().GetInt(repairRetries)
		dryRun, _ := cmd.Flags().GetBool(repairDryRun)
		verbose, _ := cmd.Flags().GetBool(repairVerbose)

		// Create repair configuration
		config := &postgres.RepairConfig{
			BatchSize:  batchSize,
			MaxRetries: retries,
			RetryDelay: 100,
			DryRun:     dryRun,
			Verbose:    verbose,
		}

		// Perform repair
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		slog.InfoContext(ctx, "Starting PostgreSQL XID counter repair",
			slog.String("database_uri", maskURL(databaseURI)),
			slog.Int("batch_size", config.BatchSize),
			slog.Bool("dry_run", config.DryRun),
			slog.Int("max_retries", config.MaxRetries))

		start := time.Now()
		result, err := pg.Repair(ctx, config)
		duration := time.Since(start)

		if err != nil {
			return fmt.Errorf("repair failed: %w", err)
		}

		// Print results
		slog.InfoContext(ctx, "Repair completed successfully",
			slog.Duration("duration", duration),
			slog.Int("created_tx_id_fixed", result.CreatedTxIdFixed),
			slog.Int("errors", len(result.Errors)))

		if len(result.Errors) > 0 {
			slog.WarnContext(ctx, "Errors encountered during repair")
			for i, err := range result.Errors {
				slog.WarnContext(ctx, "Repair error",
					slog.Int("error_index", i+1),
					slog.String("error", err.Error()))
			}
		}

		if result.CreatedTxIdFixed > 0 {
			slog.InfoContext(ctx, "XID counter repair completed successfully! Advanced XID counter to prevent wraparound issues.")
		} else {
			slog.InfoContext(ctx, "No XID counter repair needed. PostgreSQL XID counter is already properly positioned.")
		}

		return nil
	}
}

// maskURL masks sensitive information in database URL
func maskURL(url string) string {
	if len(url) < 10 {
		return "***"
	}
	return url[:10] + "***"
}
