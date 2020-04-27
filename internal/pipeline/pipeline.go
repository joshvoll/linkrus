package pipeline

import "context"

// Payload is implemented by the pipeline by values that can be sent through a pipeline.
type Payload interface {
	// Clone return a new payload that is deep-copy of the original
	Clone() Payload

	// MarkAsProcessed()  envolve by the pipeline when the payload either
	// reaches the pipeline sink or it get discarded by one of the
	// pipeline stages
	MarkAsProcessed()
}

// Processor is implemente by that can process payload as part of the
// pipeline stage
type Processor interface {
	// Process operate on the input of the payload and return back a new payload.
	// to be forwarder to the next pipeline stage. Processor may also opt
	// to prevent the payload from reaching the rest of the pipeline by
	// returning nil payload value instead.
	Process(context.Context, Payload) (Payload, error)
}

// ProcessorFunc is an adapter to allow to use as plan functions as Processor instance
// if is a function with appropier signature. ProcessorFunc(f)
// is a Processor that call f.
type ProcessorFunc func(context.Context, Payload) (Payload, error)

// Process call f(ctx, p).
func (f ProcessorFunc) Process(ctx context.Context, p Payload) (Payload, error) {
	return f(ctx, p)
}

// StageParams encapsulate the information required for executing a pipeline stage.
// The pipeline pass a StageParams instance to the Run() method of each stage
type StageParams interface {
	// StageIndex return the position of the stage
	StageIndex() int

	// Input return the channel fo reading the input of the payload
	Input() <-chan Payload

	// Output return the channel for writting the ouput payload
	Output() chan<- Payload

	// Error return the error for writting erros for each stage
	Error() chan<- error
}

// StageRunner is implemented by types that can be strung togetther for form
// multi-stage pipeline.
type StageRunner interface {
	// Run implement de process logic for this stage by reading
	// incomming payloads from an input channel, processing them and
	// outputting them to another channel
	//
	// Calls to Run expected to block until:
	// - this stage input channel are close OR
	// - the provide context is expires OR
	// - an error occurs while processing payloads.
	Run(context.Context, StageParams)
}

// Source is implement by types generated payload instance wich can be use
// as inputs to the Pipeline instance.
type Source interface {
	// Next fetch the next payload from the source. if no more items are available
	// or an error occurs, call to Next return false
	Next(context.Context) bool

	// Payload return the next payload of the source
	Payload() Payload

	// Error return the last erro observed by the source
	Error() error
}
