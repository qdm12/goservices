package goservices

import (
	"context"
	"fmt"
	"sync"
)

var _ Service = (*Restarter)(nil)

// Restarter implements a service which restarts an
// underlying service if it crashes. The restarter
// only crashes if the underlying services fails to
// start on a subsequent run.
type Restarter struct {
	service        Service
	hooks          Hooks
	startStopMutex sync.Mutex
	state          State
	stateMutex     sync.RWMutex
	interceptStop  chan struct{}
	interceptDone  chan struct{}
}

// NewRestarter creates a new restarter given the settings.
// It returns an error if any of the settings is not valid.
func NewRestarter(settings RestarterSettings) (restarter *Restarter, err error) {
	settings.setDefaults()

	err = settings.validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	return &Restarter{
		service: settings.Service,
		hooks:   settings.Hooks,
		state:   StateStopped,
	}, nil
}

func (r *Restarter) String() string {
	return r.service.String()
}

// Start starts the underlying service.
//
// If the underlying service fails to start, the `startErr` is returned.
//
// If the underlying service fails after this method call returns
// without error, it is automatically restarted and no error is emitted
// in the `runError` channel.
//
// If a subsequent service start fails, the start error is sent in the
// `runError` channel, this channel is closed and the restarter stops.
// A caller should listen on `runError` until the `Stop` method
// call fully completes, since a run error can theoretically happen
// at the same time the caller calls `Stop` on the restarter.
//
// If the restarter is already running, the `ErrAlreadyStarted` error
// is returned.
//
// If the context is canceled, the service starting operation is canceled,
// and the context error is wrapped in the `startErr` returned.
func (r *Restarter) Start(ctx context.Context) (runError <-chan error, startErr error) {
	// Prevent concurrent Stop and Start calls.
	r.startStopMutex.Lock()
	defer r.startStopMutex.Unlock()

	// Lock the state in case the restarter is already running.
	r.stateMutex.RLock()
	state := r.state
	// no need to keep a lock on the state since the `startStopMutex`
	// prevents concurrent calls to `Start` and `Stop`.
	r.stateMutex.RUnlock()
	if state == StateRunning {
		return nil, fmt.Errorf("%s: %w", r, ErrAlreadyStarted)
	}

	r.state = StateStarting

	serviceString := r.service.String()

	r.hooks.OnStart(serviceString)
	serviceRunError, startErr := r.service.Start(ctx)
	r.hooks.OnStarted(serviceString, startErr)

	if startErr != nil {
		startErr = addCtxErrorIfNeeded(startErr, ctx.Err())
		return nil, startErr
	}

	// Hold the state mutex until the intercept run error goroutine is ready
	// and we change the state to running.
	// This is as such because the intercept goroutine may catch a service run error
	// as soon as it starts, and try to set the restarter state as crashed.
	// With this lock, the goroutine must wait for the mutex unlock below before
	// changing the state to crashed.
	r.stateMutex.Lock()

	interceptReady := make(chan struct{})
	runErrorCh := make(chan error)
	r.interceptStop = make(chan struct{})
	r.interceptDone = make(chan struct{})
	go r.interceptRunError(interceptReady, serviceString, //nolint:contextcheck
		serviceRunError, runErrorCh)
	<-interceptReady

	r.state = StateRunning
	r.stateMutex.Unlock()

	return runErrorCh, nil
}

func (r *Restarter) interceptRunError(ready chan<- struct{},
	serviceName string, input <-chan error, output chan<- error) {
	defer close(r.interceptDone)
	close(ready)

	for {
		select {
		case <-r.interceptStop:
			return
		case err := <-input:
			// Lock the state mutex in case we are stopping
			// or trying to stop the restarter at the same time.
			r.stateMutex.Lock()
			if r.state == StateStopping {
				// Discard the eventual single service run error
				// if we are stopping the restarter.
				r.stateMutex.Unlock()
				return
			}

			r.hooks.OnCrash(serviceName, err)

			r.hooks.OnStart(serviceName)

			// When an error is received from the input channel and
			// the restarter is not stopping yet, the state mutex is
			// locked and therefore it is not possible to stop the
			// restarter at the same time as the execution of the code
			// below. Therefore, it is fine to set the service start
			// context as context.Background() and not cancel it.
			var startErr error
			input, startErr = r.service.Start(context.Background())
			r.hooks.OnStarted(serviceName, startErr)

			if startErr != nil {
				r.state = StateCrashed
				r.stateMutex.Unlock()
				output <- fmt.Errorf("restarting after crash: %w", startErr)
				close(output)
				return
			}
			r.state = StateRunning
			r.stateMutex.Unlock()
		}
	}
}

// Stop stops the underlying service and the internal
// run error restart-watcher goroutine.
// If the restarter is already stopped, the `ErrAlreadyStopped` error
// is returned.
// Note if the restarter is currently restarting the underlying
// service, it has to finish the start before the stopping can start.
func (r *Restarter) Stop() (err error) {
	r.startStopMutex.Lock()
	defer r.startStopMutex.Unlock()

	r.stateMutex.Lock()
	switch r.state {
	case StateRunning: // continue stopping the restarter
	case StateCrashed:
		// service crashed and failed to restart, just wait
		// for the intercept goroutine to finish.
		<-r.interceptDone
		return nil
	case StateStopped:
		r.stateMutex.Unlock()
		return fmt.Errorf("%s: %w", r, ErrAlreadyStopped)
	case StateStarting, StateStopping:
		r.stateMutex.Unlock()
		panic("bad restarter implementation code: this code path should be unreachable")
	}
	r.state = StateStopping
	r.stateMutex.Unlock()

	serviceString := r.service.String()

	r.hooks.OnStop(serviceString)
	err = r.service.Stop()
	r.hooks.OnStopped(serviceString, err)

	// Stop the intercept error goroutine after we stop
	// the restarter underlying service.
	close(r.interceptStop)
	<-r.interceptDone

	r.state = StateStopped

	return err
}
