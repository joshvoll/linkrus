package bspgraph

import (
	"context"

	"github.com/joshvoll/linkrus/internal/bspgraph/message"
)

// Aggregator is implemented by types that provide concurrent-safe aggregation
// primitives (e.g. counters, min/max, topN).
type Aggregator interface {
	// Type Return the type of the aggregator
	Type() string

	// Set the aggregator to specified value.
	Set(val interface{})

	// Get the current aggregator value
	Get() interface{}

	// Aggregate updates the aggregator value base on the provided value
	Aggregate(val interface{})

	// Delta returns the change in the aggregator's value since the last
	// call to Delta. The delta values can be used in distributed
	// aggregator use-cases to reduce local, partially-aggregated values
	// into a single value across by feeding them into the Aggregate method
	// of a top-level aggregator.
	//
	// For example, in a distributed counter scenario, each node maintains
	// its own local counter instance. At the end of each step, the master
	// node calls delta on each local counter and aggregates the values
	// to obtain the correct total which is then pushed back to the workers.
	Delta() interface{}
}

// Relayer is implemented by types that can relay messages to vertices that
// are managed by a remote graph instance.
type Relayer interface {
	// Relay a message to a vertex that is not known locally. Calls to
	// Relay must return ErrDestinationIsLocal if the provided dst value is
	// not a valid remote destination.
	Relay(ctx context.Context, dst string, msg message.Message) error
}

// ComputeFunc is a function that a graph instance invokes on each vertex when
// executing a superstep.
type ComputeFunc func(g *Graph, v *Vertex, msgIt message.Iterator) error
