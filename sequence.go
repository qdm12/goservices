package goservices

import (
	"context"
	"fmt"
	"sync"
)

var _ Service = (*Sequence)(nil)

// Sequence is a sequence of services to start and stop in
// a pre-defined order. It implements the Service interface
// itself.
type Sequence struct {
	name           string
	servicesStart  []Service
	servicesStop   []Service
	hooks          Hooks
	startStopMutex sync.Mutex
	state          State
	stateMutex     sync.RWMutex
	fanIn          *errorsFanIn
	// runningServices contains service names that are currently running.
	runningServices map[string]struct{}
	interceptStop   chan struct{}
	interceptDone   chan struct{}
}

// NewSequence creates a new sequence of services given the settings,
// and returns an error if any setting is not valid.
func NewSequence(settings SequenceSettings) (sequence *Sequence, err error) {
	settings.setDefaults()

	err = settings.validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	servicesStart := make([]Service, len(settings.ServicesStart))
	copy(servicesStart, settings.ServicesStart)

	servicesStop := make([]Service, len(settings.ServicesStop))
	copy(servicesStop, settings.ServicesStop)

	return &Sequence{
		name:            settings.Name,
		servicesStart:   servicesStart,
		servicesStop:    servicesStop,
		hooks:           settings.Hooks,
		state:           StateStopped,
		runningServices: make(map[string]struct{}, len(servicesStart)),
	}, nil
}

func (s *Sequence) String() string {
	if s.name == "" {
		return "sequence"
	}
	return "sequence " + s.name
}

// Start starts services in the order specified by the
// sequence of services.
//
// If a service fails to start, the `startErr` is returned
// and all other running services are stopped in the order
// specified by the stop sequence of services.
//
// If a service fails after this method call returns without error,
// all other running services are stopped and the error is
// returned in the `runError` channel which is then closed.
// A caller should listen on `runError` until the `Stop` method
// call fully completes, since a run error can theoretically happen
// at the same time the caller calls `Stop` on the sequence.
//
// If the sequence is already running then a start error ErrAlreadyStarted
// is returned.
//
// If the context is canceled, the current service starting is canceled,
// all already running services are stopped and the context error is wrapped
// in the `startErr` returned.
func (s *Sequence) Start(ctx context.Context) (runError <-chan error, startErr error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	// Lock the state in case the sequence is already running.
	s.stateMutex.RLock()
	state := s.state
	// no need to keep a lock on the state since the `startStopMutex`
	// prevents concurrent calls to `Start` and `Stop`.
	s.stateMutex.RUnlock()
	if state == StateRunning {
		return nil, fmt.Errorf("%s: %w", s, ErrAlreadyStarted)
	}

	s.state = StateStarting

	var fanInErrorCh <-chan serviceError
	s.fanIn, fanInErrorCh = newErrorsFanIn()

	for _, service := range s.servicesStart {
		serviceString := service.String()

		s.hooks.OnStart(serviceString)
		serviceRunError, err := service.Start(ctx)
		s.hooks.OnStarted(serviceString, err)

		if err != nil {
			err = addCtxErrorIfNeeded(err, ctx.Err())
			_ = s.stop()
			return nil, fmt.Errorf("starting %s: %w", serviceString, err)
		}

		s.runningServices[serviceString] = struct{}{}

		s.fanIn.add(serviceString, serviceRunError)
	}

	// Hold the state mutex until the intercept run error goroutine is ready
	// and we change the state to running.
	// This is as such because the intercept goroutine may catch a service run error
	// as soon as it starts, and try to set the sequence state as crashed.
	// With this lock, the goroutine must wait for the mutex unlock below before
	// changing the state to crashed.
	s.stateMutex.Lock()

	runErrorCh := make(chan error)
	interceptReady := make(chan struct{})
	s.interceptStop = make(chan struct{})
	s.interceptDone = make(chan struct{})
	go s.interceptRunError(interceptReady, fanInErrorCh, runErrorCh)
	<-interceptReady

	s.state = StateRunning
	s.stateMutex.Unlock()

	return runErrorCh, nil
}

// interceptRunError, if it catches an error from the input
// channel, registers the crashed service of the sequence,
// stops other running services and forwards the error
// to the output channel and finally closes this channel.
// If the stop channel triggers, the function returns.
func (s *Sequence) interceptRunError(ready chan<- struct{},
	input <-chan serviceError, output chan<- error) {
	defer close(s.interceptDone)
	close(ready)

	select {
	case <-s.interceptStop:
	case serviceErr := <-input:
		// Lock the state mutex in case we are stopping
		// or trying to stop the sequence at the same time.
		s.stateMutex.Lock()
		if s.state == StateStopping {
			// Discard the eventual single service run error
			// fanned-in if we are stopping the sequence.
			s.stateMutex.Unlock()
			return
		}

		// The first and only service fanned-in run error was
		// caught and we are not currently stopping the sequence.
		s.state = StateCrashed
		delete(s.runningServices, serviceErr.serviceName)
		s.stateMutex.Unlock()

		s.hooks.OnCrash(serviceErr.serviceName, serviceErr.err)
		_ = s.stop()
		output <- &serviceErr
		close(output)
	}
}

// Stop stops running services of the sequence
// in the order specified by the sequence of services.
// If an error occurs for any of the service stop,
// the other running services will still be stopped.
// Only the first non nil service stop error encountered
// is returned, but the hooks can be used to process each
// error returned.
// If the sequence is already stopped, the `ErrAlreadyStopped` error
// is returned.
func (s *Sequence) Stop() (err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	s.stateMutex.Lock()
	switch s.state {
	case StateRunning: // continue stopping the sequence
	case StateCrashed:
		s.stateMutex.Unlock()
		// sequence is already stopped or stopping from
		// the intercept goroutine, so just wait for the
		// intercept goroutine to finish.
		<-s.interceptDone
		return nil
	case StateStopped:
		s.stateMutex.Unlock()
		return fmt.Errorf("%s: %w", s, ErrAlreadyStopped)
	case StateStarting, StateStopping:
		s.stateMutex.Unlock()
		panic("bad sequence implementation code: this code path should be unreachable")
	}
	s.state = StateStopping
	s.stateMutex.Unlock()

	err = s.stop()

	// Stop the intercept error goroutine after we stop
	// all the sequence services. This means the fan in might
	// send one error to the intercept goroutine, but it will
	// discard it since we are in the stopping state.
	// The error fan in takes care of reading and discarding
	// errors from other services once it caught the first error.
	close(s.interceptStop)
	<-s.interceptDone

	s.state = StateStopped

	return err
}

// stop stops all running services in the sequence.
// If a service fails to stop in the sequence, its error
// is returned but the other services are still stopped.
// All service stop errors are wrapped together in the format
// stopping <name_1>: %w; stopping <name_2>: %w; ...
// and can be checked individually with errors.Is(err, ErrDefined).
// Hooks can be used to access each stopping and stop result.
func (s *Sequence) stop() (err error) {
	for _, service := range s.servicesStop {
		serviceString := service.String()

		_, running := s.runningServices[serviceString]
		if !running {
			continue
		}

		s.hooks.OnStop(serviceString)
		stopErr := service.Stop()
		s.hooks.OnStopped(serviceString, stopErr)
		err = addStopError(err, serviceString, stopErr)
		delete(s.runningServices, serviceString)
	}

	// Only stop the fan in after stopping all services
	// so it can read and discard any eventual run errors
	// from these whilst we stop them.
	s.fanIn.stop()

	return err
}
