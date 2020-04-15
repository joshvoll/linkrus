package pipeline

import (
	"context"

	"golang.org/x/xerrors"
)

// fifo is a private definition struct that instance the Processor interface
type fifo struct {
	proc Processor
}

// FIFO returns a StageRunner the process incoming payloads in a first-in, first out fashion
// Each input is passed to the specified processor and its output is emited to the next stage
func FIFO(proc Processor) StageRunner {
	return fifo{
		proc: proc,
	}
}

// Run implementation of the StageRunner interface
// it is a infinite loop to run and recieve all request coming int, like a proxy service
func (r fifo) Run(ctx context.Context, params StageParams) {
	for {
		select {
		case <-ctx.Done():
			return
		case payloadIn, ok := <-params.Input():
			if !ok {
				return
			}
			payloadOut, err := r.proc.Process(ctx, payloadIn)
			if err != nil {
				wrappedErr := xerrors.Errorf("pipeline stage %d : %w ", params.StageIndex(), err)
				maybeEmitError(wrappedErr, params.Error())
				return
			}
			if payloadOut == nil {
				payloadIn.MarkAsProcessed()
				continue
			}
			select {
			case params.Output() <- payloadOut:
			case <-ctx.Done():
				return
			}
		}
	}

}
