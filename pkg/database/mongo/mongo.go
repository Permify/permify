package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo -
type Mongo struct {
	maxPoolSize int
	connTimeout time.Duration
	database    string

	client *mongo.Client
}

// New -
func New(uri string, database string, opts ...Option) (*Mongo, error) {
	mn := &Mongo{
		maxPoolSize: _defaultMaxPoolSize,
		connTimeout: _defaultConnTimeout,
	}

	mn.database = database

	// Custom options
	for _, opt := range opts {
		opt(mn)
	}

	var err error
	mn.client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri).SetMaxPoolSize(uint64(mn.maxPoolSize)))
	if err != nil {
		return nil, err
	}

	_, err = mn.IsReady(context.Background())
	if err != nil {
		return nil, err
	}

	return mn, nil
}

// Database -
func (m *Mongo) Database() *mongo.Database {
	return m.client.Database(m.database)
}

// IsReady -
func (m *Mongo) IsReady(ctx context.Context) (bool, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, m.connTimeout)
	defer cancel()

	if err := m.client.Ping(ctx, readpref.Primary()); err != nil {
		return false, err
	}

	return true, nil
}

// GetConnectionType -
func (m *Mongo) GetConnectionType() string {
	return "mongo"
}

// Close -.
func (m *Mongo) Close() {
	if m.client != nil {
		_ = m.client.Disconnect(context.TODO())
	}
}
