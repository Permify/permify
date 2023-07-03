package gossip

import (
	"errors"
	"fmt"
	hash "github.com/Permify/permify/pkg/consistent"
	"github.com/hashicorp/serf/serf"
	"io"
	"log"
)

type Serf struct {
	Enabled bool
	serf    *serf.Serf
	EventCh chan serf.Event
}

func NewSerfGossip(node string) (*Serf, error) {
	config := serf.DefaultConfig()
	config.RejoinAfterLeave = true
	eventChannel := make(chan serf.Event, 256)
	// disable logger for serf
	config.LogOutput = io.Discard
	config.Logger = log.New(io.Discard, "", 0)
	config.MemberlistConfig.LogOutput = io.Discard
	config.EventCh = eventChannel

	// Create serf instance
	s, err := serf.Create(config)
	if err != nil {
		log.Fatalf("Failed to create serf instance: %v", err)
		return nil, err
	}

	_, err = s.Join([]string{node}, true)
	if err != nil {
		log.Fatalf("Failed to join cluster: %v", err)
		return nil, err
	}

	// Return a new Gossip instance with the initialized memberlist.
	return &Serf{
		Enabled: true,
		serf:    s,
		EventCh: eventChannel,
	}, nil
}

func (s *Serf) SyncNodes(consistent *hash.ConsistentHash, nodeName, port string) {
	for {
		select {
		case e := <-s.EventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				me := e.(serf.MemberEvent)
				for _, m := range me.Members {
					if m.Name != nodeName {
						if _, exists := consistent.Nodes[fmt.Sprintf("%s:%d", m.Addr.String(), m.Port)]; !exists {
							fmt.Printf("Adding node %s:%s to the consistent hash\n", m.Addr.String(), port)
							consistent.AddWithWeight(fmt.Sprintf("%s:%s", m.Addr.String(), port), 100)
						}
					}
				}
			case serf.EventMemberLeave, serf.EventMemberFailed, serf.EventMemberReap:
				me := e.(serf.MemberEvent)
				for _, m := range me.Members {
					if m.Name != nodeName {
						fmt.Printf("Removing node %s:%d to the consistent hash\n", m.Addr.String(), m.Port)
						if _, exists := consistent.Nodes[fmt.Sprintf("%s:%d", m.Addr.String(), m.Port)]; exists {
							consistent.Remove(fmt.Sprintf("%s:%s", m.Addr.String(), port))
						}
					}
				}
			}
		}
	}
}

func (s *Serf) Shutdown() error {
	return errors.Join(s.serf.Leave(), s.serf.Shutdown())
}
