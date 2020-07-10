package message

import "context"

// Message is implemented by types that can be processd by a Queue.
type Message interface {
	Type() string
}

// Queue is implemented by types that can serve as message queues/
type Queue interface {
	// Cleanly shutdown the queue.
	Close() error

	// Enqueue inserts a message to the end of the queue.
	Enqueue(ctx context.Context, msg Message) error

	// PendingMessages returns true if the queue contains any messages.
	PendingMessages() bool

	// Flush drops all pending message from the queue.
	DiscardMessages() error

	// Messages returns an iterator for accesing the queued message.
	Messages() Iterator
}

// Iterator provides an API for iteranting a list of messages
type Iterator interface {
	// Next advance the iterator so that the next message can be retrieved
	// via a call to Message(). if no more messages are available or an
	// error occurs, Next() returns false.
	Next() bool

	// Message returns the message currently pointed to by the iterator
	Message() Message

	// Error returns the last error that the iterator encountered.
	Error() error
}

// QueueFactory is a function that can create new Queue instances.
type QueueFactory func() Queue
