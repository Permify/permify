package balancer

import (
	"container/list"
	"sync"
)

// Queue is a basic FIFO queue based on a doubly linked list.
type Queue struct {
	list *list.List
	mu   sync.Mutex
}

// NewQueue returns a new queue.
func NewQueue() *Queue {
	return &Queue{
		list: list.New(),
	}
}

// EnQueue adds an item to the end of the queue.
func (q *Queue) EnQueue(value interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.list.PushBack(value)
}

// DeQueue removes and returns the first item in the queue.
func (q *Queue) DeQueue() (interface{}, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	e := q.list.Front()
	if e == nil {
		return nil, false
	}
	v := e.Value
	q.list.Remove(e)
	return v, true
}

// IsEmpty returns true if the queue is empty.
func (q *Queue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.list.Len() == 0
}

// Len returns the number of items in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.list.Len()
}
