package balancer

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/cespare/xxhash/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"

	"github.com/Permify/permify/pkg/balancer"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var grpcServicePolicy = fmt.Sprintf(`{
		"loadBalancingPolicy": "%s"
	}`, balancer.Policy)

// Balancer is a wrapper around the balancer hash implementation that
type Balancer struct {
	schemaReader storage.SchemaReader
	checker      invoke.Check
	client       base.PermissionClient
	options      []grpc.DialOption
}

// NewCheckEngineWithBalancer
// struct with the provided cache.Cache instance.
func NewCheckEngineWithBalancer(
	checker invoke.Check,
	schemaReader storage.SchemaReader,
	dst *config.Distributed,
	srv *config.GRPC,
	authn *config.Authn,
) (invoke.Check, error) {
	var err error

	var options []grpc.DialOption

	var creds credentials.TransportCredentials

	if srv.TLSConfig.CertPath != "" && srv.TLSConfig.KeyPath != "" {
		creds, err = credentials.NewClientTLSFromFile(srv.TLSConfig.CertPath, srv.TLSConfig.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("could not load TLS certificate: %s", err)
		}
	} else {
		creds = insecure.NewCredentials()
	}

	// TODO: Add client-side authentication using a key from KeyAuthn.
	// 1. Initialize the KeyAuthn structure using the provided configuration.
	// 2. Convert the KeyAuthn instance into PerRPCCredentials.
	// 3. Append grpc.WithPerRPCCredentials() to the options slice.

	options = append(
		options,
		grpc.WithDefaultServiceConfig(grpcServicePolicy),
		grpc.WithTransportCredentials(creds),
	)

	conn, err := grpc.Dial(dst.Address, options...)
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
		slog.Error(err.Error())
		// If an error occurs while reading the entity definition, deny permission and return the error.
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	// Assume the request is not relational by default.
	isRelational := false

	// Check if the permission requested matches any reference attribute in the entity definition.
	tor, err := schema.GetTypeOfReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err == nil && tor != base.EntityDefinition_REFERENCE_ATTRIBUTE {
		isRelational = true
	}

	if err != nil {
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	// Create a new xxhash instance.
	h := xxhash.New()

	// Generate a unique key for the request based on its relational state.
	// This key helps in distributing the request.
	_, err = h.Write([]byte(engines.GenerateKey(request, isRelational)))
	if err != nil {
		slog.Error(err.Error())
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}
	k := hex.EncodeToString(h.Sum(nil))

	// Add a timeout of 2 seconds to the context and also set the generated key as a value.
	withTimeout, cancel := context.WithTimeout(context.WithValue(ctx, balancer.Key, k), 4*time.Second)
	defer cancel()

	// Logging the intention to forward the request to the underlying client.
	slog.Debug("Forwarding request with key to the underlying client", slog.String("key", k))

	// Perform the actual permission check by making a call to the underlying client.
	response, err := c.client.Check(withTimeout, request)
	if err != nil {
		// Log the error and return it.
		slog.Error(err.Error())
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
