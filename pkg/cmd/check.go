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

// NewCheckCommand runs a permission Check against a remote Permify gRPC server.
func NewCheckCommand() *cobra.Command {
	var (
		credentialsPath string
		tenantID        string
		subjectStr      string
		resourceStr     string
		permission      string
	)

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check whether a subject has a permission on a resource",
		Long: `Calls the Permify Permission.Check RPC.

Subject is --entity (e.g. user:1). Resource is --resource (e.g. document:1).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(tenantID) == "" {
				return fmt.Errorf("--tenant-id is required")
			}
			if strings.TrimSpace(subjectStr) == "" {
				return fmt.Errorf("--entity is required")
			}
			if strings.TrimSpace(resourceStr) == "" {
				return fmt.Errorf("--resource is required")
			}
			if strings.TrimSpace(permission) == "" {
				return fmt.Errorf("--permission is required")
			}

			subject, err := ParseSubjectRef(subjectStr)
			if err != nil {
				return fmt.Errorf("parse subject: %w", err)
			}
			entity, err := ParseEntityRef(resourceStr)
			if err != nil {
				return fmt.Errorf("parse resource: %w", err)
			}

			conn, err := DialGRPC(credentialsPath)
			if err != nil {
				return fmt.Errorf("connect to permify: %w", err)
			}
			defer func() { _ = conn.Close() }()

			rpcCtx, cancel := newGRPCCallContext(cmd.Context())
			defer cancel()

			client := basev1.NewPermissionClient(conn)
			return runPermissionCheck(rpcCtx, os.Stdout, client, tenantID, entity, subject, permission)
		},
	}

	fs := cmd.Flags()
	fs.StringVar(&credentialsPath, "credentials", "", "path to gRPC credentials file (default: $HOME/.permify/credentials)")
	fs.StringVar(&tenantID, "tenant-id", "", "tenant identifier (required)")
	fs.StringVar(&subjectStr, "entity", "", "subject as type:id (e.g. user:1)")
	fs.StringVar(&resourceStr, "resource", "", "resource entity as type:id (e.g. document:1)")
	fs.StringVar(&permission, "permission", "", "permission name to evaluate (e.g. view)")
	_ = cmd.MarkFlagRequired("tenant-id")
	_ = cmd.MarkFlagRequired("entity")
	_ = cmd.MarkFlagRequired("resource")
	_ = cmd.MarkFlagRequired("permission")

	return cmd
}

type permissionCheckClient interface {
	Check(ctx context.Context, in *basev1.PermissionCheckRequest, opts ...grpc.CallOption) (*basev1.PermissionCheckResponse, error)
}

func runPermissionCheck(
	ctx context.Context,
	w io.Writer,
	client permissionCheckClient,
	tenantID string,
	entity *basev1.Entity,
	subject *basev1.Subject,
	permission string,
) error {
	req := &basev1.PermissionCheckRequest{
		TenantId:   tenantID,
		Metadata:   &basev1.PermissionCheckRequestMetadata{},
		Entity:     entity,
		Subject:    subject,
		Permission: permission,
	}

	resp, err := client.Check(ctx, req)
	if err != nil {
		return GRPCStatusError(err)
	}
	formatCheckResult(w, resp)
	return nil
}
