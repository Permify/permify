package gossip

import (
	"context"
	"errors"
	"fmt"
	"net"

	hash "github.com/Permify/permify/pkg/consistent"
)

// IGossip is an interface that represents the basic operations
// of a gossip-based membership protocol. Implementations of this
// interface should provide mechanisms for synchronizing cluster
// membership and managing the lifecycle of the gossip protocol.
type IGossip interface {
	SyncNodes(ctx context.Context, consistent *hash.ConsistentHash, nodeName, port string)
	// Shutdown gracefully stops the gossip protocol and performs
	// any necessary cleanup. It returns an error if the shutdown
	// process encounters any issues.
	Shutdown() error
}

// InitMemberList initializes a memberlist instance with the given
func InitMemberList(nodes, name string) (IGossip, error) {
	switch name {
	case "serf":
		return NewSerfGossip(nodes)
	default:
		return nil, fmt.Errorf("protocol not implamented: %s ", name)
	}
}

// ExternalIP returns the first non-loopback IPv4 address
func ExternalIP() (string, error) {
	// Get a list of network interfaces.
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// Iterate over the network interfaces.
	for _, iface := range interfaces {
		// Skip the interface if it's down or a loopback interface.
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Get a list of addresses associated with the interface.
		addresses, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		// Iterate over the addresses.
		for _, addr := range addresses {
			// Extract the IP address from the address.
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip the address if it's a loopback address or not IPv4.
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			// Return the IPv4 address as a string.
			return ip.String(), nil
		}
	}

	// Return an empty string if no external IPv4 address is found.
	return "", errors.New("network error")
}
