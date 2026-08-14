package balancer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	grpcbalancer "google.golang.org/grpc/balancer"

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

// Balancer wraps permission checking with consistent-hash load balancing.
// All requests (single and batch) are distributed across nodes via the consistent
// hash picker so each entity lands on the node that caches it.
type Balancer struct {
	schemaReader storage.SchemaReader
	checker      invoke.Check
	client       base.PermissionClient
	builder      balancer.Builder
}

// NewCheckEngineWithBalancer creates a new check engine with a load balancer.
// It takes a Check interface, SchemaReader, distributed config, gRPC config, and authn config as input.
// It returns a Check interface and an error if any.
func NewCheckEngineWithBalancer(
	ctx context.Context,
	checker invoke.Check,
	schemaReader storage.SchemaReader,
	builder balancer.Builder,
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
			return nil, fmt.Errorf("could not load TLS certificate: %w", err)
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
		builder:      builder,
	}, nil
}

// Check distributes permission checks across cluster nodes via consistent hashing.
// Each entity ID is routed to the node determined by its hash key, ensuring cache locality.
// Entity IDs that hash to the same node are grouped into a single BulkCheck RPC.
func (c *Balancer) Check(ctx context.Context, request *invoke.BatchCheckRequest) (*invoke.BatchCheckResponse, error) {
	// Get the current picker; fall back to local if not ready.
	nodePicker := c.builder.Picker()
	if nodePicker == nil {
		return c.checker.Check(ctx, request)
	}

	deniedResp := invoke.NewBatchCheckResponse(base.CheckResult_CHECK_RESULT_DENIED, request.EntityIDs...)

	// Read entity definition once (shared across all entity IDs).
	en, _, err := c.schemaReader.ReadEntityDefinition(ctx, request.TenantID, request.EntityType, request.Metadata.GetSchemaVersion())
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return deniedResp, err
	}

	isRelational := engines.IsRelational(en, request.Permission)

	// Group entity IDs by target SubConn.
	groups := map[grpcbalancer.SubConn][]string{}
	for _, entityID := range request.EntityIDs {
		protoReq := request.ToPermissionCheckRequest(entityID)
		key := []byte(engines.GenerateKey(protoReq, isRelational))

		sc, err := nodePicker.Pick(key)
		if err != nil {
			slog.ErrorContext(ctx, "Pick failed, falling back to local", "error", err.Error())
			return c.checker.Check(ctx, request)
		}
		groups[sc] = append(groups[sc], entityID)
	}

	// Fan out: one RPC per node group, concurrently.
	type groupResult struct {
		resp *invoke.BatchCheckResponse
		err  error
	}

	results := make([]groupResult, len(groups))
	var wg sync.WaitGroup
	i := 0
	for sc, entityIDs := range groups {
		wg.Add(1)
		go func(idx int, sc grpcbalancer.SubConn, entityIDs []string) {
			defer wg.Done()
			routeCtx := context.WithValue(ctx, balancer.SubConnKey, sc)
			withTimeout, cancel := context.WithTimeout(routeCtx, 4*time.Second)
			defer cancel()

			if len(entityIDs) == 1 {
				// Single entity: use Check RPC.
				protoReq := request.ToPermissionCheckRequest(entityIDs[0])
				response, err := c.client.Check(withTimeout, protoReq)
				if err != nil {
					results[idx] = groupResult{err: err}
					return
				}
				results[idx] = groupResult{resp: &invoke.BatchCheckResponse{
					Results:  map[string]base.CheckResult{entityIDs[0]: response.GetCan()},
					Metadata: response.GetMetadata(),
				}}
			} else {
				// Multiple entities for same node: use BulkCheck RPC.
				items := make([]*base.PermissionBulkCheckRequestItem, len(entityIDs))
				for j, entityID := range entityIDs {
					items[j] = &base.PermissionBulkCheckRequestItem{
						Entity:     &base.Entity{Type: request.EntityType, Id: entityID},
						Permission: request.Permission,
						Subject:    request.Subject,
					}
				}
				bulkReq := &base.PermissionBulkCheckRequest{
					TenantId:  request.TenantID,
					Metadata:  request.Metadata,
					Items:     items,
					Context:   request.Context,
					Arguments: request.Arguments,
				}
				bulkResp, err := c.client.BulkCheck(withTimeout, bulkReq)
				if err != nil {
					results[idx] = groupResult{err: err}
					return
				}
				// Map BulkCheck results back by entity ID (results are ordered same as items).
				resp := &invoke.BatchCheckResponse{
					Results:  make(map[string]base.CheckResult, len(entityIDs)),
					Metadata: &base.PermissionCheckResponseMetadata{},
				}
				for j, entityID := range entityIDs {
					if j < len(bulkResp.GetResults()) {
						resp.Results[entityID] = bulkResp.GetResults()[j].GetCan()
						resp.Metadata.CheckCount += bulkResp.GetResults()[j].GetMetadata().GetCheckCount()
					} else {
						resp.Results[entityID] = base.CheckResult_CHECK_RESULT_DENIED
					}
				}
				results[idx] = groupResult{resp: resp}
			}
		}(i, sc, entityIDs)
		i++
	}
	wg.Wait()

	// Merge results from all node groups.
	merged := &invoke.BatchCheckResponse{
		Results:  make(map[string]base.CheckResult, len(request.EntityIDs)),
		Metadata: &base.PermissionCheckResponseMetadata{},
	}
	for _, r := range results {
		if r.err != nil {
			return deniedResp, r.err
		}
		if r.resp != nil {
			for entityID, result := range r.resp.Results {
				merged.Results[entityID] = result
			}
			merged.Metadata.CheckCount += r.resp.Metadata.GetCheckCount()
		}
	}

	// Fill in any missing entity IDs as DENIED.
	for _, entityID := range request.EntityIDs {
		if _, ok := merged.Results[entityID]; !ok {
			merged.Results[entityID] = base.CheckResult_CHECK_RESULT_DENIED
		}
	}

	return merged, nil
}
