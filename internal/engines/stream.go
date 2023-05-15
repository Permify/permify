package engines

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

type Stream struct {
	ResultChan chan *base.Subject
	g          *errgroup.Group
	ctx        context.Context
	// Results
	Results []*base.Subject
	mu      sync.Mutex
	// callback function to handle the result of each permission check
	callback func(subject *base.Subject)
}

func NewStream(ctx context.Context, callback func(subject *base.Subject)) *Stream {
	return &Stream{
		ResultChan: make(chan *base.Subject),
		g:          &errgroup.Group{},
		ctx:        ctx,
		mu:         sync.Mutex{},
		callback:   callback,
	}
}

func (s *Stream) Publish(d *base.Subject) {
	s.g.Go(func() error {
		select {
		case s.ResultChan <- d:
			return nil
		case <-s.ctx.Done():
			return s.ctx.Err()
		}
	})
}

// AddResult -
func (s *Stream) AddResult(subject ...*base.Subject) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Results = append(s.Results, subject...)
}

// ConsumeData -
func (s *Stream) ConsumeData() {
	s.g.Go(func() error {
		for {
			select {
			case d, ok := <-s.ResultChan:
				if !ok {
					return nil
				}

				s.callback(d) // call the callback function
			case <-s.ctx.Done():
				return s.ctx.Err()
			}
		}
	})
}

// Wait waits for all goroutines in the errgroup to finish.
// Returns an error if any of the goroutines encounter an error.
func (s *Stream) Wait() error {
	if err := s.g.Wait(); err != nil {
		return err
	}
	return nil
}
