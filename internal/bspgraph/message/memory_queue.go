package message

import (
	"context"
	"sync"
)

// inMemoryQueue imeplements a queue that stores the message in the memory. Messages
// can be enqueued concurrently byt the returned iterator is not safe for
// concurrent access.
type inMemoryQueue struct {
	mu         sync.Mutex
	msgs       []Message
	latchedMsg Message
}

// NewInMemoryQueue create a new in-memory queue instnace. this function can
// server as a QueueFactory
func NewInMemoryQueue() Queue {
	return new(inMemoryQueue)
}

// Close implementes the Queue
func (*inMemoryQueue) Close() error {
	return nil
}

// Enqueue implements Queue.
func (q *inMemoryQueue) Enqueue(ctx context.Context, msg Message) error {
	q.mu.Lock()
	q.msgs = append(q.msgs, msg)
	q.mu.Unlock()
	return nil
}

// PendingMessages implementes Queue.
func (q *inMemoryQueue) PendingMessages() bool {
	q.mu.Lock()
	pending := len(q.msgs) != 0
	q.mu.Unlock()
	return pending
}

// DiscardMessages implements Queue.
func (q *inMemoryQueue) DiscardMessages() error {
	q.mu.Lock()
	q.msgs = q.msgs[:0]
	q.latchedMsg = nil
	q.mu.Unlock()
	return nil
}

// Messages implemente Queue.
func (q *inMemoryQueue) Messages() Iterator {
	return q
}

// Next imeplentes Iterator
// Dequeue message from the tail of the queue.
func (q *inMemoryQueue) Next() bool {
	q.mu.Lock()
	qLen := len(q.msgs)
	if qLen == 0 {
		q.mu.Unlock()
		return false
	}
	q.latchedMsg = q.msgs[qLen-1]
	q.msgs = q.msgs[:qLen-1]
	q.mu.Unlock()
	return true
}

// Message implementes Queue.
func (q *inMemoryQueue) Message() Message {
	q.mu.Lock()
	msg := q.latchedMsg
	q.mu.Unlock()
	return msg
}

// Error implemente the erro from Queue.
func (q *inMemoryQueue) Error() error {
	return nil
}
