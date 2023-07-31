package goservices

import "fmt"

// State is the state of a service.
// Is it exported to ease the implementation of services.
type State uint8

const (
	// StateStopped is the state of a service that is stopped.
	StateStopped State = iota
	// StateStarting is the state of a service that is starting.
	StateStarting
	// StateRunning is the state of a service that is running.
	StateRunning
	// StateStopping is the state of a service that is stopping.
	StateStopping
	// StateCrashed is the state of a service that has crashed.
	StateCrashed
)

func (s State) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateCrashed:
		return "crashed"
	default:
		panic(fmt.Sprintf("State %d has no corresponding string", s))
	}
}
