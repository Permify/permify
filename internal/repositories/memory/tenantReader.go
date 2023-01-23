package memory

import (
	"context"

	"github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TenantReader - Structure for Tenant Reader
type TenantReader struct {
	database *db.Memory
	// logger
	logger logger.Interface
}

// NewTenantReader creates a new TenantReader
func NewTenantReader(database *db.Memory, logger logger.Interface) *TenantReader {
	return &TenantReader{
		database: database,
		logger:   logger,
	}
}

// ListTenants -
func (r *TenantReader) ListTenants(ctx context.Context, pagination database.Pagination) (tenants []*base.Tenant, ct database.EncodedContinuousToken, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var lowerBound uint64 = 0
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, err
		}
		lowerBound = t.(utils.ContinuousToken).Value
	}

	var result memdb.ResultIterator
	result, err = txn.LowerBound(TenantsTable, "id", lowerBound)
	if err != nil {
		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	tenants = make([]*base.Tenant, 0, pagination.PageSize()+1)
	for obj := result.Next(); obj != nil; obj = result.Next() {
		t, ok := obj.(repositories.Tenant)
		if !ok {
			return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		tenants = append(tenants, t.ToTenant())
		if len(tenants) > int(pagination.PageSize()) {
			return tenants[:pagination.PageSize()], utils.NewContinuousToken(t.ID).Encode(), nil
		}
	}

	return tenants, utils.NewNoopContinuousToken().Encode(), err
}
