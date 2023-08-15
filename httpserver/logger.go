package httpserver

// Infoer is the logging interface required by the
// HTTP server service implementation.
type Infoer interface {
	Info(message string)
}

type noopLogger struct{}

func (noopLogger) Info(_ string) {}
