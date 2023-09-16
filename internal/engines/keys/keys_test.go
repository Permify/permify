package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Permify/permify/pkg/cache/ristretto"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestEngineKeys_SetCheckKey(t *testing.T) {
	// Initialize a new Ristretto cache with a capacity of 10 keys
	cache, err := ristretto.New()
	assert.Nil(t, err)

	l := logger.New("debug")

	// Initialize a new EngineKeys struct with a new cache.Cache instance
	engineKeys := CheckEngineWithKeys{nil, nil, cache, l}

	// Create a new PermissionCheckRequest and PermissionCheckResponse
	checkReq := &base.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: "test_version",
			SnapToken:     "test_snap_token",
			Depth:         20,
		},
		Entity: &base.Entity{
			Type: "test-entity",
			Id:   "e1",
		},
		Permission: "test-permission",
		Subject: &base.Subject{
			Type: "user",
			Id:   "u1",
		},
	}

	checkResp := &base.PermissionCheckResponse{
		Can: base.CheckResult_CHECK_RESULT_ALLOWED,
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: 0,
		},
	}

	// Set the value for the given key in the cache
	success := engineKeys.setCheckKey(checkReq, checkResp, true)

	cache.Wait()

	// Check that the operation was successful
	assert.True(t, success)

	// Retrieve the value for the given key from the cache
	resp, found := engineKeys.getCheckKey(checkReq, true)

	// Check that the key was found and the retrieved value is the same as the original value
	assert.True(t, found)
	assert.Equal(t, checkResp, resp)
}

func TestEngineKeys_SetCheckKey_WithHashError(t *testing.T) {
	// Initialize a new Ristretto cache with a capacity of 10 keys
	cache, err := ristretto.New()
	assert.Nil(t, err)

	l := logger.New("debug")

	// Initialize a new EngineKeys struct with a new cache.Cache instance
	engineKeys := CheckEngineWithKeys{nil, nil, cache, l}

	// Create a new PermissionCheckRequest and PermissionCheckResponse
	checkReq := &base.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: "test_version",
			SnapToken:     "test_snap_token",
			Depth:         20,
		},
		Entity: &base.Entity{
			Type: "test-entity",
			Id:   "e1",
		},
		Permission: "test-permission",
		Subject: &base.Subject{
			Type: "user",
			Id:   "u1",
		},
	}

	checkResp := &base.PermissionCheckResponse{
		Can: base.CheckResult_CHECK_RESULT_ALLOWED,
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: 0,
		},
	}

	// Force an error while writing the key to the hash object by passing a nil key
	success := engineKeys.setCheckKey(nil, checkResp, true)

	cache.Wait()

	// Check that the operation was unsuccessful
	assert.False(t, success)

	// Retrieve the value for the given key from the cache
	resp, found := engineKeys.getCheckKey(checkReq, true)

	// Check that the key was not found
	assert.False(t, found)
	assert.Nil(t, resp)
}

func TestEngineKeys_GetCheckKey_KeyNotFound(t *testing.T) {
	// Initialize a new Ristretto cache with a capacity of 10 keys
	cache, err := ristretto.New()
	assert.Nil(t, err)

	l := logger.New("debug")

	// Initialize a new EngineKeys struct with a new cache.Cache instance
	engineKeys := CheckEngineWithKeys{nil, nil, cache, l}

	// Create a new PermissionCheckRequest
	checkReq := &base.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: "test_version",
			SnapToken:     "test_snap_token",
			Depth:         20,
		},
		Entity: &base.Entity{
			Type: "test-entity",
			Id:   "e1",
		},
		Permission: "test-permission",
		Subject: &base.Subject{
			Type: "user",
			Id:   "u1",
		},
	}

	// Retrieve the value for a non-existent key from the cache
	resp, found := engineKeys.getCheckKey(checkReq, true)

	// Check that the key was not found
	assert.False(t, found)
	assert.Nil(t, resp)
}

