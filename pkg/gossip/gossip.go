package gossip

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"

	"github.com/Permify/permify/internal/config"
)

type IGossip interface {
	SyncMemberList() (nodes []string)
	Shutdown() error
}

type Engine struct {
	Enabled    bool
	memberList *memberlist.Memberlist
}

// InitMemberList initializes a memberlist instance with the provided seed nodes and config.
func InitMemberList(seedNodes []string, cfg config.Distributed) (*Engine, error) {
	conf := memberlist.DefaultLocalConfig()

	conf.Logger = log.New(io.Discard, "", 0)

	conf.BindAddr = "0.0.0.0"
	port, err := strconv.Atoi(cfg.AdvertisePort)
	if err != nil {
		return nil, fmt.Errorf("port convert error: %v", err)
	}
	conf.BindPort = port

	ip, err := ExternalIP()
	if err != nil {
		return nil, fmt.Errorf("external ip error: %v", err)
	}

	// config.AdvertiseAddr and config.AdvertisePort are used to tell other nodes how to reach this node.
	conf.AdvertiseAddr = ip
	conf.AdvertisePort = port

	list, err := memberlist.Create(conf)
	if err != nil {
		return nil, fmt.Errorf("memberlist Create Error %v", err)
	}

	if len(seedNodes) > 0 {
		_, err := list.Join(seedNodes)
		if err != nil {
			return nil, fmt.Errorf("starter ring join error: %v", err)
		}
	}

	return &Engine{
		Enabled:    true,
		memberList: list,
	}, nil
}

// SyncMemberList returns a list of all nodes in the cluster.
func (g *Engine) SyncMemberList() (nodes []string) {
	members := g.memberList.Members()
	for _, member := range members {
		nodes = append(nodes, member.Address())
	}

	return
}

// Shutdown gracefully shuts down the memberlist instance.
func (g *Engine) Shutdown() error {
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
