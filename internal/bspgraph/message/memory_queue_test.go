package message_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/joshvoll/linkrus/internal/bspgraph/message"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(new(inMemoryQueueTest))

type inMemoryQueueTest struct {
	q message.Queue
}

type msg struct {
	payload string
}

func (msg) Type() string {
	return "mgs"
}

func (s *inMemoryQueueTest) SetUpTest(c *gc.C) {
	s.q = message.NewInMemoryQueue()
}

func (s *inMemoryQueueTest) TearDownTest(c *gc.C) {
	c.Assert(s.q.Close(), gc.IsNil)
}

func (s *inMemoryQueueTest) TestEnqueueDequeue(c *gc.C) {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		err := s.q.Enqueue(ctx, msg{payload: fmt.Sprint(i)})
		c.Assert(err, gc.IsNil)
	}
	c.Assert(s.q.PendingMessages(), gc.Equals, true)
	var (
		it        = s.q.Messages()
		processed int
	)
	for expNext := 9; it.Next(); expNext-- {
		got := it.Message().(msg).payload
		c.Assert(got, gc.Equals, fmt.Sprint(expNext))
		processed++
	}
	c.Assert(processed, gc.Equals, 10)
	c.Assert(it.Error(), gc.IsNil)
}

func (s *inMemoryQueueTest) TestDiscard(c *gc.C) {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		err := s.q.Enqueue(ctx, msg{payload: fmt.Sprint(i)})
		c.Assert(err, gc.IsNil)
	}
	c.Assert(s.q.PendingMessages(), gc.Equals, true)
	c.Assert(s.q.DiscardMessages(), gc.IsNil)
	c.Assert(s.q.PendingMessages(), gc.Equals, false)
}

func Test(t *testing.T) {
	gc.TestingT(t)
}
