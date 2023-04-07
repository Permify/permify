package postgres

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockPQDatabase simulates the behavior of a real PQDatabase.
type MockPQDatabase struct {
	mock.Mock
}

func (m *MockPQDatabase) GetEngineType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPQDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPQDatabase) IsReady(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}
