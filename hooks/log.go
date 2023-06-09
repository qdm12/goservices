package hooks

type LogHooks struct {
	logger Logger
}

type Logger interface {
	Debug(s string)
	Warn(s string)
}

func NewWithLog(logger Logger) *LogHooks {
	return &LogHooks{
		logger: logger,
	}
}

func (h *LogHooks) OnStart(service string) {
	h.logger.Debug(service + " starting")
}

func (h *LogHooks) OnStarted(service string, err error) {
	if err != nil {
		h.logger.Warn("starting " + service + ": " + err.Error())
	} else {
		h.logger.Debug(service + " started")
	}
}

func (h *LogHooks) OnStop(service string) {
	h.logger.Debug(service + " stopping")
}

func (h *LogHooks) OnStopped(service string, err error) {
	if err != nil {
		h.logger.Warn("stopping " + service + ": " + err.Error())
	} else {
		h.logger.Debug(service + " stopped")
	}
}

func (h *LogHooks) OnCrash(service string, err error) {
	if err != nil {
		h.logger.Warn(service + " crashed: " + err.Error())
	} else {
		h.logger.Debug(service + " crashed")
	}
}
