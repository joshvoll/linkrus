package bspgraph

import "context"

// ExecutorCallbacks encapsulates a series of callbacks that are invoked by an
// Executor instance on a graph. All callbacks are optional and will be ignored
// if not specified.
type ExecutorCallbacks struct {
	// PreStep, if defined, is invoked before running the next superstep.
	// This is a good place to initialize variables, aggregators etc. that
	// will be used for the next superstep.
	PreStep func(ctx context.Context, g *Graph) error

	// PosStep if defined, is invoked after running a superstep.
	PostStep func(ctx context.Context, g *Graph, activeInStep int) error

	// PostStepKeepRunning, if defined, is invoked after running a superstep
	// to decide whether the stop condition for terminating the run has
	// been met. The number of the active vertices in the last step is
	// passed as the second argument.
	PostStepKeepRunning func(ctx context.Context, g *Graph, activeInStep int) (bool, error)
}

// Executor wraps a Graph instance and provides an orchestration layer for
// executing super-steps until an error occurs or an exit condition is met.
// Users can provide an optional set of callbacks to be executed before and
// after each super-step.
type Executor struct {
	g  *Graph
	cb ExecutorCallbacks
}

// NewExecutor returns an Executor instance for graph g that invokes the
// provided list of callbacks inside each execution loop.
func NewExecutor(g *Graph, cb ExecutorCallbacks) *Executor {
	patchEmptyCallbacks(&cb)
	g.superstep = 0
	return &Executor{
		g:  g,
		cb: cb,
	}
}

// RunToCompletion keeps executing supersteps until the context expires, an
// error occurs or one of the Pre/PostStepKeepRunning callbacks specified at
// configuration time returns false.
func (ex *Executor) RunToCompletion(ctx context.Context) error {
	return ex.run(ctx, -1)
}

// run is a private method to runt the steps
func (ex *Executor) run(ctx context.Context, maxSteps int) error {
	var (
		activeInStep int
		err          error
		keepRunning  bool
		cb           = ex.cb
	)
	for ; maxSteps != 0; ex.g.superstep, maxSteps = ex.g.superstep+1, maxSteps-1 {
		if err = ensureContextnotExpired(ctx); err != nil {
			break
		} else if err = cb.PreStep(ctx, ex.g); err != nil {
			break
		} else if activeInStep, err = ex.g.step(); err != nil {
			break
		} else if err = cb.PostStep(ctx, ex.g, activeInStep); err != nil {
			break
		} else if keepRunning, err = cb.PostStepKeepRunning(ctx, ex.g, activeInStep); !keepRunning || err != nil {
			break
		}
	}

	return err
}

// ensureContextnotExpired private method for expiration context
func ensureContextnotExpired(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// patchEmptyCallbacks just return if the patch is eampty
func patchEmptyCallbacks(cb *ExecutorCallbacks) {
	if cb.PreStep == nil {
		cb.PreStep = func(context.Context, *Graph) error { return nil }
	}
	if cb.PostStep == nil {
		cb.PostStep = func(context.Context, *Graph, int) error { return nil }
	}
	if cb.PostStepKeepRunning == nil {
		cb.PostStepKeepRunning = func(context.Context, *Graph, int) (bool, error) {
			return true, nil
		}
	}
}
