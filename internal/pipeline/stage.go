package pipeline

import (
	"context"
	"sync"

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

// fixedWorkerPool  model definition
type fixedWorkerPool struct {
	fifos []StageRunner
}

// FixedWorkerPool return a StageRuuner that spins up a pool containing
// numWorkers to process the payloads in parallel and emits their outputs
// to the next stages.
func FixedWorkerPool(proc Processor, numWorkers int) StageRunner {
	if numWorkers <= 0 {
		panic("FixerWorkerPool: numbers int must be > 0")
	}
	fifos := make([]StageRunner, numWorkers)
	for i := 0; i < numWorkers; i++ {
		fifos[i] = FIFO(proc)
	}
	return &fixedWorkerPool{
		fifos: fifos,
	}
}

// Run implemente the StageRunner interface with the Run Method
// spin up each worker in the pool and wait for them to exit
func (p *fixedWorkerPool) Run(ctx context.Context, params StageParams) {
	var wg sync.WaitGroup
	for i := 0; i < len(p.fifos); i++ {
		wg.Add(1)
		go func(fifoIndex int) {
			p.fifos[fifoIndex].Run(ctx, params)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

// dynamicWorkerPool defintion
type dynamicWorkerPool struct {
	proc      Processor
	tokenPool chan struct{}
}

// DyanamicWorkerPool return a StageRunner that maintains a dynamic worker pool
// that can scale to max workers for processing incoming inputs in parallel
// and emitting their outputs
func DyanamicWorkerPool(proc Processor, maxWorkers int) StageRunner {
	if maxWorkers <= 0 {
		panic("DyanamicWokerPool: must have more than 0 ")
	}
	tokenPool := make(chan struct{}, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		tokenPool <- struct{}{}
	}
	return &dynamicWorkerPool{
		proc:      proc,
		tokenPool: tokenPool,
	}
}

// Run implements the interface from the StageRunner
// if the process id not output a payload for the next stage there is nothing we need to do
// output the process data
// wait for all workers to exit by traying to empy the token pool
func (p *dynamicWorkerPool) Run(ctx context.Context, params StageParams) {
stop:
	for {
		select {
		case <-ctx.Done():
			break stop
		case payloadIn, ok := <-params.Input():
			if !ok {
				break stop
			}
			var token struct{}
			select {
			case token = <-p.tokenPool:
			case <-ctx.Done():
			}
			go func(payloadIn Payload, token struct{}) {
				defer func() { p.tokenPool <- token }()
				payloadOut, err := p.proc.Process(ctx, payloadIn)
				if err != nil {
					wrappedErr := xerrors.Errorf("pipeline stage: %d : %w ", params.StageIndex(), err)
					maybeEmitError(wrappedErr, params.Error())
					return
				}
				if payloadOut == nil {
					payloadIn.MarkAsProcessed()
					return
				}
				select {
				case params.Output() <- payloadOut:
				case <-ctx.Done():
				}
			}(payloadIn, token)
		}
	}
	for i := 0; i < cap(p.tokenPool); i++ {
		<-p.tokenPool
	}
}

// broadcast model definition
type broadcast struct {
	fifos []StageRunner
}

// Broadcast returns a StageRunner that passes a copy of each incoming payload
// to all specified processors and emits their outputs to the next stage.
func Broadcast(procs ...Processor) StageRunner {
	if len(procs) == 0 {
		panic("Broadcast: at least one processor must be specific")
	}
	fifos := make([]StageRunner, len(procs))
	for i, p := range procs {
		fifos[i] = FIFO(p)
	}
	return &broadcast{
		fifos: fifos,
	}
}

// Run implement the Run interface from StageRunner
func (b *broadcast) Run(ctx context.Context, params StageParams) {
	var (
		wg   sync.WaitGroup
		inCh = make([]chan Payload, len(b.fifos))
	)
	for i := 0; i < len(b.fifos); i++ {
		wg.Add(1)
		inCh[i] = make(chan Payload)
		go func(fifoIndex int) {
			fifoParams := &workerParams{
				stage: params.StageIndex(),
				inCh:  inCh[fifoIndex],
				outCh: params.Output(),
				errCh: params.Error(),
			}
			b.fifos[fifoIndex].Run(ctx, fifoParams)
			wg.Done()
		}(i)
	}
done:
	for {
		select {
		case <-ctx.Done():
			break done
		case payload, ok := <-params.Input():
			if !ok {
				break done
			}
			for i := len(b.fifos) - 1; i >= 0; i++ {
				// as each FIFO might modify the payload, to
				// avoid data race we need to copy of each payload
				// for all FiFO execpt the first one
				var fifoPayload = payload
				if i != 0 {
					fifoPayload = payload.Clone()
				}
				select {
				case <-ctx.Done():
					break done
				case inCh[i] <- fifoPayload:

				}
			}
		}
	}
	for _, ch := range inCh {
		close(ch)
	}
	wg.Wait()

}
