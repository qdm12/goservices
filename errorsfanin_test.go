package goservices

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_errorsFanIn(t *testing.T) {
	t.Parallel()

	e, reader := newErrorsFanIn()

	goodRuntimeErr := make(chan error)
	e.add("good", goodRuntimeErr)

	badRuntimeErr := make(chan error, 1)
	e.add("bad", badRuntimeErr)

	errTest := errors.New("test error")
	badRuntimeErr <- errTest

	err := <-reader

	checkErrIsErrTest(t, err, "bad", errTest)

	e.stop()

	_, ok := <-reader
	assert.False(t, ok)
}

func Test_newErrorsFanIn(t *testing.T) {
	t.Parallel()

	actual, reader := newErrorsFanIn()

	assert.NotNil(t, reader)
	assert.NotNil(t, actual.output)
	actual.output = nil

	expected := &errorsFanIn{}
	assert.Equal(t, expected, actual)
}

func Test_errorsFanIn_add(t *testing.T) {
	t.Parallel()

	const serviceName = "test"

	t.Run("stop fan in", func(t *testing.T) {
		t.Parallel()

		e, reader := newErrorsFanIn()
		runError := make(chan error)

		e.add(serviceName, runError)

		require.Len(t, e.serviceToFaninStop, 1)
		require.Len(t, e.serviceToFaninDone, 1)

		e.stop()
		<-reader
	})

	t.Run("fan in error", func(t *testing.T) {
		t.Parallel()

		e, reader := newErrorsFanIn()
		runError := make(chan error)

		e.add(serviceName, runError)

		require.Len(t, e.serviceToFaninStop, 1)
		require.Len(t, e.serviceToFaninDone, 1)

		errTest := errors.New("test error")
		runError <- errTest

		err := <-reader
		checkErrIsErrTest(t, err, serviceName, errTest)

		e.stop()
	})
}

func Test_errorsFanIn_fanIn(t *testing.T) {
	t.Parallel()

	t.Run("stop fan in", func(t *testing.T) {
		t.Parallel()

		e := &errorsFanIn{
			output: make(chan serviceError),
		}
		const serviceName = "test"
		input := make(chan error)
		stop := make(chan struct{})
		close(stop)
		done := make(chan struct{})
		ready := make(chan struct{})

		e.fanIn(serviceName, input, ready, stop, done)

		_, ok := <-ready
		assert.False(t, ok)
		_, ok = <-done
		assert.False(t, ok)
	})

	t.Run("stop_and_input_race", func(t *testing.T) {
		t.Parallel()

		e := &errorsFanIn{
			output: make(chan serviceError),
		}
		input := make(chan error, 1)
		input <- errors.New("test error")
		stop := make(chan struct{})
		close(stop)
		done := make(chan struct{})
		ready := make(chan struct{})

		e.fanIn("", input, ready, stop, done)

		_, ok := <-ready
		assert.False(t, ok)
		_, ok = <-done
		assert.False(t, ok)

		// Check input is drained
		select {
		case <-input:
			t.Error("input channel is not drained")
		default:
		}
	})

	t.Run("input_closed", func(t *testing.T) {
		t.Parallel()

		e := &errorsFanIn{}
		input := make(chan error)
		close(input)
		done := make(chan struct{})
		ready := make(chan struct{})

		const expectedPanicMessage = "run error service channel closed unexpectedly"
		assert.PanicsWithValue(t, expectedPanicMessage, func() {
			e.fanIn("", input, ready, nil, done)
		})
		<-done
	})

	t.Run("discard_input_errors_after_first", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")

		e := &errorsFanIn{
			output: make(chan serviceError, 1),
		}

		input := make(chan error, 1)
		stop := make(chan struct{})

		service := "A"
		done := make(chan struct{})
		ready := make(chan struct{})
		input <- errTest
		e.fanIn(service, input, ready, stop, done)

		_, ok := <-ready
		assert.False(t, ok)
		err := <-e.output
		checkErrIsErrTest(t, err, service, errTest)
		// Check output is now closed
		_, ok = <-e.output
		assert.False(t, ok)
		_, ok = <-done
		assert.False(t, ok)

		service = "B"
		done = make(chan struct{})
		ready = make(chan struct{})
		input <- errTest
		e.fanIn(service, input, ready, stop, done)

		_, ok = <-ready
		assert.False(t, ok)
		// Check output remains closed
		_, ok = <-e.output
		assert.False(t, ok)
		_, ok = <-done
		assert.False(t, ok)
	})

	t.Run("fan in error", func(t *testing.T) {
		t.Parallel()

		e := &errorsFanIn{
			output: make(chan serviceError, 1),
		}
		const service = "test"
		errTest := errors.New("test error")
		input := make(chan error, 1)
		input <- errTest
		stop := make(chan struct{})
		done := make(chan struct{})
		ready := make(chan struct{})

		e.fanIn(service, input, ready, stop, done)

		_, ok := <-ready
		assert.False(t, ok)

		err := <-e.output
		checkErrIsErrTest(t, err, service, errTest)

		_, ok = <-done
		assert.False(t, ok)
	})
}

func Test_errorsFanIn_stop(t *testing.T) {
	t.Parallel()

	e, reader := newErrorsFanIn()

	const numberOfServices = 2
	for i := range numberOfServices {
		e.add(fmt.Sprint(i), make(chan error))
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, ok := <-reader
		assert.False(t, ok)
	}()

	e.stop()

	_, ok := <-reader
	assert.False(t, ok)
}
