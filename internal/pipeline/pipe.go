package pipeline

// maybeEmitError attempt to queue error and buffered on a channel
// if the channel if full the error is droped
func maybeEmitError(err error, errCha chan<- error) {
	select {
	case errCha <- err:
	default:
	}

}
