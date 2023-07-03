package gossip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/serf/serf"

	hash "github.com/Permify/permify/pkg/consistent"
)

// Serf structure with the serf instance and EventCh channel
type Serf struct {
	Enabled bool
	serf    *serf.Serf
	EventCh chan serf.Event
}

// NewSerfGossip is a function that initializes and returns a new Serf instance
func NewSerfGossip(node string) (*Serf, error) {
	// Getting the default serf configuration
	config := serf.DefaultConfig()

	// This makes sure that the node re-joins the cluster even after it leaves
	config.RejoinAfterLeave = true

	// Creating a buffered channel for serf events
	eventChannel := make(chan serf.Event, 256)

	// Disabling the logger for serf
	config.LogOutput = io.Discard
	config.Logger = log.New(io.Discard, "", 0)
	config.MemberlistConfig.LogOutput = io.Discard

	// Assigning the created event channel to the serf configuration
	config.EventCh = eventChannel

	// Creating a serf instance with the configuration
	s, err := serf.Create(config)
	if err != nil {
		log.Fatalf("Failed to create serf instance: %v", err)
		return nil, err
	}

	// Joining the serf cluster
	_, err = s.Join([]string{node}, true)
	if err != nil {
		log.Fatalf("Failed to join cluster: %v", err)
		return nil, err
	}

	// Return a new Serf instance with the created serf instance and event channel.
	return &Serf{
		Enabled: true,
		serf:    s,
		EventCh: eventChannel,
	}, nil
}

// SyncNodes is a method on the Serf struct that synchronizes nodes between the Serf cluster
// and the provided ConsistentHash. It adds new nodes that join the cluster, and removes nodes
// that leave or fail. The synchronization is done in a loop which continues indefinitely
// until the provided context is done.
func (s *Serf) SyncNodes(ctx context.Context, consistent *hash.ConsistentHash, nodeName, port string) {
	for {
		select {
		case e := <-s.EventCh: // Listen for events from the Serf cluster.
			switch e.EventType() {
			case serf.EventMemberJoin: // If a new node has joined the cluster...
				me := e.(serf.MemberEvent)
				for _, m := range me.Members {
					if m.Name != nodeName { // And the node is not the current node...
						if _, exists := consistent.Nodes[fmt.Sprintf("%s:%d", m.Addr.String(), m.Port)]; !exists {
							fmt.Printf("Adding node %s:%s to the consistent hash\n", m.Addr.String(), port)
							// Add the new node to the consistent hash.
							if err := consistent.AddWithWeight(fmt.Sprintf("%s:%s", m.Addr.String(), port), 100); err != nil {
								fmt.Printf("error adding node %s:%d to the consistent hash: %v\n", m.Addr.String(), m.Port, err)
								// If there was an error adding the node, log the error and continue to the next node.
								continue
							}
						}
					}
				}
			case serf.EventMemberLeave, serf.EventMemberFailed, serf.EventMemberReap: // If a node has left or failed...
				me := e.(serf.MemberEvent)
				for _, m := range me.Members {
					if m.Name != nodeName { // And the node is not the current node...
						fmt.Printf("Removing node %s:%d to the consistent hash\n", m.Addr.String(), m.Port)
						if _, exists := consistent.Nodes[fmt.Sprintf("%s:%d", m.Addr.String(), m.Port)]; exists {
							// Remove the node from the consistent hash.
							if err := consistent.Remove(fmt.Sprintf("%s:%s", m.Addr.String(), port)); err != nil {
								fmt.Printf("Error removing node %s:%d from the consistent hash: %v\n", m.Addr.String(), m.Port, err)
								// If there was an error removing the node, log the error and continue to the next node.
								continue
							}
						}
					}
				}
			}
		case <-ctx.Done(): // If the context is done, stop the loop.
			fmt.Println("Stopping sync nodes")
			return
		}
	}
}

// Shutdown is a method on the Serf struct that gracefully shuts down the Serf agent.
// It first leaves the Serf cluster, and then shuts down the agent.
// It returns any errors that occurred during the shutdown process.
func (s *Serf) Shutdown() error {
	return errors.Join(s.serf.Leave(), s.serf.Shutdown())
}
