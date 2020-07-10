package bspgraph

import "github.com/joshvoll/linkrus/internal/bspgraph/message"

// Vertex represent a vertex in the graph.
type Vertex struct {
	id       string
	value    interface{}
	active   bool
	msgQueue [2]message.Queue
	edges    []*Edge
}

/*
   Verxtex helper functions
   defining the following set of helpers methods so that
   we can access and/or safely manipulate the Vertex instance
*/

// ID returns the verxtex ID
func (v *Vertex) ID() string {
	return v.id
}

// Edges return the list of outgoing edges from this vertex.
func (v *Vertex) Edges() []*Edge {
	return v.edges
}

// Freeze marks the vertex as inactive. Inactive vertices will not be processed
// in the following supersteps unless they receive a message in which case they
// will be re-activated.
func (v *Vertex) Freeze() {
	v.active = false
}

// Value return he value associated with this vertex.
func (v *Vertex) Value() interface{} {
	return v.value
}

// SetValue sets the value associeted with this vertex.
func (v *Vertex) SetValue(val interface{}) {
	v.value = val
}

// Edge represents a directed edge in the graph.
type Edge struct {
	value interface{}
	dstID string
}

// DstID return the vertex ID that correspond to this edge's target endpoint
func (e *Edge) DstID() string {
	return e.dstID
}

// Value return the value vertex of the correspond this edge.
func (e *Edge) Value() interface{} {
	return e.value
}

// SetValue sets the value of the correspondent edge
func (e *Edge) SetValue(val interface{}) {
	e.value = val
}

// Graph implements a parallel graph processor base on the concepts described
// in the Pregel paper.
type Graph struct {
}
