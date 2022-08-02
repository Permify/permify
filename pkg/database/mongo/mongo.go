package mongo

import (
	"context"
	"fmt"
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

	// helper.Pre(cr.String())
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo - NewMongo - mongo.Connect: %w", err)
	}

	mn.client = client

	return mn, nil
}

// Database -
func (m *Mongo) Database() *mongo.Database {
	return m.client.Database(m.database)
}

// IsReady -
func (m *Mongo) IsReady(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
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
