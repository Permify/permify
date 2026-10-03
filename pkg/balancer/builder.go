package balancer

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/exp/slog"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"

	"github.com/Permify/permify/pkg/consistent"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Package-level constants for the balancer name and context keys.
const (
	Name       = "consistenthashing"   // Name of the balancer.
	SubConnKey = contextKey("subconn") // Context key for pre-computed SubConn.
)

// Config represents the configuration for the consistent hashing balancer.
type Config struct {
	serviceconfig.LoadBalancingConfig `json:"-"` // Embedding the base load balancing config.
	PartitionCount                    int        `json:"partitionCount,omitempty"`    // Number of partitions in the consistent hash ring.
	ReplicationFactor                 int        `json:"replicationFactor,omitempty"` // Number of replicas for each member.
	Load                              float64    `json:"load,omitempty"`              // Load factor for balancing traffic.
	PickerWidth                       int        `json:"pickerWidth,omitempty"`       // Number of closest members to consider in the picker.
}

// ServiceConfigJSON generates the JSON representation of the load balancer configuration.
func (c *Config) ServiceConfigJSON() (string, error) {
	// Define the JSON wrapper structure for the load balancing config.
	type Wrapper struct {
		LoadBalancingConfig []map[string]*Config `json:"loadBalancingConfig"`
	}

	// Apply default values for zero fields.
	if c.PartitionCount <= 0 {
		c.PartitionCount = consistent.DefaultPartitionCount
	}
	if c.ReplicationFactor <= 0 {
		c.ReplicationFactor = consistent.DefaultReplicationFactor
	}
	if c.Load <= 1.0 {
		c.Load = consistent.DefaultLoad
	}
	if c.PickerWidth < 1 {
		c.PickerWidth = consistent.DefaultPickerWidth
	}

	// Create the wrapper with the current configuration.
	wrapper := Wrapper{
		LoadBalancingConfig: []map[string]*Config{
			{Name: c},
		},
	}

	// Marshal the wrapped configuration to JSON.
	jsonData, err := json.Marshal(wrapper)
	if err != nil {
		return "", fmt.Errorf("failed to marshal service config: %w", err)
	}

	return string(jsonData), nil
}

var defaultBuilder atomic.Pointer[builder]

// NewBuilder initializes a new builder with the given hashing function.
// The builder is also stored as the package default, accessible via GetBuilder.
func NewBuilder(fn consistent.Hasher) Builder {
	b := &builder{hasher: fn}
	defaultBuilder.Store(b)
	return b
}

// GetBuilder returns the builder created by NewBuilder, or nil.
func GetBuilder() Builder {
	return defaultBuilder.Load()
}

// ConsistentMember represents a member in the consistent hashing ring.
type ConsistentMember struct {
	balancer.SubConn        // Embedded SubConn for the gRPC connection.
	name             string // Unique identifier for the member.
}

// String returns the name of the ConsistentMember.
func (s ConsistentMember) String() string { return s.name }

// builder is responsible for creating and configuring the consistent hashing balancer.
type builder struct {
	sync.Mutex                          // Mutex for thread-safe updates to config.
	hasher     consistent.Hasher        // Hashing function for the consistent hash ring.
	config     Config                   // Current balancer configuration.
	bal        atomic.Pointer[Balancer] // Reference to the active gRPC balancer (lock-free).
}

// Builder defines the interface for the consistent hashing balancer builder.
type Builder interface {
	balancer.Builder      // Interface for building balancers.
	balancer.ConfigParser // Interface for parsing balancer configurations.

	// Picker returns the current NodePicker for routing keys to SubConns.
	// Returns nil if the balancer is not ready.
	Picker() NodePicker
}

// Name returns the name of the balancer.
func (b *builder) Name() string { return Name }

// Build creates a new instance of the consistent hashing balancer.
func (b *builder) Build(cc balancer.ClientConn, _ balancer.BuildOptions) balancer.Balancer {
	// Initialize a new balancer with default values.
	// picker starts as nil (no SubConns yet); gRPC gets an errPicker via UpdateState below.
	bal := &Balancer{
		clientConn:            cc,
		addressSubConns:       resolver.NewAddressMap(),
		subConnStates:         make(map[balancer.SubConn]connectivity.State),
		connectivityEvaluator: &balancer.ConnectivityStateEvaluator{},
		state:                 connectivity.Connecting,
		hasher:                b.hasher,
	}

	b.bal.Store(bal)

	return bal
}

// Picker returns the current NodePicker. Nil if the balancer is not ready.
func (b *builder) Picker() NodePicker {
	if bal := b.bal.Load(); bal != nil {
		return bal.picker.Load()
	}
	return nil
}

// ParseConfig parses the balancer configuration from the provided JSON.
func (b *builder) ParseConfig(rm json.RawMessage) (serviceconfig.LoadBalancingConfig, error) {
	var cfg Config
	// Unmarshal the JSON configuration into the Config struct.
	if err := json.Unmarshal(rm, &cfg); err != nil {
		return nil, fmt.Errorf("consistenthash: unable to unmarshal LB policy config: %s, error: %w", string(rm), err)
	}

	// Log the parsed configuration using structured logging.
	slog.Info("Parsed balancer configuration",
		slog.String("raw_json", string(rm)), // Log the raw JSON string.
		slog.Any("config", cfg),             // Log the unmarshaled Config struct.
	)

	// Set default values for configuration if not provided.
	if cfg.PartitionCount <= 0 {
		cfg.PartitionCount = consistent.DefaultPartitionCount
	}
	if cfg.ReplicationFactor <= 0 {
		cfg.ReplicationFactor = consistent.DefaultReplicationFactor
	}
	if cfg.Load <= 1.0 {
		cfg.Load = consistent.DefaultLoad
	}
	if cfg.PickerWidth < 1 {
		cfg.PickerWidth = consistent.DefaultPickerWidth
	}

	// Update the builder's configuration with thread safety.
	b.Lock()
	b.config = cfg
	b.Unlock()

	return &cfg, nil
}
