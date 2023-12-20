package memory

import (
	"context"

	"github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"

	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"
	"github.com/Permify/permify/internal/storage/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantReader - Structure for Tenant Reader
type TenantReader struct {
	database *db.Memory
}

// NewTenantReader creates a new TenantReader
func NewTenantReader(database *db.Memory) *TenantReader {
	return &TenantReader{
		database: database,
	}
}

// ListTenants -
func (r *TenantReader) ListTenants(_ context.Context, pagination database.Pagination) (tenants []*base.Tenant, ct database.EncodedContinuousToken, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var lowerBound string
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, err
		}
		lowerBound = t.(utils.ContinuousToken).Value
	}

	var result memdb.ResultIterator
	result, err = txn.LowerBound(constants.TenantsTable, "id", lowerBound)
	if err != nil {
		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	tenants = make([]*base.Tenant, 0, pagination.PageSize()+1)
	for obj := result.Next(); obj != nil; obj = result.Next() {
		t, ok := obj.(storage.Tenant)
		if !ok {
			return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		tenants = append(tenants, t.ToTenant())
		if len(tenants) > int(pagination.PageSize()) {
			return tenants[:pagination.PageSize()], utils.NewContinuousToken(t.ID).Encode(), nil
		}
	}

	return tenants, database.NewNoopContinuousToken().Encode(), err
}
