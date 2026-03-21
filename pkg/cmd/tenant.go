package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// NewTenantCommand groups tenant RPCs (create, list, delete).
func NewTenantCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "tenant",
		Short: "Create, list, or delete tenants on a remote Permify server",
	}
	root.AddCommand(newTenantCreateCommand())
	root.AddCommand(newTenantListCommand())
	root.AddCommand(newTenantDeleteCommand())
	return root
}

func newTenantCreateCommand() *cobra.Command {
	var credentialsPath, id, name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tenant (Tenancy.Create)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(id) == "" {
				return fmt.Errorf("--id is required")
			}
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("--name is required")
			}

			conn, err := DialGRPC(credentialsPath)
			if err != nil {
				return fmt.Errorf("connect to permify: %w", err)
			}
			defer func() { _ = conn.Close() }()

			rpcCtx, cancel := newGRPCCallContext(cmd.Context())
			defer cancel()

			client := basev1.NewTenancyClient(conn)
			return runTenantCreate(rpcCtx, os.Stdout, client, id, name)
		},
	}

	fs := cmd.Flags()
	fs.StringVar(&credentialsPath, "credentials", "", "path to gRPC credentials file (default: $HOME/.permify/credentials)")
	fs.StringVar(&id, "id", "", "tenant id (required)")
	fs.StringVar(&name, "name", "", "tenant display name (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTenantListCommand() *cobra.Command {
	var credentialsPath string
	var pageSize uint32

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tenants (Tenancy.List)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if pageSize == 0 {
				pageSize = 100
			}

			conn, err := DialGRPC(credentialsPath)
			if err != nil {
				return fmt.Errorf("connect to permify: %w", err)
			}
			defer func() { _ = conn.Close() }()

			rpcCtx, cancel := newGRPCCallContext(cmd.Context())
			defer cancel()

			client := basev1.NewTenancyClient(conn)
			return runTenantList(rpcCtx, os.Stdout, client, pageSize, "")
		},
	}

	fs := cmd.Flags()
	fs.StringVar(&credentialsPath, "credentials", "", "path to gRPC credentials file (default: $HOME/.permify/credentials)")
	fs.Uint32Var(&pageSize, "page-size", 100, "page size (1–100)")
	return cmd
}

func newTenantDeleteCommand() *cobra.Command {
	var credentialsPath, id string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a tenant (Tenancy.Delete)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(id) == "" {
				return fmt.Errorf("--id is required")
			}

			conn, err := DialGRPC(credentialsPath)
			if err != nil {
				return fmt.Errorf("connect to permify: %w", err)
			}
			defer func() { _ = conn.Close() }()

			rpcCtx, cancel := newGRPCCallContext(cmd.Context())
			defer cancel()

			client := basev1.NewTenancyClient(conn)
			return runTenantDelete(rpcCtx, os.Stdout, client, id)
		},
	}

	fs := cmd.Flags()
	fs.StringVar(&credentialsPath, "credentials", "", "path to gRPC credentials file (default: $HOME/.permify/credentials)")
	fs.StringVar(&id, "id", "", "tenant id to delete (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

type tenancyCreateClient interface {
	Create(ctx context.Context, in *basev1.TenantCreateRequest, opts ...grpc.CallOption) (*basev1.TenantCreateResponse, error)
}

type tenancyListClient interface {
	List(ctx context.Context, in *basev1.TenantListRequest, opts ...grpc.CallOption) (*basev1.TenantListResponse, error)
}

type tenancyDeleteClient interface {
	Delete(ctx context.Context, in *basev1.TenantDeleteRequest, opts ...grpc.CallOption) (*basev1.TenantDeleteResponse, error)
}

func runTenantCreate(ctx context.Context, w io.Writer, client tenancyCreateClient, id, name string) error {
	resp, err := client.Create(ctx, &basev1.TenantCreateRequest{
		Id:   id,
		Name: name,
	})
	if err != nil {
		return GRPCStatusError(err)
	}
	t := resp.GetTenant()
	if t == nil {
		_, _ = fmt.Fprintln(w, "Tenant created.")
		return nil
	}
	_, _ = fmt.Fprintf(w, "Tenant created: %s (%s)\n", t.GetId(), t.GetName())
	if ts := t.GetCreatedAt(); ts != nil {
		_, _ = fmt.Fprintf(w, "Created at: %s\n", formatTimestamp(ts))
	}
	return nil
}

func runTenantList(ctx context.Context, w io.Writer, client tenancyListClient, pageSize uint32, continuousToken string) error {
	resp, err := client.List(ctx, &basev1.TenantListRequest{
		PageSize:        pageSize,
		ContinuousToken: continuousToken,
	})
	if err != nil {
		return GRPCStatusError(err)
	}
	tenants := resp.GetTenants()
	if len(tenants) == 0 {
		_, _ = fmt.Fprintln(w, "No tenants.")
		return nil
	}
	_, _ = fmt.Fprintf(w, "Tenants (%d):\n", len(tenants))
	for _, t := range tenants {
		line := fmt.Sprintf("  • %s — %s", t.GetId(), t.GetName())
		if ts := t.GetCreatedAt(); ts != nil {
			line += fmt.Sprintf(" (created %s)", formatTimestamp(ts))
		}
		_, _ = fmt.Fprintln(w, line)
	}
	if tok := resp.GetContinuousToken(); tok != "" {
		_, _ = fmt.Fprintln(w, "Note: more tenants may be available (pagination token returned).")
	}
	return nil
}

func runTenantDelete(ctx context.Context, w io.Writer, client tenancyDeleteClient, id string) error {
	resp, err := client.Delete(ctx, &basev1.TenantDeleteRequest{Id: id})
	if err != nil {
		return GRPCStatusError(err)
	}
	_, _ = fmt.Fprintf(w, "Deleted tenant: %s\n", resp.GetTenantId())
	return nil
}

func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	t := ts.AsTime()
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
