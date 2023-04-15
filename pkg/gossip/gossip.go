package gossip

import (
	"errors"
	"fmt"
	"github.com/Permify/permify/internal/config"
	"github.com/hashicorp/memberlist"
	"strconv"
	"time"
)

type Engine struct {
	Enabled    bool
	memberList *memberlist.Memberlist
}

func InitMemberList(seedNodes []string, cfg config.Server) (*Engine, error) {
	config := memberlist.DefaultLocalConfig()

	config.BindAddr = cfg.Address
	port, err := strconv.Atoi(cfg.HTTP.Port)
	if err != nil {
		return nil, fmt.Errorf("port convert error: %v", err)
	}
	config.BindPort = port

	// config.AdvertiseAddr and config.AdvertisePort are used to tell other nodes how to reach this node.
	config.AdvertiseAddr = cfg.Address
	config.AdvertisePort = 3476

	list, err := memberlist.Create(config)
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

func (g *Engine) SyncMemberList(list *memberlist.Memberlist) {

	members := list.Members()
	for _, member := range members {
		fmt.Printf("Düğüm: %s, IP: %s, Port: %d\n", member.Name, member.Addr, member.Port)
	}
}

func (g *Engine) Shutdown() error {
	return errors.Join(g.memberList.Leave(time.Second), g.memberList.Shutdown())
}
