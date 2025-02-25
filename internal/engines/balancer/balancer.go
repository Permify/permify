package balancer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/balancer"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Balancer is a wrapper around the balancer hash implementation that
type Balancer struct {
	schemaReader storage.SchemaReader
	checker      invoke.Check
	client       base.PermissionClient
	options      []grpc.DialOption
}

// NewCheckEngineWithBalancer creates a new check engine with a load balancer.
// It takes a Check interface, SchemaReader, distributed config, gRPC config, and authn config as input.
// It returns a Check interface and an error if any.
func NewCheckEngineWithBalancer(
	ctx context.Context,
	checker invoke.Check,
	schemaReader storage.SchemaReader,
	no string,
	dst *config.Distributed,
	srv *config.GRPC,
	authn *config.Authn,
) (invoke.Check, error) {
	var (
		creds    credentials.TransportCredentials
		options  []grpc.DialOption
		isSecure bool
		err      error
	)

	// Set up TLS credentials if paths are provided
	if srv.TLSConfig.Enabled && srv.TLSConfig.CertPath != "" {
		isSecure = true
		creds, err = credentials.NewClientTLSFromFile(srv.TLSConfig.CertPath, no)
		if err != nil {
			return nil, fmt.Errorf("could not load TLS certificate: %s", err)
		}
	} else {
		creds = insecure.NewCredentials()
	}

	bc := &balancer.Config{
		PartitionCount:    dst.PartitionCount,
		ReplicationFactor: dst.ReplicationFactor,
		Load:              dst.Load,
		PickerWidth:       dst.PickerWidth,
	}

	bcjson, err := bc.ServiceConfigJSON()
	if err != nil {
		return nil, err
	}

	// Append common options
	options = append(
		options,
		grpc.WithDefaultServiceConfig(bcjson),
		grpc.WithTransportCredentials(creds),
	)

	// Handle authentication if enabled
	if authn != nil && authn.Enabled {
		token, err := setupAuthn(ctx, authn)
		if err != nil {
			return nil, err
		}
		if isSecure {
			options = append(options, grpc.WithPerRPCCredentials(secureTokenCredentials{"authorization": "Bearer " + token}))
		} else {
			options = append(options, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{"authorization": "Bearer " + token}))
		}
	}

	conn, err := grpc.NewClient(dst.Address, options...)
	if err != nil {
		return nil, err
	}

	return &Balancer{
		schemaReader: schemaReader,
		checker:      checker,
		client:       base.NewPermissionClient(conn),
	}, nil
}

// Check performs a permission check using the schema reader to obtain
// entity definitions, then distributes the request based on a generated key.
func (c *Balancer) Check(ctx context.Context, request *base.PermissionCheckRequest) (*base.PermissionCheckResponse, error) {
	// Fetch the EntityDefinition for the given tenant, entity type, and schema version.
	en, _, err := c.schemaReader.ReadEntityDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		// If an error occurs while reading the entity definition, deny permission and return the error.
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	isRelational := engines.IsRelational(en, request.GetPermission())

	// Add a timeout of 2 seconds to the context and also set the generated key as a value.
	withTimeout, cancel := context.WithTimeout(context.WithValue(ctx, balancer.Key, []byte(engines.GenerateKey(request, isRelational))), 4*time.Second)
	defer cancel()

	// Logging the intention to forward the request to the underlying client.
	slog.InfoContext(ctx, "Forwarding request with key to the underlying client")

	// Perform the actual permission check by making a call to the underlying client.
	response, err := c.client.Check(withTimeout, request)
	if err != nil {
		// Log the error and return it.
		slog.ErrorContext(ctx, err.Error())
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	// Return the response received from the client.
	return response, nil
}
