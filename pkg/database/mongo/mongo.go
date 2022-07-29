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

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, fmt.Errorf("postgres - NewMongo - Ping: %w", err)
	}

	mn.client = client

	return mn, nil
}

// Database -
func (m *Mongo) Database() *mongo.Database {
	return m.client.Database(m.database)
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
