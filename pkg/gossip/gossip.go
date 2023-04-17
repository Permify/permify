package gossip

import (
	"errors"
	"fmt"
	"github.com/Permify/permify/internal/config"
	"github.com/hashicorp/memberlist"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type Engine struct {
	Enabled    bool
	memberList *memberlist.Memberlist
}

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

func (g *Engine) SyncMemberList() (nodes []string) {
	members := g.memberList.Members()
	for _, member := range members {
		nodes = append(nodes, member.Address())
	}

	return
}

func (g *Engine) Shutdown() error {
	return errors.Join(g.memberList.Leave(time.Second), g.memberList.Shutdown())
}

func ExternalIP() (string, error) {
	faces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range faces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		address, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range address {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
