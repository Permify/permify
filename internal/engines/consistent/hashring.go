package consistent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/Permify/permify/internal/invoke"
	hash "github.com/Permify/permify/pkg/consistent"
	"github.com/Permify/permify/pkg/gossip"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// Hashring is a wrapper around the consistent hash implementation that
type Hashring struct {
	checker           invoke.Check
	gossip            gossip.IGossip
	consistent        hash.Consistent
	localNodeAddress  string
	connectionOptions []grpc.DialOption
	l                 *logger.Logger
}

// NewCheckEngineWithHashring creates a new instance of EngineKeyManager by initializing an EngineKeys
// struct with the provided cache.Cache instance.
func NewCheckEngineWithHashring(checker invoke.Check, consistent *hash.ConsistentHash, g gossip.IGossip, port string, l *logger.Logger, options ...grpc.DialOption) (invoke.Check, error) {
	// Return a new instance of EngineKeys with the provided cache
	ip, err := gossip.ExternalIP()
	if err != nil {
		return nil, err
	}

	return &Hashring{
		checker:           checker,
		localNodeAddress:  ip + ":" + port,
		gossip:            g,
		consistent:        consistent,
		connectionOptions: options,
		l:                 l,
	}, nil
}

func (c *Hashring) Check(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	// Generate a unique checkKey string based on the provided PermissionCheckRequest
	k := fmt.Sprintf("check_%s_%s:%s:%s@%s", request.GetTenantId(), request.GetMetadata().GetSchemaVersion(), request.GetMetadata().GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   request.GetEntity(),
		Relation: request.GetPermission(),
	}), tuple.SubjectToString(request.GetSubject()))

	_, ok := c.consistent.Get(k)
	if !ok {
		ok := c.consistent.AddKey(k)
		if !ok {
			// If there's an error, return false
			return &base.PermissionCheckResponse{
				Can: base.CheckResult_RESULT_DENIED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}, errors.New("error adding key to consistent hash")
		}
		c.l.Info("added key %s to consistent hash", k)
	}
	node, found := c.consistent.Get(k)
	if !found {
		// If the responsible node is not found, return false
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, errors.New("node not found to key")
	}

	if node == c.localNodeAddress {
		// Get the value from the cache using the generated cache key
		return c.checker.Check(ctx, request)
	}

	resp, err := c.forwardRequestToNode(ctx, node, request)
	if err != nil {
		return &base.PermissionCheckResponse{
			Can: base.CheckResult_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	return resp, nil
}

// forwardRequestGetToNode forwards a request to the responsible node
func (c *Hashring) forwardRequestToNode(ctx context.Context, node string, request *base.PermissionCheckRequest) (*base.PermissionCheckResponse, error) {
	// Set up a connection to the server.
	conn, err := grpc.DialContext(ctx, node, c.connectionOptions...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a PermissionClient using the connection.
	client := base.NewPermissionClient(conn)

	// Prepare a context with a timeout.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	return client.Check(ctx, request)
}
