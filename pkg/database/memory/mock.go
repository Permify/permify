package memory

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockIMDatabase simulates the behavior of a real IMDatabase.
type MockIMDatabase struct {
	mock.Mock
}

func (m *MockIMDatabase) GetEngineType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIMDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockIMDatabase) IsReady(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}
