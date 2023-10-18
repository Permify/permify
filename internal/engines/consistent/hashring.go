package consistent

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"

	"github.com/Permify/permify/internal/engines/keys"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	hash "github.com/Permify/permify/pkg/consistent"
	"github.com/Permify/permify/pkg/gossip"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// Hashring is a wrapper around the consistent hash implementation that
type Hashring struct {
	// schemaReader is responsible for reading schema information
	schemaReader      storage.SchemaReader
	checker           invoke.Check
	gossip            gossip.IGossip
	consistent        hash.Consistent
	localNodeAddress  string
	connectionOptions []grpc.DialOption
}

// NewCheckEngineWithHashring creates a new instance of EngineKeyManager by initializing an EngineKeys
// struct with the provided cache.Cache instance.
func NewCheckEngineWithHashring(checker invoke.Check, schemaReader storage.SchemaReader, consistent *hash.ConsistentHash, g gossip.IGossip, port string) (invoke.Check, error) {
	// Return a new instance of EngineKeys with the provided cache
	ip, err := gossip.ExternalIP()
	if err != nil {
		return nil, err
	}

	return &Hashring{
		schemaReader:     schemaReader,
		checker:          checker,
		localNodeAddress: ip + ":" + port,
		gossip:           g,
		consistent:       consistent,
	}, nil
}

func (c *Hashring) Check(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	// Retrieve entity definition
	var en *base.EntityDefinition
	en, _, err = c.schemaReader.ReadEntityDefinition(ctx, request.GetTenantId(), request.GetEntity().GetType(), request.GetMetadata().GetSchemaVersion())
	if err != nil {
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	isRelational := false

	// Determine the type of the reference by name in the given entity definition.
	tor, err := schema.GetTypeOfReferenceByNameInEntityDefinition(en, request.GetPermission())
	if err == nil {
		if tor != base.EntityDefinition_REFERENCE_ATTRIBUTE {
			isRelational = true
		}
	}

	// Generate a unique checkKey string based on the provided PermissionCheckRequest
	k := keys.GenerateKey(request, isRelational)

	_, _, ok := c.consistent.Get(k)
	if !ok {
		// If there's an error, return false
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, errors.New("error adding key %s to consistent hash")
	}

	node, conn, found := c.consistent.Get(k)
	if !found {
		// If the responsible node is not found, return false
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, errors.New("node not found to key")
	}

	if node == c.localNodeAddress {
		// Get the value from the cache using the generated cache key
		return c.checker.Check(ctx, request)
	}

	resp, err := c.forwardRequestToNode(ctx, conn, request)
	if err != nil {
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_CHECK_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	return resp, nil
}

// forwardRequestGetToNode forwards a request to the responsible node
func (c *Hashring) forwardRequestToNode(ctx context.Context, conn *grpc.ClientConn, request *base.PermissionCheckRequest) (*base.PermissionCheckResponse, error) {
	// Create a PermissionClient using the connection.
	client := base.NewPermissionClient(conn)

	// Prepare a context with a timeout.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	return client.Check(ctx, request)
}