func TestEngineKeys_SetAndGetMultipleKeys(t *testing.T) {
	// Initialize a new Ristretto cache with a capacity of 10 keys
	cache, err := ristretto.New()
	assert.Nil(t, err)

	l := logger.New("debug")

	// Initialize a new EngineKeys struct with a new cache.Cache instance
	engineKeys := CheckEngineWithKeys{nil, nil, cache, l}

	// Create some new PermissionCheckRequests and PermissionCheckResponses
	checkReq1 := &base.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: "test_version",
			SnapToken:     "test_snap_token",
			Depth:         20,
		},
		Entity: &base.Entity{
			Type: "test-entity",
			Id:   "e1",
		},
		Permission: "test-permission",
		Subject: &base.Subject{
			Type: "user",
			Id:   "u1",
		},
	}
	checkResp1 := &base.PermissionCheckResponse{
		Can: base.CheckResult_CHECK_RESULT_ALLOWED,
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: 0,
		},
	}

	checkReq2 := &base.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: "test_version",
			SnapToken:     "test_snap_token",
			Depth:         20,
		},
		Entity: &base.Entity{
			Type: "test-entity",
			Id:   "e2",
		},
		Permission: "test-permission",
		Subject: &base.Subject{
			Type: "user",
			Id:   "u1",
		},
	}
	checkResp2 := &base.PermissionCheckResponse{
		Can: base.CheckResult_CHECK_RESULT_DENIED,
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: 0,
		},
	}

	checkReq3 := &base.PermissionCheckRequest{
		TenantId: "t2",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: "test_version",
			SnapToken:     "test_snap_token",
			Depth:         20,
		},
		Entity: &base.Entity{
			Type: "test-entity",
			Id:   "e1",
		},
		Permission: "test-permission",
		Subject: &base.Subject{
			Type: "user",
			Id:   "u2",
		},
	}
	checkResp3 := &base.PermissionCheckResponse{
		Can: base.CheckResult_CHECK_RESULT_DENIED,
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: 0,
		},
	}

	// Set the values for the given keys in the cache
	success1 := engineKeys.setCheckKey(checkReq1, checkResp1, true)
	success2 := engineKeys.setCheckKey(checkReq2, checkResp2, true)
	success3 := engineKeys.setCheckKey(checkReq3, checkResp3, true)

	cache.Wait()

	// Check that all the operations were successful
	assert.True(t, success1)
	assert.True(t, success2)
	assert.True(t, success3)

	// Retrieve the value for the given key from the cache
	resp1, found1 := engineKeys.getCheckKey(checkReq1, true)
	resp2, found2 := engineKeys.getCheckKey(checkReq2, true)
	resp3, found3 := engineKeys.getCheckKey(checkReq3, true)

	// Check that the key was not found
	assert.True(t, found1)
	assert.Equal(t, checkResp1, resp1)

	assert.True(t, found2)
	assert.Equal(t, checkResp2, resp2)

	assert.True(t, found3)
	assert.Equal(t, checkResp3, resp3)
}

func TestEngineKeys_SetCheckKeyWithArguments(t *testing.T) {
	// Initialize a new Ristretto cache with a capacity of 10 keys
	cache, err := ristretto.New()
	assert.Nil(t, err)

	l := logger.New("debug")

	// Initialize a new EngineKeys struct with a new cache.Cache instance
	engineKeys := CheckEngineWithKeys{nil, nil, cache, l}

	// Create a new PermissionCheckRequest and PermissionCheckResponse
	checkReq := &base.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: "test_version",
			SnapToken:     "test_snap_token",
			Depth:         20,
		},
		Arguments: []*base.Argument{
			{
				Type: &base.Argument_ComputedAttribute{
					ComputedAttribute: &base.ComputedAttribute{
						Name: "test_argument_1",
					},
				},
			},
			{
				Type: &base.Argument_ComputedAttribute{
					ComputedAttribute: &base.ComputedAttribute{
						Name: "test_argument_2",
					},
				},
			},
		},
		Entity: &base.Entity{
			Type: "test-entity",
			Id:   "e1",
		},
		Permission: "test-rule",
		Subject: &base.Subject{
			Type: "user",
			Id:   "u1",
		},
	}

	checkResp := &base.PermissionCheckResponse{
		Can: base.CheckResult_CHECK_RESULT_ALLOWED,
		Metadata: &base.PermissionCheckResponseMetadata{
			CheckCount: 0,
		},
	}

	// Set the value for the given key in the cache
	success := engineKeys.setCheckKey(checkReq, checkResp, true)

	cache.Wait()

	// Check that the operation was successful
	assert.True(t, success)

	// Retrieve the value for the given key from the cache
	resp, found := engineKeys.getCheckKey(checkReq, true)

	// Check that the key was found and the retrieved value is the same as the original value
	assert.True(t, found)
	assert.Equal(t, checkResp, resp)
}

func TestEngineKeys_SetCheckKeyWithContext(t *testing.T) {
	value, err := anypb.New(&base.BooleanValue{Data: true})
	if err != nil {
	}

	data, err := structpb.NewStruct(map[string]interface{}{
		"day_of_a_week": "saturday",
		"day_of_a_year": 356,
	})
	if err != nil {
	}

	// Create a new PermissionCheckRequest and PermissionCheckResponse
	checkReq := &base.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: "test_version",
			SnapToken:     "test_snap_token",
			Depth:         20,
		},
		Arguments: []*base.Argument{
			{
				Type: &base.Argument_ComputedAttribute{
					ComputedAttribute: &base.ComputedAttribute{
						Name: "test_argument_1",
					},
				},
			},
			{
				Type: &base.Argument_ComputedAttribute{
					ComputedAttribute: &base.ComputedAttribute{
						Name: "test_argument_2",
					},
				},
			},
		},
		Context: &base.Context{
			Tuples: []*base.Tuple{
				{
					Entity: &base.Entity{
						Type: "entity_type",
						Id:   "entity_id",
					},
					Relation: "relation",
					Subject: &base.Subject{
						Type: "subject_type",
						Id:   "subject_id",
					},
				},
			},
			Attributes: []*base.Attribute{
				{
					Entity: &base.Entity{
						Type: "entity_type",
						Id:   "entity_id",
					},
					Attribute: "is_public",
					Value:     value,
				},
			},
			Data: data,
		},
		Entity: &base.Entity{
			Type: "test-entity",
			Id:   "e1",
		},
		Permission: "test-rule",
		Subject: &base.Subject{
			Type: "user",
			Id:   "u1",
		},
	}

	assert.Equal(t, "check|t1|test_version|test_snap_token|entity_type:entity_id#relation@subject_type:subject_id,entity_type:entity_id$is_public|boolean:true,day_of_a_week:saturday,day_of_a_year:35|test-entity:e1$test-rule(test_argument_1,test_argument_2)", GenerateKey(checkReq, false))
}
