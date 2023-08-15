package helpers

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

type DummyServiceSettings struct {
	// Name is the name of the service returned by the String function.
	Name string
	// MaxStart is the maximum time the service will take to start.
	// Setting it to 0 will start the dummy service instantly.
	MaxStart time.Duration
	// MaxLife is the maximum time the service will run before crashing.
	// Setting it to 0 will make the service run forever.
	MaxLife time.Duration
}

func NewDummyService(settings DummyServiceSettings) *DummyService {
	return &DummyService{
		settings: settings,
	}
}

type DummyService struct {
	settings DummyServiceSettings
}

func (s *DummyService) String() string {
	return s.settings.Name
}

var ErrCrashed = errors.New("crashed")

func (s *DummyService) Start(ctx context.Context) (runError <-chan error, startErr error) {
	if s.settings.MaxLife > 0 {
		readWriteRunError := make(chan error)
		time.AfterFunc(s.settings.MaxLife, func() {
			readWriteRunError <- ErrCrashed
		})
		runError = readWriteRunError
	}

	if s.settings.MaxStart == 0 {
		return runError, nil
	}

	startTime := time.Duration(rand.Intn(int(s.settings.MaxStart))) //nolint:gosec
	timer := time.NewTimer(startTime)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return runError, nil
	}
}

func (s *DummyService) Stop() error {
	return nil
}
