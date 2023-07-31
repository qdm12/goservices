package goservices

import (
	"errors"
	"fmt"
)

var (
	ErrServiceIsNil = errors.New("service is nil")

	ErrNoService = errors.New("no service specified")

	ErrNoServiceStart            = errors.New("no service start order specified")
	ErrNoServiceStop             = errors.New("no service stop order specified")
	ErrServicesStartStopMismatch = errors.New("services to start and stop mismatch")
	ErrServicesNotUnique         = errors.New("services are not unique")

	ErrAlreadyStarted = errors.New("already started")
	ErrAlreadyStopped = errors.New("already stopped")
)

const (
	errorFormatCrash = "%s crashed: %s"
	errorFormatStart = "starting %s: %s"
	errorFormatStop  = "stopping %s: %s"
)

var _ error = serviceError{}

type serviceError struct {
	format      string
	serviceName string
	err         error
}

func (s serviceError) Error() string {
	if s.err == nil {
		panic("cannot have nil error in serviceError")
	}
	return fmt.Sprintf(s.format, s.serviceName, s.err.Error())
}

func (s serviceError) Unwrap() error {
	return s.err
}

func addStopError(collected error, serviceName string,
	newErr error) (newCollected error) {
	if newErr == nil {
		return collected
	}

	newErr = fmt.Errorf("stopping %s: %w", serviceName, newErr)
	if collected == nil {
		return newErr
	}
	return fmt.Errorf("%w; %w", collected, newErr)
}

// addCtxErrorIfNeeded adds the ctxErr to the serviceErr if
// ctxErr is not nil and the serviceErr does not wrap the ctxErr
// already.
// This is done in case one of the services implementation do not
// wrap the context error in its error, if its start context is
// canceled.
func addCtxErrorIfNeeded(serviceErr, ctxErr error) (result error) {
	switch {
	case ctxErr == nil:
		return serviceErr
	case serviceErr == nil:
		return ctxErr
	case errors.Is(serviceErr, ctxErr):
		return serviceErr
	}
	return fmt.Errorf("%w: %w", serviceErr, ctxErr)
}
