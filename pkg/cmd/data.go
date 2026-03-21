package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"

	basev1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// NewDataCommand groups data RPCs (write, read).
func NewDataCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "data",
		Short: "Write or read relationship data on a remote Permify server",
	}
	root.AddCommand(newDataWriteCommand())
	root.AddCommand(newDataReadCommand())
	return root
}

type dataYAMLDoc struct {
	Metadata struct {
		SchemaVersion string `yaml:"schema_version"`
	} `yaml:"metadata"`
	Tuples []dataYAMLTuple `yaml:"tuples"`
}

type dataYAMLTuple struct {
	Entity   dataYAMLRef     `yaml:"entity"`
	Relation string          `yaml:"relation"`
	Subject  dataYAMLSubject `yaml:"subject"`
}

type dataYAMLRef struct {
	Type string `yaml:"type"`
	ID   string `yaml:"id"`
}

type dataYAMLSubject struct {
	Type     string `yaml:"type"`
	ID       string `yaml:"id"`
	Relation string `yaml:"relation"`
}

func newDataWriteCommand() *cobra.Command {
	var credentialsPath, tenantID, filePath string

	cmd := &cobra.Command{
		Use:   "write",
		Short: "Write tuples from a YAML file (Data.Write)",
		Long: `YAML format:

metadata:
  schema_version: ""   # optional; empty uses latest schema
tuples:
  - entity: {type: organization, id: "1"}
    relation: admin
    subject: {type: user, id: "3"}`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(tenantID) == "" {
				return fmt.Errorf("--tenant-id is required")
			}
			if strings.TrimSpace(filePath) == "" {
				return fmt.Errorf("--file is required")
			}
			raw, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read data file: %w", err)
			}

			tuples, meta, err := parseDataYAML(raw)
			if err != nil {
				return fmt.Errorf("parse data YAML: %w", err)
			}

			conn, err := DialGRPC(credentialsPath)
			if err != nil {
				return fmt.Errorf("connect to permify: %w", err)
			}
			defer func() { _ = conn.Close() }()

			rpcCtx, cancel := newGRPCCallContext(cmd.Context())
			defer cancel()

			client := basev1.NewDataClient(conn)
			return runDataWrite(rpcCtx, os.Stdout, client, tenantID, meta, tuples)
		},
	}

	fs := cmd.Flags()
	fs.StringVar(&credentialsPath, "credentials", "", "path to gRPC credentials file (default: $HOME/.permify/credentials)")
	fs.StringVar(&tenantID, "tenant-id", "", "tenant identifier (required)")
	fs.StringVar(&filePath, "file", "", "path to YAML data file")
	_ = cmd.MarkFlagRequired("tenant-id")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func parseDataYAML(raw []byte) ([]*basev1.Tuple, *basev1.DataWriteRequestMetadata, error) {
	var doc dataYAMLDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, nil, fmt.Errorf("parse data YAML: %w", err)
	}
	if len(doc.Tuples) == 0 {
		return nil, nil, fmt.Errorf("data YAML: at least one tuple is required")
	}
	meta := &basev1.DataWriteRequestMetadata{
		SchemaVersion: doc.Metadata.SchemaVersion,
	}
	var out []*basev1.Tuple
	for i, row := range doc.Tuples {
		if row.Entity.Type == "" || row.Entity.ID == "" {
			return nil, nil, fmt.Errorf("tuple %d: entity type and id are required", i)
		}
		if row.Subject.Type == "" || row.Subject.ID == "" {
			return nil, nil, fmt.Errorf("tuple %d: subject type and id are required", i)
		}
		if row.Relation == "" {
			return nil, nil, fmt.Errorf("tuple %d: relation is required", i)
		}
		out = append(out, &basev1.Tuple{
			Entity: &basev1.Entity{
				Type: row.Entity.Type,
				Id:   row.Entity.ID,
			},
			Relation: row.Relation,
			Subject: &basev1.Subject{
				Type:     row.Subject.Type,
				Id:       row.Subject.ID,
				Relation: row.Subject.Relation,
			},
		})
	}
	return out, meta, nil
}

func newDataReadCommand() *cobra.Command {
	var (
		credentialsPath string
		tenantID        string
		entityStr       string
		pageSize        uint32
	)

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read relationships for an entity (Data.ReadRelationships)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(tenantID) == "" {
				return fmt.Errorf("--tenant-id is required")
			}
			if strings.TrimSpace(entityStr) == "" {
				return fmt.Errorf("--entity is required")
			}
			ent, err := ParseEntityRef(entityStr)
			if err != nil {
				return fmt.Errorf("parse entity: %w", err)
			}
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

			client := basev1.NewDataClient(conn)
			return runDataReadRelationships(rpcCtx, os.Stdout, client, tenantID, ent, pageSize)
		},
	}

	fs := cmd.Flags()
	fs.StringVar(&credentialsPath, "credentials", "", "path to gRPC credentials file (default: $HOME/.permify/credentials)")
	fs.StringVar(&tenantID, "tenant-id", "", "tenant identifier (required)")
	fs.StringVar(&entityStr, "entity", "", "entity filter as type:id (e.g. document:1)")
	fs.Uint32Var(&pageSize, "page-size", 100, "maximum tuples to return (1–100)")
	_ = cmd.MarkFlagRequired("tenant-id")
	_ = cmd.MarkFlagRequired("entity")
	return cmd
}

type dataWriteClient interface {
	Write(ctx context.Context, in *basev1.DataWriteRequest, opts ...grpc.CallOption) (*basev1.DataWriteResponse, error)
}

type dataReadClient interface {
	ReadRelationships(ctx context.Context, in *basev1.RelationshipReadRequest, opts ...grpc.CallOption) (*basev1.RelationshipReadResponse, error)
}

func runDataWrite(
	ctx context.Context,
	w io.Writer,
	client dataWriteClient,
	tenantID string,
	meta *basev1.DataWriteRequestMetadata,
	tuples []*basev1.Tuple,
) error {
	resp, err := client.Write(ctx, &basev1.DataWriteRequest{
		TenantId: tenantID,
		Metadata: meta,
		Tuples:   tuples,
	})
	if err != nil {
		return GRPCStatusError(err)
	}
	_, _ = fmt.Fprintf(w, "Write succeeded.\nSnap token: %s\n", resp.GetSnapToken())
	return nil
}

func runDataReadRelationships(
	ctx context.Context,
	w io.Writer,
	client dataReadClient,
	tenantID string,
	entity *basev1.Entity,
	pageSize uint32,
) error {
	req := &basev1.RelationshipReadRequest{
		TenantId: tenantID,
		Metadata: &basev1.RelationshipReadRequestMetadata{},
		Filter: &basev1.TupleFilter{
			Entity: &basev1.EntityFilter{
				Type: entity.GetType(),
				Ids:  []string{entity.GetId()},
			},
		},
		PageSize: pageSize,
	}

	resp, err := client.ReadRelationships(ctx, req)
	if err != nil {
		return GRPCStatusError(err)
	}
	if len(resp.GetTuples()) == 0 {
		_, _ = fmt.Fprintln(w, "No relationships found.")
		return nil
	}
	_, _ = fmt.Fprintf(w, "Relationships (%d):\n", len(resp.GetTuples()))
	for _, t := range resp.GetTuples() {
		_, _ = fmt.Fprintln(w, " ", formatTupleLine(t))
	}
	if tok := resp.GetContinuousToken(); tok != "" {
		_, _ = fmt.Fprintf(w, "More results available (continuous_token present). Re-run with pagination when supported.\n")
	}
	return nil
}
