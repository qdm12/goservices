package goservices

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_serviceError(t *testing.T) {
	t.Parallel()
	errTest := errors.New("test error")

	testCases := map[string]struct {
		serviceError serviceError
		errString    string
		errUnwrapped error
		panicValue   string
	}{
		"no err set panics": {
			serviceError: serviceError{
				format:      errorFormatCrash,
				serviceName: "A",
			},
			panicValue: "cannot have nil error in serviceError",
		},
		"error set": {
			serviceError: serviceError{
				format:      errorFormatCrash,
				serviceName: "A",
				err:         errTest,
			},
			errString:    "A crashed: test error",
			errUnwrapped: errTest,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.panicValue != "" {
				assert.PanicsWithValue(t, testCase.panicValue, func() {
					_ = testCase.serviceError.Error()
				})
				return
			}

			assert.ErrorIs(t, testCase.serviceError, testCase.errUnwrapped)
			assert.EqualError(t, testCase.serviceError, testCase.errString)
		})
	}
}

func Test_addStopError(t *testing.T) {
	t.Parallel()

	errTest := errors.New("test error")
	errTest2 := errors.New("test error 2")

	testCases := map[string]struct {
		collected           error
		serviceName         string
		newErr              error
		newCollectedErrors  []error
		newCollectedMessage string
	}{
		"all_nils": {},
		"collected_nil_new_error": {
			collected:           fmt.Errorf("stopping A: %w", errTest),
			newCollectedErrors:  []error{errTest},
			newCollectedMessage: "stopping A: test error",
		},
		"nil_collected_new_error": {
			serviceName:         "A",
			newErr:              errTest,
			newCollectedErrors:  []error{errTest},
			newCollectedMessage: "stopping A: test error",
		},
		"collected_new_error": {
			serviceName:         "B",
			collected:           fmt.Errorf("stopping A: %w", errTest),
			newErr:              errTest2,
			newCollectedErrors:  []error{errTest, errTest2},
			newCollectedMessage: "stopping A: test error; stopping B: test error 2",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			newCollected := addStopError(testCase.collected, testCase.serviceName,
				testCase.newErr)

			if len(testCase.newCollectedErrors) == 0 {
				assert.NoError(t, newCollected)
				return
			}

			for _, err := range testCase.newCollectedErrors {
				assert.ErrorIs(t, newCollected, err)
			}
			assert.EqualError(t, newCollected, testCase.newCollectedMessage)
		})
	}
}

func Test_addCtxErrorIfNeeded(t *testing.T) {
	t.Parallel()

	errTest := errors.New("test error")

	testCases := map[string]struct {
		serviceErr    error
		ctxErr        error
		resultErrors  []error
		resultMessage string
	}{
		"all_nils": {},
		"service_error_only": {
			serviceErr:    errTest,
			resultErrors:  []error{errTest},
			resultMessage: "test error",
		},
		"ctx_error_only": {
			ctxErr:        errTest,
			resultErrors:  []error{errTest},
			resultMessage: "test error",
		},
		"service_is_ctx_error": {
			serviceErr:    fmt.Errorf("service crashed: %w", context.Canceled),
			ctxErr:        context.Canceled,
			resultErrors:  []error{context.Canceled},
			resultMessage: "service crashed: context canceled",
		},
		"service_and_ctx_distinct_errors": {
			serviceErr:    errTest,
			ctxErr:        context.Canceled,
			resultErrors:  []error{errTest, context.Canceled},
			resultMessage: "test error: context canceled",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := addCtxErrorIfNeeded(testCase.serviceErr, testCase.ctxErr)

			if len(testCase.resultErrors) == 0 {
				assert.NoError(t, result)
				return
			}
			for _, err := range testCase.resultErrors {
				assert.ErrorIs(t, result, err)
			}
			assert.EqualError(t, result, testCase.resultMessage)
		})
	}
}
