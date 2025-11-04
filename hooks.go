package goservices

// Hooks is the interface required to hook
// into service events.
type Hooks interface {
	HooksStart
	HooksStop
	HooksCrash
}

// HooksStart is the interface required to hook
// into service start and started events.
type HooksStart interface {
	OnStart(service string)
	OnStarted(service string, err error)
}

// HooksStop is the interface required to hook
// into service stop and stopped events.
type HooksStop interface {
	OnStop(service string)
	OnStopped(service string, err error)
}

// HooksCrash is the interface required to hook
// into service crash events.
type HooksCrash interface {
	OnCrash(service string, err error)
}
