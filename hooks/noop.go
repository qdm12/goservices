package hooks

// NoopHooks implements service handler hooks
// that do no operation.
type NoopHooks struct{}

// NewNoop returns a new NoopHooks instance.
func NewNoop() *NoopHooks {
	return &NoopHooks{}
}

// OnStart does nothing.
func (h *NoopHooks) OnStart(string) {}

// OnStarted does nothing.
func (h *NoopHooks) OnStarted(string, error) {}

// OnStop does nothing.
func (h *NoopHooks) OnStop(string) {}

// OnStopped does nothing.
func (h *NoopHooks) OnStopped(string, error) {}

// OnCrash does nothing.
func (h *NoopHooks) OnCrash(string, error) {}
