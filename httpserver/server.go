// Package httpserver implements an HTTP server.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/qdm12/goservices"
)

// Server is an HTTP server implementation.
type Server struct {
	// Dependencies injected
	settings Settings

	// Internal fields
	server                http.Server
	startStopMutex        sync.Mutex
	state                 goservices.State
	stateMutex            sync.RWMutex
	listeningAddress      string
	listeningAddressMutex sync.RWMutex
}

// New creates a new HTTP server with a name, listening on
// the address specified and using the HTTP handler provided.
func New(settings Settings) (server *Server, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	return &Server{
		settings: settings,
		state:    goservices.StateStopped,
	}, nil
}

func (s *Server) String() string {
	if *s.settings.Name == "" {
		return "http server"
	}
	return *s.settings.Name + " http server"
}

// GetAddress obtains the address the HTTP server is listening on.
func (s *Server) GetAddress() (address string) {
	s.listeningAddressMutex.RLock()
	defer s.listeningAddressMutex.RUnlock()
	return s.listeningAddress
}

// Start starts the HTTP server service.
// The context argument is ignored since starting the
// HTTP server is rather instantaneous.
// The listening address is accessible only AFTER the
// call to Start completes, to ensure the server is started
// successfully.
func (s *Server) Start(_ context.Context) (runError <-chan error, err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	// Lock the state in case the server service is already running.
	s.stateMutex.RLock()
	state := s.state
	// no need to keep a lock on the state since the `startStopMutex`
	// prevents concurrent calls to `Start` and `Stop`.
	s.stateMutex.RUnlock()
	if state == goservices.StateRunning {
		return nil, fmt.Errorf("%s: %w", s, goservices.ErrAlreadyStarted)
	}

	s.state = goservices.StateStarting

	listener, err := net.Listen("tcp", *s.settings.Address)
	if err != nil {
		return nil, err
	}

	s.listeningAddressMutex.Lock()
	defer s.listeningAddressMutex.Unlock()
	s.listeningAddress = listener.Addr().String()
	s.server = http.Server{
		Addr:              s.listeningAddress,
		Handler:           s.settings.Handler,
		ReadHeaderTimeout: s.settings.ReadHeaderTimeout,
		ReadTimeout:       s.settings.ReadTimeout,
	}
	s.settings.Logger.Info(fmt.Sprintf("%s listening on %s", s, s.listeningAddress))

	runErrorBiDirectional := make(chan error)
	runError = runErrorBiDirectional
	ready := make(chan struct{})

	// Hold the state mutex locked in case the server Serve
	// function returns an error instantly.
	s.stateMutex.Lock()

	go func() {
		close(ready)
		err = s.server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		s.stateMutex.Lock()
		s.state = goservices.StateCrashed
		s.stateMutex.Unlock()
		runErrorBiDirectional <- err
	}()

	<-ready
	s.state = goservices.StateRunning
	s.stateMutex.Unlock()

	return runError, nil
}

// Stop stops the HTTP server service.
func (s *Server) Stop() (err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	s.stateMutex.Lock()
	switch s.state {
	case goservices.StateRunning: // continue stopping the server
	case goservices.StateCrashed: // server is already stopped
		s.stateMutex.Unlock()
		return nil
	case goservices.StateStopped:
		s.stateMutex.Unlock()
		return fmt.Errorf("%s: %w", s, goservices.ErrAlreadyStopped)
	case goservices.StateStarting, goservices.StateStopping:
		s.stateMutex.Unlock()
		panic("bad implementation code: this code path should be unreachable")
	}
	s.state = goservices.StateStopping
	s.stateMutex.Unlock()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), s.settings.ShutdownTimeout)
	defer cancel()
	err = s.server.Shutdown(shutdownCtx)

	s.state = goservices.StateStopped
	return err
}
