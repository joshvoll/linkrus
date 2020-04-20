package pipeline

// workersParams model defintiion
type workerParams struct {
	stage int
	// channel for the workers definitions
	inCh  <-chan Payload
	outCh chan<- Payload
	errCh chan<- error
}

// all methods for the workesr params
func (p *workerParams) StageIndex() int        { return p.stage }
func (p *workerParams) Input() <-chan Payload  { return p.inCh }
func (p *workerParams) Output() chan<- Payload { return p.outCh }
func (p *workerParams) Error() chan<- error    { return p.errCh }

// maybeEmitError attempt to queue error and buffered on a channel
// if the channel if full the error is droped
func maybeEmitError(err error, errCha chan<- error) {
	select {
	case errCha <- err:
	default:
	}

}
