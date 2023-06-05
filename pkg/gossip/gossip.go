package gossip

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/hashicorp/memberlist"
)

// IGossip is an interface that represents the basic operations
// of a gossip-based membership protocol. Implementations of this
// interface should provide mechanisms for synchronizing cluster
// membership and managing the lifecycle of the gossip protocol.
type IGossip interface {
	// SyncMemberList retrieves the current list of nodes in the
	// gossip cluster and returns them as a slice of strings.
	SyncMemberList() (nodes []string)

	// Shutdown gracefully stops the gossip protocol and performs
	// any necessary cleanup. It returns an error if the shutdown
	// process encounters any issues.
	Shutdown() error
}

// Gossip is a struct that represents a gossip-based membership
// protocol implementation using the Memberlist library. It contains
// the following fields:
//   - Enabled: a boolean that indicates if the gossip protocol is enabled.
//   - memberList: a pointer to the memberlist.Memberlist instance, which
//     manages the cluster membership and communication between nodes.
type Gossip struct {
	Enabled    bool
	memberList *memberlist.Memberlist
}

// InitMemberList initializes a memberlist instance with the provided seed nodes and config.
func InitMemberList(nodes []string, grpcPort int) (*Gossip, error) {
	// Set up the default configuration for the memberlist.
	conf := memberlist.DefaultLocalConfig()

	// Discard logging to avoid cluttering the console.
	conf.Logger = log.New(io.Discard, "", 0)

	// Get the external IP address of the local machine.
	ip, err := ExternalIP()
	if err != nil {
		return nil, fmt.Errorf("external ip error: %w", err)
	}

	// Set the IP and port that the memberlist will advertise to other nodes.
	conf.AdvertiseAddr = ip
	conf.AdvertisePort = grpcPort

	// Create a new memberlist instance with the provided configuration.
	list, err := memberlist.Create(conf)
	if err != nil {
		return nil, fmt.Errorf("memberlist Create Error %w", err)
	}

	// If seed nodes are provided, attempt to join them.
	if len(nodes) > 0 {
		_, err := list.Join(nodes)
		if err != nil {
			return nil, fmt.Errorf("starter ring join error: %w", err)
		}
	}

	// Return a new Gossip instance with the initialized memberlist.
	return &Gossip{
		Enabled:    true,
		memberList: list,
	}, nil
}

// SyncMemberList returns a list of all nodes in the cluster.
func (g *Gossip) SyncMemberList() (nodes []string) {
	members := g.memberList.Members()
	for _, member := range members {
		nodes = append(nodes, member.Address())
	}

	return
}

// Shutdown gracefully shuts down the memberlist instance.
func (g *Gossip) Shutdown() error {
	return errors.Join(g.memberList.Leave(time.Second), g.memberList.Shutdown())
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
