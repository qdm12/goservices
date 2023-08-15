package hooks

// LogHooks implements service handler hooks
// that logs using a leveled logger. It logs at
// the debug level and logs errors at the warning
// log level.
type LogHooks struct {
	logger Logger
}

// Logger is the logger interface required for
// the LogHooks implementation.
type Logger interface {
	Debug(s string)
	Warn(s string)
}

// NewWithLog creates a new LogHooks instance.
func NewWithLog(logger Logger) *LogHooks {
	return &LogHooks{
		logger: logger,
	}
}

// OnStart logs at the debug level the service starting.
func (h *LogHooks) OnStart(service string) {
	h.logger.Debug(service + " starting")
}

// OnStarted logs at the debug level the service started,
// and at the warning level if the service failed to start.
func (h *LogHooks) OnStarted(service string, err error) {
	if err != nil {
		h.logger.Warn("starting " + service + ": " + err.Error())
	} else {
		h.logger.Debug(service + " started")
	}
}

// OnStop logs at the debug level the service stopping.
func (h *LogHooks) OnStop(service string) {
	h.logger.Debug(service + " stopping")
}

// OnStopped logs at the debug level the service stopped,
// and at the warning level if the service failed to stop.
func (h *LogHooks) OnStopped(service string, err error) {
	if err != nil {
		h.logger.Warn("stopping " + service + ": " + err.Error())
	} else {
		h.logger.Debug(service + " stopped")
	}
}

// OnCrash logs at the warning level the service crashing
// with its eventual crash error.
func (h *LogHooks) OnCrash(service string, err error) {
	if err != nil {
		h.logger.Warn(service + " crashed: " + err.Error())
	} else {
		h.logger.Warn(service + " crashed")
	}
}
