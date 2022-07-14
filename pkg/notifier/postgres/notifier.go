package postgres

import (
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/Permify/permify/pkg/logger"
)

const (
	_defaultMinReconnectInterval = 10 * time.Second
	_defaultMaxReconnectInterval = time.Minute
	_defaultSslMode              = "disable"
)

// Notifier -
type Notifier struct {
	l               *pq.Listener
	sslMode         string
	minReconnect    time.Duration
	maxReconnect    time.Duration
	shutdownTimeout time.Duration
	exit            chan bool
	terminated      chan struct{}
	ch              chan command
	subscribers     []chan<- *pq.Notification
	logger          logger.Interface
}

// New -.
func New(url string, logger logger.Interface, opts ...Option) (notifier *Notifier, err error) {
	notifier = &Notifier{
		sslMode:      _defaultSslMode,
		minReconnect: _defaultMinReconnectInterval,
		maxReconnect: _defaultMaxReconnectInterval,
		logger:       logger,
		exit:         make(chan bool),
		terminated:   make(chan struct{}),
		ch:           make(chan command),
	}

	// Custom options
	for _, opt := range opts {
		opt(notifier)
	}

	if notifier.sslMode == "disable" {
		url += "?sslmode=disable"
	}

	l := pq.NewListener(url, notifier.minReconnect, notifier.maxReconnect, nil)

	notifier.l = l

	return notifier, err
}

// Register -
func (n *Notifier) Register(channels ...string) error {
	for _, channel := range channels {
		if err := n.l.Listen(channel); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe -
func (n *Notifier) Subscribe(c chan<- *pq.Notification) {
	n.sendCommand(func() {
		n.subscribers = append(n.subscribers, c)
	})
}

// Unsubscribe -
func (n *Notifier) Unsubscribe(c chan *pq.Notification) {
	terminated := make(chan struct{})
	go func() {
		for {
			select {
			case <-c:
			case <-terminated:
				return
			}
		}
	}()
	n.sendCommand(func() {
		newSubscribers := make([]chan<- *pq.Notification, 0)
		for _, existing := range n.subscribers {
			if existing != c {
				newSubscribers = append(newSubscribers, existing)
			}
		}
		n.subscribers = newSubscribers
	})
	close(terminated)
}

// Start -
func (n *Notifier) Start() error {
	go func() {
		for {
			select {
			case <-n.terminated:
				return
			case <-n.exit:
				var err error
				err = n.l.UnlistenAll()
				err = n.l.Close()
				if err != nil {
					return
				}
				for _, sub := range n.subscribers {
					close(sub)
				}
				close(n.terminated)
			case cmd := <-n.ch:
				cmd.fun()
				close(cmd.ack)
			case notify := <-n.l.Notify:
				if notify != nil {
					for _, sub := range n.subscribers {
						sub <- notify
					}
				}
			case <-time.After(90 * time.Second):
				go func() {
					err := n.l.Ping()
					if err != nil {
						n.logger.Error(fmt.Errorf("notifier - Ping: %w", err))
					}
				}()
				continue
			}
		}
	}()

	return nil
}

// command -
type command struct {
	fun func()
	ack chan struct{}
}

// sendCommand -
func (n *Notifier) sendCommand(c func()) {
	cmd := command{c, make(chan struct{})}

	select {
	case <-n.terminated:
		return
	case n.ch <- cmd:
	}

	select {
	case <-n.terminated:
		return
	case <-cmd.ack:
		return
	}
}

// Close -
func (n *Notifier) Close() {
	select {
	case <-n.terminated:
		return
	case n.exit <- true:
	}
	<-n.terminated
}
