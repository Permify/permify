package keys_test

/*
import (
	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/pkg/cache/ristretto"
	hash "github.com/Permify/permify/pkg/consistent"
	"github.com/Permify/permify/pkg/gossip"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestEngineKeys_SetCheckKey(t *testing.T) {
	httpmock.Activate()

	defer httpmock.DeactivateAndReset()

	cache, err := ristretto.New()
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Server.Address = "172.0.0.1"
	cfg.Distributed.Enabled = true
	cfg.Distributed.SeedNodes = []string{"172.0.0.1"}
	cfg.Distributed.AdvertisePort = "3780"

	consistent := hash.NewConsistentHash(100, []string{"172.0.0.2:3780"}, nil)
	memberList, err := gossip.InitMemberList([]string{"172.0.0.1:3780", "172.0.0.2:3780"}, cfg.Distributed)
	if err != nil {
		t.Fatal(err)
	}

	keyManager := keys.NewCheckEngineKeys(cache, consistent, memberList, cfg.Server)

	// Create a new PermissionCheckRequest and PermissionCheckResponse
	checkReq := &base.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: "test_version",
			SnapToken:     "test_snap_token",
			Exclusion:     false,
			Depth:         20,
		},
		Entity: &base.Entity{
			Type: "test-entity",
			Id:   "e1",
		},
		Permission: "test-permission",
		Subject: &base.Subject{
			Type: tuple.USER,
			Id:   "u1",
		},
	}

	checkResp := &base.PermissionCheckResponse{
		Can: base.PermissionCheckResponse_RESULT_ALLOWED,
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: 0,
		},
	}

	// Set the value for the given key in the cache
	success := keyManager.SetCheckKey(checkReq, checkResp)

	cache.Wait()

	// Check that the operation was successful
	assert.True(t, success)

	// Retrieve the value for the given key from the cache
	resp, found := keyManager.GetCheckKey(checkReq)

	// Check that the key was found and the retrieved value is the same as the original value
	assert.True(t, found)
	assert.Equal(t, checkResp, resp)
}

func DockerTest() {
	containertest
}

*/
