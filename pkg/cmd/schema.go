package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// NewSchemaCommand groups schema RPCs (write, read).
func NewSchemaCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "schema",
		Short: "Read or write authorization schema on a remote Permify server",
	}

	root.AddCommand(newSchemaWriteCommand())
	root.AddCommand(newSchemaReadCommand())
	return root
}

func newSchemaWriteCommand() *cobra.Command {
	var credentialsPath, tenantID, filePath string

	cmd := &cobra.Command{
		Use:   "write",
		Short: "Write schema from a file (Schema.Write)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(tenantID) == "" {
				return fmt.Errorf("--tenant-id is required")
			}
			if strings.TrimSpace(filePath) == "" {
				return fmt.Errorf("--file is required")
			}
			schemaBytes, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read schema file: %w", err)
			}

			conn, err := DialGRPC(credentialsPath)
			if err != nil {
				return fmt.Errorf("connect to permify: %w", err)
			}
			defer func() { _ = conn.Close() }()

			rpcCtx, cancel := newGRPCCallContext(cmd.Context())
			defer cancel()

			client := basev1.NewSchemaClient(conn)
			return runSchemaWrite(rpcCtx, os.Stdout, client, tenantID, string(schemaBytes))
		},
	}

	fs := cmd.Flags()
	fs.StringVar(&credentialsPath, "credentials", "", "path to gRPC credentials file (default: $HOME/.permify/credentials)")
	fs.StringVar(&tenantID, "tenant-id", "", "tenant identifier (required)")
	fs.StringVar(&filePath, "file", "", "path to Permify schema file (.perm)")
	_ = cmd.MarkFlagRequired("tenant-id")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func newSchemaReadCommand() *cobra.Command {
	var credentialsPath, tenantID, schemaVersion string

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read the latest (or pinned) schema (Schema.Read)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(tenantID) == "" {
				return fmt.Errorf("--tenant-id is required")
			}

			conn, err := DialGRPC(credentialsPath)
			if err != nil {
				return fmt.Errorf("connect to permify: %w", err)
			}
			defer func() { _ = conn.Close() }()

			rpcCtx, cancel := newGRPCCallContext(cmd.Context())
			defer cancel()

			client := basev1.NewSchemaClient(conn)
			return runSchemaRead(rpcCtx, os.Stdout, client, tenantID, schemaVersion)
		},
	}

	fs := cmd.Flags()
	fs.StringVar(&credentialsPath, "credentials", "", "path to gRPC credentials file (default: $HOME/.permify/credentials)")
	fs.StringVar(&tenantID, "tenant-id", "", "tenant identifier (required)")
	fs.StringVar(&schemaVersion, "schema-version", "", "optional schema version; empty uses server default")
	_ = cmd.MarkFlagRequired("tenant-id")
	return cmd
}

type schemaWriteClient interface {
	Write(ctx context.Context, in *basev1.SchemaWriteRequest, opts ...grpc.CallOption) (*basev1.SchemaWriteResponse, error)
}

type schemaReadClient interface {
	Read(ctx context.Context, in *basev1.SchemaReadRequest, opts ...grpc.CallOption) (*basev1.SchemaReadResponse, error)
}

func runSchemaWrite(ctx context.Context, w io.Writer, client schemaWriteClient, tenantID, schema string) error {
	resp, err := client.Write(ctx, &basev1.SchemaWriteRequest{
		TenantId: tenantID,
		Schema:   schema,
	})
	if err != nil {
		return GRPCStatusError(err)
	}
	_, _ = fmt.Fprintf(w, "Schema written.\nSchema version: %s\n", resp.GetSchemaVersion())
	return nil
}

func runSchemaRead(ctx context.Context, w io.Writer, client schemaReadClient, tenantID, schemaVersion string) error {
	resp, err := client.Read(ctx, &basev1.SchemaReadRequest{
		TenantId: tenantID,
		Metadata: &basev1.SchemaReadRequestMetadata{
			SchemaVersion: schemaVersion,
		},
	})
	if err != nil {
		return GRPCStatusError(err)
	}
	formatSchemaSummary(w, resp.GetSchema())
	return nil
}
