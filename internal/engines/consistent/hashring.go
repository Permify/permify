package consistent

import (
	"context"
	"errors"
	"fmt"

	v1 "github.com/Permify/permify-go/generated/base/v1"
	client "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Permify/permify/internal/invoke"
	hash "github.com/Permify/permify/pkg/consistent"
	"github.com/Permify/permify/pkg/gossip"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// Hashring is a wrapper around the consistent hash implementation that
type Hashring struct {
	checker          invoke.Check
	gossip           gossip.IGossip
	consistent       hash.Consistent
	localNodeAddress string
	l                *logger.Logger
}

// NewCheckEngineWithHashring creates a new instance of EngineKeyManager by initializing an EngineKeys
// struct with the provided cache.Cache instance.
func NewCheckEngineWithHashring(checker invoke.Check, consistent *hash.ConsistentHash, g *gossip.Engine, port string, l *logger.Logger) (invoke.Check, error) {
	// Return a new instance of EngineKeys with the provided cache

	ip, err := gossip.ExternalIP()
	if err != nil {
		return nil, err
	}

	return &Hashring{
		checker:          checker,
		localNodeAddress: ip + ":" + port,
		gossip:           g,
		consistent:       consistent,
		l:                l,
	}, nil
}

func (c *Hashring) Check(ctx context.Context, request *base.PermissionCheckRequest) (response *base.PermissionCheckResponse, err error) {
	// Generate a unique checkKey string based on the provided PermissionCheckRequest
	k := fmt.Sprintf("check_%s_%s:%s:%s@%s", request.GetTenantId(), request.GetMetadata().GetSchemaVersion(), request.GetMetadata().GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   request.GetEntity(),
		Relation: request.GetPermission(),
	}), tuple.SubjectToString(request.GetSubject()))

	ok := c.consistent.AddKey(k)
	if !ok {
		// If there's an error, return false
		return &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, errors.New("error adding key %s to consistent hash")
	}
	c.l.Info("added key %s to consistent hash", k)

	node, found := c.consistent.Get(k)
	if !found {
		// If the responsible node is not found, return false
		return &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, errors.New("node not found for key")
	}

	if node == c.localNodeAddress {
		// Get the value from the cache using the generated cache key
		return c.checker.Check(ctx, request)
	}

	resp, err := c.forwardRequestToNode(ctx, node, request)
	if err != nil {
		return &base.PermissionCheckResponse{
			Can: base.PermissionCheckResponse_RESULT_DENIED,
			Metadata: &base.PermissionCheckResponseMetadata{
				CheckCount: 0,
			},
		}, err
	}

	return resp, nil
}

// forwardRequestGetToNode forwards a request to the responsible node
func (c *Hashring) forwardRequestToNode(ctx context.Context, node string, request *base.PermissionCheckRequest) (*base.PermissionCheckResponse, error) {
	// generate new client
	p, err := client.NewClient(
		client.Config{
			Endpoint: node,
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	var res *v1.PermissionCheckResponse
	res, err = p.Permission.Check(ctx, &v1.PermissionCheckRequest{
		TenantId: request.GetTenantId(),
		Metadata: &v1.PermissionCheckRequestMetadata{
			SchemaVersion: request.GetMetadata().GetSchemaVersion(),
			SnapToken:     request.GetMetadata().GetSnapToken(),
			Exclusion:     request.GetMetadata().GetExclusion(),
			Depth:         request.GetMetadata().GetDepth(),
		},
		Entity: &v1.Entity{
			Type: request.GetEntity().GetType(),
			Id:   request.GetEntity().GetId(),
		},
		Permission: request.GetPermission(),
		Subject: &v1.Subject{
			Type:     request.GetSubject().GetType(),
			Id:       request.GetSubject().GetId(),
			Relation: request.GetSubject().GetRelation(),
		},
	})

	return &base.PermissionCheckResponse{
		Can: base.PermissionCheckResponse_Result(res.GetCan()),
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: res.GetMetadata().GetCheckCount(),
		},
	}, err
}
