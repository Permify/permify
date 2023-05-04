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

type IGossip interface {
	SyncMemberList() (nodes []string)
	Shutdown() error
}

type Gossip struct {
	Enabled    bool
	memberList *memberlist.Memberlist
}

// InitMemberList initializes a memberlist instance with the provided seed nodes and config.
func InitMemberList(nodes []string, grpcPort int) (*Gossip, error) {
	conf := memberlist.DefaultLocalConfig()

	conf.Logger = log.New(io.Discard, "", 0)

	//conf.BindAddr = "0.0.0.0"
	//conf.BindPort = gossipPort

	ip, err := ExternalIP()
	if err != nil {
		return nil, fmt.Errorf("external ip error: %v", err)
	}

	conf.AdvertiseAddr = ip
	conf.AdvertisePort = grpcPort

	list, err := memberlist.Create(conf)
	if err != nil {
		return nil, fmt.Errorf("memberlist Create Error %v", err)
	}

	if len(nodes) > 0 {
		_, err := list.Join(nodes)
		if err != nil {
			return nil, fmt.Errorf("starter ring join error: %v", err)
		}
	}

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
